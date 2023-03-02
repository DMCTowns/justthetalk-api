// This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
// Copyright (c) 2021 John Dudmesh.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.

// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package businesslogic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"justthetalk/connections"
	"justthetalk/model"

	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

var (
	indexRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "elasticsearch_index_request_count",
		Help: "Count of post index requests",
	}, []string{"success"})
)

const (
	PubSubMessageTypeTask = "task"

	PubSubMessageActionCreate = "create"
	PubSubMessageActionUpdate = "update"
	PubSubMessageActionDelete = "delete"

	PubSubMessageActionApproved = "approved"
	PubSubMessageActionRejected = "rejected"
)

type PostProcessor struct {
	userCache       *UserCache
	folderCache     *FolderCache
	discussionCache *DiscussionCache
	publishChannel  chan *model.Post
	endWait         sync.WaitGroup
	startWait       sync.WaitGroup
	isRunning       bool
}

type Envelope struct {
	Action string      `json:"action"`
	Urn    string      `json:"urn"`
	Data   interface{} `json:"data"`
}

func NewPostProcessor(userCache *UserCache, folderCache *FolderCache, discussionCache *DiscussionCache) *PostProcessor {

	pubSub := &PostProcessor{
		userCache:       userCache,
		folderCache:     folderCache,
		discussionCache: discussionCache,
		publishChannel:  make(chan *model.Post, 50),
	}

	return pubSub

}

func (p *PostProcessor) PublishPost(post *model.Post) {
	p.publishChannel <- post
}

func (p *PostProcessor) Run() {
	p.startWait.Add(1)
	go p.worker()
	p.startWait.Wait()
}

func (p *PostProcessor) Close() {
	close(p.publishChannel)
	p.endWait.Wait()
}

func (p *PostProcessor) IsRunning() bool {
	return p.isRunning
}

func (p *PostProcessor) QueueLength() int {
	return len(p.publishChannel)
}

func (p *PostProcessor) worker() {

	log.Info("Starting PostProcessor")

	defer func() {
		p.isRunning = false
	}()

	p.isRunning = true
	p.startWait.Done()

	p.endWait.Add(1)
	for post := range p.publishChannel {
		go p.DispatchToSubscribers(post)
		go p.DispatchToElasticsearch(post)
	}
	p.endWait.Done()

	log.Info("Closed PostProcessor")

}

type subscribedUser struct {
	UserId uint `gorm:"column:user_id"`
}

func (p *PostProcessor) DispatchToSubscribers(post *model.Post) (dispatchError error) {

	connections.WithDatabase(5*time.Second, func(db *gorm.DB) {

		defer func() {
			if r := recover(); r != nil {
				dispatchError = r.(error)
				log.Error(dispatchError)
				debug.PrintStack()
			}
		}()

		var frontPageEntry model.FrontPageEntry
		if result := db.Raw("call get_frontpage_entry(?)", post.DiscussionId).First(&frontPageEntry); result.Error != nil {
			panic(result.Error)
		}

		var messageData string
		if data, err := json.Marshal(frontPageEntry); err != nil {
			panic(err)
		} else {
			messageData = string(data)
		}

		log.Debug(messageData)

		if rows, err := db.Model(&subscribedUser{}).Raw("call get_subscribers_for_post(?)", post.Id).Rows(); err == nil {
			defer rows.Close()

			for rows.Next() {

				var subscriber subscribedUser
				if err := db.ScanRows(rows, &subscriber); err != nil {
					panic(err)
				}

				if p.userCache.IsActiveSubscriber(subscriber.UserId) && subscriber.UserId != post.CreatedByUserId {

					ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancelFn()

					topic := fmt.Sprintf("user:%d", subscriber.UserId)
					connections.RedisConnection().Publish(ctx, topic, messageData)

				}

			}

		}

	})

	return nil

}

func (p *PostProcessor) DispatchToElasticsearch(post *model.Post) bool {

	defer func() {
		if r := recover(); r != nil {
			indexRequestCount.WithLabelValues("failure").Inc()
			log.Errorf("ES index failure: %v", r.(error))
		}
	}()

	if post.Status == model.PostStatusOK || post.Status == model.PostStatusWatch {
		p.indexPostIntoSearchEngine(post)
	} else {
		p.deletePostFromSearchEngine(post)
	}

	return true

}

func (p *PostProcessor) deletePostFromSearchEngine(post *model.Post) {

	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()

	req := esapi.DeleteRequest{
		Index:      "posts",
		DocumentID: fmt.Sprintf("%d", post.Id),
	}

	res, err := req.Do(ctx, connections.ElasticSearchConnection())
	if err != nil {
		panic(fmt.Errorf("[%s] error deleting document ID=%d", res.Status(), post.Id))
	}
	defer res.Body.Close()

	if res.IsError() {
		panic(fmt.Errorf("[%s] error deleting document ID=%d", res.Status(), post.Id))
	} else {
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			panic(fmt.Errorf("error parsing the response body: %s", err))
		} else {
			log.Debugf("%v", r)
		}
	}

}

func (p *PostProcessor) indexPostIntoSearchEngine(post *model.Post) {

	var data []byte
	var err error

	discussion := p.discussionCache.UnsafeGet(post.DiscussionId)
	if discussion.Status != model.DiscussionStatusOk || discussion.IsLocked || discussion.IsDeleted {
		return
	}

	folder := p.folderCache.UnsafeGet(discussion.FolderId)
	if folder.Type != model.FolderTypeNormal {
		return
	}

	user := p.userCache.Get(post.CreatedByUserId)

	var doc = model.IndexablePost{
		Id:               post.Id,
		CreatedDate:      post.CreatedDate,
		Text:             post.Text,
		Username:         user.Username,
		FolderName:       folder.Description,
		DiscussionTitle:  discussion.Title,
		DiscussionHeader: discussion.Header,
	}

	if data, err = json.Marshal(doc); err != nil {
		panic(err)
	}

	req := esapi.IndexRequest{
		Index:      "posts",
		DocumentID: fmt.Sprintf("%d", post.Id),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	res, err := req.Do(ctx, connections.ElasticSearchConnection())
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.IsError() {
		panic(fmt.Errorf("[%s] error indexing document ID=%d", res.Status(), post.Id))
	} else {

		indexRequestCount.WithLabelValues("success").Inc()

		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			panic(fmt.Errorf("error parsing the response body: %s", err))
		} else {
			log.Debugf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}

	}

}

func (p *PostProcessor) IndexAllPosts() {

	connections.WithDatabase(1*time.Hour, func(db *gorm.DB) {

		rows, err := db.Raw("call get_indexable_posts();").Rows()
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		var batchCount int
		var batchNum int
		es := connections.ElasticSearchConnection()
		for rows.Next() {

			var post model.IndexablePost
			if err := db.ScanRows(rows, &post); err != nil {
				panic(err)
			}

			meta := []byte(fmt.Sprintf("{ \"index\" : { \"_id\" : \"%d\" } } \n", post.Id))

			data, err := json.Marshal(post)
			if err != nil {
				panic(err)
			}
			data = append(data, "\n"...)

			buf.Grow(len(meta) + len(data))
			buf.Write(meta)
			buf.Write(data)

			batchCount++
			if batchCount%1000 == 0 {

				res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex("posts"))
				if err != nil {
					log.Fatalf("Failure indexing batch %d: %s", batchNum, err)
				}

				var responseData map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&responseData); err != nil {
					log.Errorf("Failure to to parse response body: %s - %s", err, res.Body)
				}

				if res.IsError() {
					log.Errorf("%v", responseData)
				}

				batchNum++
				buf.Reset()

				log.Infof("Completed batch: %d, total rows=%d", batchNum, batchCount)

			}

		}

	})

}

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
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"time"

	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func createSearchHistory(queryString string, user *model.User, ipAddress string, db *gorm.DB) {

	history := model.SearchHistory{
		CreatedDate: time.Now(),
		UserId:      user.Id,
		IPAddress:   ipAddress,
		Query:       queryString,
	}

	if result := db.Table("search_history").Create(&history); result.Error != nil {
		log.Errorf("%v", result.Error)
		panic(utils.ErrInternalError)
	}

}

func SearchPosts(queryString string, size int, page int, user *model.User, ipAddress string, folderCache *FolderCache, discussionCache *DiscussionCache, db *gorm.DB, ctx context.Context) []*model.SearchResult {

	createSearchHistory(queryString, user, ipAddress, db)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": queryString,
			},
		},
		"size":    size,
		"from":    page,
		"fields":  []string{},
		"_source": false,
		"sort": []map[string]interface{}{
			{"date": "desc"},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	elastic := connections.ElasticSearchConnection()
	res, err := elastic.Search(
		elastic.Search.WithContext(ctx),
		elastic.Search.WithIndex("posts"),
		elastic.Search.WithBody(&buf),
		elastic.Search.WithTrackTotalHits(true),
		elastic.Search.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	var e map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		log.Errorf("Error parsing the response body: %s", err)
		panic(utils.ErrInternalError)
	}

	if res.IsError() {
		log.Errorf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
		panic(utils.ErrBadRequest)
	}

	hits := e["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})
	hitsList := hits["hits"].([]interface{})

	results := make([]*model.SearchResult, 0)
	for _, hit := range hitsList {

		d := hit.(map[string]interface{})
		docId := d["_id"].(string)

		var post model.Post
		if result := db.Raw("call get_post(?)", docId).First(&post); result.Error != nil {
			log.Errorf("%v", result.Error)
			panic(utils.ErrInternalError)
		}

		if post.Status == 0 {
			discussion := discussionCache.Get(post.DiscussionId, user)
			folder := folderCache.Get(discussion.FolderId, user)
			slugText := slug.Make(discussion.Title)
			post.Url = fmt.Sprintf("/%s/%d/%s/%d", folder.Key, discussion.Id, slugText, post.PostNum)
			post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
			result := &model.SearchResult{
				Post:         &post,
				Folder:       folder,
				Discussion:   discussion,
				TotalResults: int(total["value"].(float64)),
			}
			results = append(results, result)
		}

	}

	return results

}

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
	"context"
	"justthetalk/connections"
	"justthetalk/model"
	"sync"
	"time"

	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BookmarkProcessor struct {
	userCache     *UserCache
	updateChannel chan *model.UserDiscussionBookmark
	waiter        sync.WaitGroup
	ctx           context.Context
}

func NewBookmarkProcessor() *BookmarkProcessor {

	cache := &BookmarkProcessor{
		updateChannel: make(chan *model.UserDiscussionBookmark, 50),
		ctx:           context.Background(),
	}

	go cache.worker()

	return cache

}

func (cache *BookmarkProcessor) Enqueue(bookmark *model.UserDiscussionBookmark) {
	cache.updateChannel <- bookmark
}

func (cache *BookmarkProcessor) Close() {
	close(cache.updateChannel)
	cache.waiter.Wait()
}

func (cache *BookmarkProcessor) worker() {
	cache.waiter.Add(1)
	for bookmark := range cache.updateChannel {
		cache.processBookmark(bookmark)
	}
	cache.waiter.Done()
}

func (cache *BookmarkProcessor) processBookmark(bookmark *model.UserDiscussionBookmark) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Error(err)
			debug.PrintStack()
		}
	}()

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Exec("call update_user_bookmark(?, ?, ?, ?, ?)", bookmark.UserId, bookmark.DiscussionId, bookmark.LastPostId, bookmark.LastPostCount, bookmark.LastPostDate); result.Error != nil {
			panic(result.Error)
		}
	})

	cache.userCache.PutBookmark(bookmark)

}

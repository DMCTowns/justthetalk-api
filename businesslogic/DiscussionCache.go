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
	"errors"
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"

	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

type DiscussionCache struct {
	postFormatter *utils.PostFormatter
	folderCache   *FolderCache
}

func NewDiscussionCache(folderCache *FolderCache) *DiscussionCache {

	return &DiscussionCache{
		postFormatter: utils.NewPostFormatter(),
		folderCache:   folderCache,
	}

}

func (cache *DiscussionCache) Get(discussionId uint, user *model.User) *model.Discussion {

	discussion := cache.UnsafeGet(discussionId)

	if discussion.Status != model.DiscussionStatusOk || discussion.IsDeleted {
		if user == nil || !user.IsAdmin {
			panic(utils.ErrForbidden)
		}
	}

	cache.folderCache.Get(discussion.FolderId, user)

	discussion.IsBlocked = cache.IsBlocked(discussion, user)

	return discussion

}

func (cache *DiscussionCache) UnsafeGet(discussionId uint) *model.Discussion {

	var discussion model.Discussion

	key := "D" + strconv.Itoa(int(discussionId))
	//val, err := connections.RedisConnection().Get(context.Background(), key).Result()
	val := []byte{}
	err := redis.Nil
	// TODO - don't use Redis cache while running in parallel with legacy site
	if err == redis.Nil {
		log.Debug("DiscussionCache: cache miss")
		connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
			if result := db.Raw("call get_discussion(?)", discussionId).First(&discussion); result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					panic(utils.ErrNotFound)
				} else {
					utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
				}
			}
		})

		folder := cache.folderCache.UnsafeGet(discussion.FolderId)
		discussion.Url = utils.UrlForDiscussion(folder, &discussion)
		discussion.HeaderMarkup = cache.postFormatter.ApplyPostFormatting(discussion.Header, &discussion)

		cache.Put(&discussion)

	} else {
		log.Debug("DiscussionCache: cache hit")
		if err := json.Unmarshal([]byte(val), &discussion); err != nil {
			panic(err)
		}
		connections.RedisConnection().Expire(context.Background(), key, time.Hour*1)
	}

	return &discussion

}

func (cache *DiscussionCache) Put(discussion *model.Discussion) {

	key := "D" + strconv.Itoa(int(discussion.Id))
	data, err := json.Marshal(discussion)
	if err != nil {
		panic(err)
	}

	status := connections.RedisConnection().Set(context.Background(), key, data, time.Hour*1)
	if status.Err() != nil {
		panic(status.Err().Error())
	}

}

func (cache *DiscussionCache) Flush(discussionId uint) {

	discussionIdStr := strconv.Itoa(int(discussionId))

	status := connections.RedisConnection().Del(context.Background(), "D"+discussionIdStr, "B"+discussionIdStr)
	if status.Err() != nil {
		panic(status.Err().Error())
	}

}

func (cache *DiscussionCache) BlockedUsers(discussion *model.Discussion) map[uint]*model.BlockedDiscussionUser {

	var blockedUserMap map[uint]*model.BlockedDiscussionUser

	key := "B" + strconv.Itoa(int(discussion.Id))
	val, err := connections.RedisConnection().Get(context.Background(), key).Result()
	if err == redis.Nil {

		connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
			blockedUserMap = FetchBlockedUsers(discussion, db)
		})

		data, err := json.Marshal(&blockedUserMap)
		if err != nil {
			panic(err)
		}

		status := connections.RedisConnection().Set(context.Background(), key, data, time.Hour*1)
		if status.Err() != nil {
			panic(status.Err().Error())
		}

	} else {
		if err := json.Unmarshal([]byte(val), &blockedUserMap); err != nil {
			panic(err)
		}
		connections.RedisConnection().Expire(context.Background(), key, time.Hour*1)
	}

	return blockedUserMap

}

func (cache *DiscussionCache) BlockOrUnblockUser(discussion *model.Discussion, targetUser *model.User, blockNotUnblock bool, adminUser *model.User) map[uint]*model.BlockedDiscussionUser {

	var blockedUserMap map[uint]*model.BlockedDiscussionUser

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		blockedUserMap = BlockUnblockUser(discussion, targetUser, blockNotUnblock, adminUser, db)
	})

	key := "B" + strconv.Itoa(int(discussion.Id))
	data, err := json.Marshal(&blockedUserMap)
	if err != nil {
		panic(err)
	}

	status := connections.RedisConnection().Set(context.Background(), key, data, time.Hour*1)
	if status.Err() != nil {
		panic(status.Err().Error())
	}

	return blockedUserMap

}

func (cache *DiscussionCache) IsBlocked(discussion *model.Discussion, user *model.User) bool {

	blockedUserMap := cache.BlockedUsers(discussion)

	isBlocked := false
	if user != nil {
		if block, exists := blockedUserMap[user.Id]; exists {
			isBlocked = block.Status
		}
	}

	return isBlocked

}

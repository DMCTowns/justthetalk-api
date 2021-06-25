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
	"encoding/json"
	"fmt"
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserCache struct {
	subscribers     map[string]*model.User
	subscribersLock sync.RWMutex
}

func NewUserCache() *UserCache {
	cache := &UserCache{
		subscribers:     make(map[string]*model.User, 0),
		subscribersLock: sync.RWMutex{},
	}
	return cache
}

func (cache *UserCache) Get(userId uint) *model.User {

	var user model.User

	userKey := fmt.Sprintf("U%d", userId)
	val, err := connections.RedisConnection().Get(context.Background(), userKey).Result()
	if err != redis.Nil {

		if err := json.Unmarshal([]byte(val), &user); err != nil {
			panic(utils.ErrInternalError)
		}
		connections.RedisConnection().Expire(context.Background(), userKey, time.Hour*24)

	} else {
		cache.getFromDB(userId, &user)
	}

	return &user

}

func (cache *UserCache) Reload(userId uint) *model.User {
	var user model.User
	cache.getFromDB(userId, &user)
	return &user
}

func (cache *UserCache) getFromDB(userId uint, user *model.User) {

	var ignored []*model.IgnoredUser

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {

		if result := db.Raw("call get_user(?)", userId).First(&user); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}

		if result := db.Raw("call get_user_ignored_users(?)", userId).Scan(&ignored); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}

	})

	user.IgnoredUsers = make(map[uint]*model.IgnoredUser)
	for _, item := range ignored {
		user.IgnoredUsers[item.IgnoredUserId] = item
	}

	cache.Put(user)

}

func (cache *UserCache) Put(user *model.User) {

	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}

	userKey := fmt.Sprintf("U%d", user.Id)
	status := connections.RedisConnection().Set(context.Background(), userKey, string(data), time.Hour*24)
	if status.Err() != nil {
		panic(status.Err())
	}

}

func (cache *UserCache) Flush(user *model.User) {
	userKey := fmt.Sprintf("U%d", user.Id)
	connections.RedisConnection().Del(context.Background(), userKey)
	dataKey := fmt.Sprintf("US%d", user.Id)
	connections.RedisConnection().Del(context.Background(), dataKey)
}

func (cache *UserCache) ClearRefreshToken(refreshToken string) {
	tokenKey := "T" + refreshToken
	status := connections.RedisConnection().Del(context.Background(), tokenKey)
	if status.Err() != nil {
		panic(status.Err())
	}
}

func (cache *UserCache) RotateRefreshToken(user *model.User) string {

	refreshToken := uuid.NewString()
	tokenKey := "T" + refreshToken
	status := connections.RedisConnection().Set(context.Background(), tokenKey, strconv.Itoa(int(user.Id)), time.Hour*720)
	if status.Err() != nil {
		panic(status.Err())
	}

	return refreshToken

}

func (cache *UserCache) GetUserIdForRefreshToken(refreshToken string) uint {
	tokenKey := "T" + refreshToken
	if result := connections.RedisConnection().Get(context.Background(), tokenKey); result.Err() != nil {
		panic(utils.ErrForbidden)
	} else {
		if val, err := result.Int64(); err == nil {
			return uint(val)
		} else {
			panic(err)
		}
	}
}

func (cache *UserCache) PutSidebandData(userData *model.UserSidebandData) {

	data, err := json.Marshal(userData)
	if err != nil {
		panic(err)
	}

	userKey := fmt.Sprintf("US%d", userData.UserId)
	status := connections.RedisConnection().Set(context.Background(), userKey, string(data), time.Hour*24)
	if status.Err() != nil {
		panic(status.Err())
	}

}

func (cache *UserCache) GetSidebandData(userId uint) *model.UserSidebandData {

	var userData model.UserSidebandData

	userKey := fmt.Sprintf("US%d", userId)
	val, err := connections.RedisConnection().Get(context.Background(), userKey).Result()
	if err != redis.Nil {
		if err := json.Unmarshal([]byte(val), &userData); err != nil {
			panic(utils.ErrInternalError)
		}
		connections.RedisConnection().Expire(context.Background(), userKey, time.Hour*24)
	} else {

		user := cache.Get(userId)

		userData.UserId = user.Id

		cache.populateDiscussionBookmarks(&userData)
		cache.populateFolderSubscriptions(&userData)
		cache.populateFolderSubscriptionExceptions(&userData)
		cache.populateDiscussionSubscriptions(&userData)

		cache.PutSidebandData(&userData)

	}

	return &userData

}

func (cache *UserCache) FlushSidebandData(userId uint) {

	userKey := fmt.Sprintf("US%d", userId)
	status := connections.RedisConnection().Del(context.Background(), userKey)
	if status.Err() != nil {
		panic(status.Err())
	}

}

func (cache *UserCache) populateDiscussionBookmarks(userData *model.UserSidebandData) {

	userData.DiscussionBookmarks = make(map[uint]*model.UserDiscussionBookmark)

	var discussionBookmarks []*model.UserDiscussionBookmark
	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Raw("call get_user_discussion_bookmarks(?)", userData.UserId).Scan(&discussionBookmarks); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	})

	for _, item := range discussionBookmarks {
		userData.DiscussionBookmarks[item.DiscussionId] = item
	}

}

func (cache *UserCache) populateFolderSubscriptions(userData *model.UserSidebandData) {

	userData.FolderSubscriptions = make(map[uint]*model.UserFolderSubscription)

	var folderSubscriptions []*model.UserFolderSubscription
	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Raw("call get_user_folder_subscriptions(?)", userData.UserId).Scan(&folderSubscriptions); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	})

	for _, item := range folderSubscriptions {
		userData.FolderSubscriptions[item.FolderId] = item
	}

}

func (cache *UserCache) populateFolderSubscriptionExceptions(userData *model.UserSidebandData) {

	userData.FolderSubscriptionExceptions = make(map[uint]*model.UserFolderSubscriptionException)

	var folderSubscriptionExceptions []*model.UserFolderSubscriptionException
	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Raw("call get_user_folder_subscription_exceptions(?)", userData.UserId).Scan(&folderSubscriptionExceptions); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	})

	for _, item := range folderSubscriptionExceptions {
		userData.FolderSubscriptionExceptions[item.DiscussionId] = item
	}

}

func (cache *UserCache) populateDiscussionSubscriptions(userData *model.UserSidebandData) {

	var discussionSubscriptions []*model.UserDiscussionSubscription
	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Raw("call get_user_discussion_subscriptions(?)", userData.UserId).Scan(&discussionSubscriptions); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	})

	userData.DiscussionSubscriptions = make(map[uint]*model.UserDiscussionSubscription)
	for _, item := range discussionSubscriptions {
		userData.DiscussionSubscriptions[item.DiscussionId] = item
	}

}

func (cache *UserCache) SubscribeToDiscussion(discussion *model.Discussion, user *model.User) *model.UserSidebandData {

	sidebandData := cache.GetSidebandData(user.Id)

	if _, exists := sidebandData.DiscussionBookmarks[discussion.Id]; !exists {

		connections.WithDatabase(1*time.Second, func(db *gorm.DB) {

			var discussionSubscriptions []*model.UserDiscussionSubscription
			if result := db.Raw("call update_user_discussion_subscription(?, ?, ?)", user.Id, discussion.Id, 1).Scan(&discussionSubscriptions); result.Error != nil {
				utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
			}

			sidebandData.DiscussionSubscriptions = make(map[uint]*model.UserDiscussionSubscription)
			for _, item := range discussionSubscriptions {
				sidebandData.DiscussionSubscriptions[item.DiscussionId] = item
			}
			cache.PutSidebandData(sidebandData)

		})
	}

	return sidebandData

}

func (cache *UserCache) GetDiscussionSubscriptionStatus(discussion *model.Discussion, user *model.User) bool {

	isSubscribed := false

	if user != nil {
		userData := cache.GetSidebandData(user.Id)
		if _, exists := userData.DiscussionSubscriptions[discussion.Id]; exists {
			isSubscribed = true
		} else if _, exists := userData.FolderSubscriptions[discussion.FolderId]; exists {
			if _, exists := userData.FolderSubscriptionExceptions[discussion.Id]; !exists {
				isSubscribed = true
			}
		}
	}

	return isSubscribed

}

func (cache *UserCache) GetFolderSubscriptionStatus(folder *model.Folder, user *model.User) bool {

	isSubscribed := false

	if user != nil {
		userData := cache.GetSidebandData(user.Id)
		if _, exists := userData.FolderSubscriptions[folder.Id]; exists {
			isSubscribed = true
		}
	}

	return isSubscribed

}

func (cache *UserCache) AddSubscriber(user *model.User) *redis.PubSub {

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	topic := fmt.Sprintf("user:%d", user.Id)
	subscription := connections.RedisConnection().Subscribe(ctx, topic)

	cache.subscribersLock.Lock()
	defer cache.subscribersLock.Unlock()

	cache.subscribers[topic] = user

	return subscription

}

func (cache *UserCache) RemoveSubscriber(user *model.User) {

	topic := fmt.Sprintf("user:%d", user.Id)

	cache.subscribersLock.Lock()
	defer cache.subscribersLock.Unlock()

	delete(cache.subscribers, topic)

}

func (cache *UserCache) IsActiveSubscriber(topic string) bool {

	cache.subscribersLock.RLock()
	defer cache.subscribersLock.RUnlock()

	_, exists := cache.subscribers[topic]
	return exists

}

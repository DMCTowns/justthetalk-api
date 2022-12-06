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
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserCache struct {
	subscribers     map[uint]bool
	subscribersLock sync.RWMutex
}

func NewUserCache() *UserCache {
	cache := &UserCache{
		subscribers:     make(map[uint]bool),
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
	cache.FlushById(user.Id)
}

func (cache *UserCache) FlushById(userId uint) {
	userKey := fmt.Sprintf("U%d", userId)
	connections.RedisConnection().Del(context.Background(), userKey)
	dataKey := fmt.Sprintf("US%d", userId)
	connections.RedisConnection().Del(context.Background(), dataKey)
}

func (cache *UserCache) ClearRefreshToken(refreshToken string) {
	log.Debugf("Clear refresh token '%s'", refreshToken)
	tokenKey := "T" + refreshToken
	status := connections.RedisConnection().Del(context.Background(), tokenKey)
	if status.Err() != nil {
		panic(status.Err())
	}
}

func (cache *UserCache) RotateRefreshToken(user *model.User) string {
	log.Debugf("Rotate refresh token for %d", user.Id)
	refreshToken := uuid.NewString()
	tokenKey := "T" + refreshToken
	status := connections.RedisConnection().Set(context.Background(), tokenKey, strconv.Itoa(int(user.Id)), time.Hour*720)
	if status.Err() != nil {
		panic(status.Err())
	}

	return refreshToken

}

func (cache *UserCache) GetUserIdForRefreshToken(refreshToken string) uint {
	log.Debugf("Get refresh token '%s'", refreshToken)
	tokenKey := "T" + refreshToken
	if result := connections.RedisConnection().Get(context.Background(), tokenKey); result.Err() != nil {
		log.Errorf("fetching cached refresh token: %v", result.Err())
		panic(utils.ErrForbidden)
	} else {
		if val, err := result.Int64(); err == nil {
			return uint(val)
		} else {
			panic(err)
		}
	}
}

func (cache *UserCache) AddSubscriber(user *model.User) {

	cache.subscribersLock.Lock()
	defer cache.subscribersLock.Unlock()

	cache.subscribers[user.Id] = true

}

func (cache *UserCache) RemoveSubscriber(user *model.User) {

	cache.subscribersLock.Lock()
	defer cache.subscribersLock.Unlock()

	delete(cache.subscribers, user.Id)

}

func (cache *UserCache) IsActiveSubscriber(userId uint) bool {

	cache.subscribersLock.RLock()
	defer cache.subscribersLock.RUnlock()

	_, exists := cache.subscribers[userId]
	return exists

}

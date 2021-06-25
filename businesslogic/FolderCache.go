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
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"sync"
	"time"

	"gorm.io/gorm"
)

type FolderCache struct {
	entries       []*model.Folder
	byId          map[uint]*model.Folder
	updateChannel chan *model.Post
	mutex         sync.Mutex
}

func NewFolderCache() *FolderCache {

	cache := &FolderCache{
		byId:          make(map[uint]*model.Folder),
		updateChannel: make(chan *model.Post, 50),
	}

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Raw("call get_folders()").Scan(&cache.entries); result.Error != nil {
			panic(result.Error)
		}
	})

	for _, entry := range cache.entries {
		cache.byId[entry.ModelBase.Id] = entry
	}

	return cache

}

func (cache *FolderCache) Entries() []*model.Folder {
	return cache.entries
}

func (cache *FolderCache) Get(id uint, user *model.User) *model.Folder {

	var folder *model.Folder

	if f, exists := cache.byId[id]; exists {
		if f.Type == model.FolderTypeNormal {
			folder = f
		} else if user != nil && user.IsAdmin {
			folder = f
		}
	}

	if folder == nil {
		panic(utils.ErrForbidden)
	}

	return folder

}

func (cache *FolderCache) SafeGet(id uint) *model.Folder {

	if f, exists := cache.byId[id]; exists && f.Type == model.FolderTypeNormal {
		return f
	}

	return nil

}

func (cache *FolderCache) UnsafeGet(id uint) *model.Folder {

	if f, exists := cache.byId[id]; exists {
		return f
	}

	return nil

}

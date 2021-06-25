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
	"errors"
	"justthetalk/connections"
	"justthetalk/utils"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestSearch(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	userId := uint(5540)
	user := userCache.Get(userId)

	query := "johnnythesailor"
	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		var count1 int64
		var count2 int64

		db.Table("search_history").Count(&count1)

		posts := SearchPosts(query, 20, 0, user, "8.8.8.8", folderCache, discussionCache, db, context.Background())

		if len(posts) == 0 {
			t.Error("No results")
		}

		db.Table("search_history").Count(&count2)
		if count2-count1 != 1 {
			t.Error("No search history")
		}

	})

}

func TestSearchFailure(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrBadRequest):
				t.Log("Got bad request")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	userId := uint(5540)
	user := userCache.Get(userId)

	query := ":::::"
	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		SearchPosts(query, 20, 0, user, "8.8.8.8", folderCache, discussionCache, db, context.Background())

		t.Error("Unexpected success")

	})

}

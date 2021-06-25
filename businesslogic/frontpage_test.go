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

	"testing"
	"time"

	"gorm.io/gorm"
)

func TestGetFrontPageLatestAndStartedByMe(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	adminUser := userCache.Get(50)
	normalUser := userCache.Get(5540)
	folder := folderCache.Get(33, adminUser)
	discussion := discussionCache.Get(13506, adminUser)

	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {

		postSpec := model.Post{
			Text: "This is an admin post",
		}

		post := CreatePost(folder, discussion, adminUser, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}

		posts := GetFrontPage(adminUser, "latest", 20, 0, userCache, discussionCache, db)
		if len(posts) != 20 {
			t.Error("Not enough posts for admin user")
		}

		if !(posts[0].DiscussionId == discussion.Id && posts[0].LastPostId == post.Id) {
			t.Error("New post not first in latest list")
		}

		posts = GetFrontPage(adminUser, "startedbyme", 20, 0, userCache, discussionCache, db)
		if !(posts[0].DiscussionId == discussion.Id && posts[0].LastPostId == post.Id) {
			t.Error("New post not first in startedbyme list")
		}

		posts = GetFrontPage(normalUser, "latest", 20, 0, userCache, discussionCache, db)
		if len(posts) != 20 {
			t.Error("Not enough posts for normal user")
		}

		if posts[0].DiscussionId == discussion.Id && posts[0].LastPostId == post.Id {
			t.Error("New post should not be in list for ordinary user")
		}

		posts = GetFrontPage(normalUser, "startedbyme", 20, 0, userCache, discussionCache, db)
		if posts[0].DiscussionId == discussion.Id && posts[0].LastPostId == post.Id {
			t.Error("New post should not be in started by me list for ordinary user")
		}
	})

}

func TestGetFrontPageMostActiveSmokeTest(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	normalUser := userCache.Get(5540)

	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {
		posts := GetFrontPage(normalUser, "mostactive", 20, 0, userCache, discussionCache, db)
		if len(posts) != 20 {
			t.Errorf("Not enough posts for normal user - got %d", len(posts))
		}
		for _, p := range posts {
			folderCache.Get(p.FolderId, normalUser) // will panic if no access
		}
	})

}

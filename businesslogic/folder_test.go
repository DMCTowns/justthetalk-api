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
	"fmt"
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"strings"
	"testing"
	"time"

	"errors"

	"gorm.io/gorm"
)

func TestResubscribingToAFolderSubscriptionExceptionRemovesTheException(t *testing.T) {
	t.Fail()
}

func TestLoggingInWritesToLoginHistory(t *testing.T) {
	t.Fail()
}

func TestPostedOnDiscussionCannotBeDeleted(t *testing.T) {
	t.Fail()
}

func TestPostedOnDiscussionCannotBeEdited(t *testing.T) {
	t.Fail()
}

func TestNormalUserCannotGetAdminFolders(t *testing.T) {
	t.Fail()
}

func TestNormalUserCannotGetAdminFolderPosts(t *testing.T) {
	t.Fail()
}

func TestNormalUserCannotPostAdminFolders(t *testing.T) {
	t.Fail()
}

func TestSubscribeToDiscussion(t *testing.T) {
	t.Fail()
}

func TestSubscribeToFolder(t *testing.T) {
	t.Fail()
}

func TestEditPostsAddsRowToPostEdits(t *testing.T) {
	t.Fail()
}

func TestGetDiscussion(t *testing.T) {
	t.Fail()
}

func TestGetDiscussions(t *testing.T) {
	t.Fail()
}

func TestGetFolders(t *testing.T) {

	folderCache := NewFolderCache()
	folders := folderCache.Entries()
	if len(folders) == 0 {
		t.Fail()
	}

}

func TestCreateDiscussion(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(26, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {

		discussionSpec := model.Discussion{
			Title:  fmt.Sprintf("New & <script>alert('hello')</script>discussion: %s", time.Now().Format("02/01/2006 15:04:05")),
			Header: "This is an test discussion & <script>alert('hello')</script>",
		}

		discussion := CreateDiscussion(folder, &discussionSpec, user, userCache, discussionCache, db)
		if discussion.Id == 0 {
			t.Error("Failed to create discussion")
		}

		if !strings.Contains(discussion.Title, "&amp;") {
			t.Error("Failed to escape ampersand in title")
		}

		if !strings.Contains(discussion.Title, "&lt;script&gt;") {
			t.Error("Failed to escape ampersand in title")
		}

		if !strings.Contains(discussion.Header, "&amp;") {
			t.Error("Failed to escape ampersand in header")
		}

		if !strings.Contains(discussion.Header, "&lt;script&gt;") {
			t.Error("Failed to escape ampersand in header")
		}

	})

}

func TestEditDiscussion(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(26, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {

		discussionSpec := model.Discussion{
			Title:  fmt.Sprintf("New & <script>alert('hello')</script>discussion: %s", time.Now().Format("02/01/2006 15:04:05")),
			Header: "This is an test discussion & <script>alert('hello')</script>",
		}

		discussion := CreateDiscussion(folder, &discussionSpec, user, userCache, discussionCache, db)
		if discussion.Id == 0 {
			t.Error("Failed to create discussion")
		}

		discussion.Title += ":Edited"
		discussion.Header += ":Edited"
		updated := EditDiscussion(folder, discussion, user, discussionCache, db)

		if !strings.HasSuffix(updated.Title, ":Edited") {
			t.Error("Edit Failed")
		}
		if !strings.HasSuffix(updated.Header, ":Edited") {
			t.Error("Edit Failed")
		}

	})

}

func TestDeleteDiscussion(t *testing.T) {
	t.Fail()
}

func TestCreatePostAdminByAdminUser(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(50)
	folder := folderCache.Get(33, user)
	discussion := discussionCache.Get(13506, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		postSpec := model.Post{
			Text: "This is an admin post",
		}
		post := CreatePost(folder, discussion, user, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}
	})

}

func TestCreatePostByNonAdminUser(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(26, user)
	discussion := discussionCache.Get(130, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		postSpec := model.Post{
			Text: "This is an test post <script>alert('hello')</script>",
		}
		post := CreatePost(folder, discussion, user, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}
		if !strings.Contains(post.Text, "&lt;script&gt;") {
			t.Error("Failed to escape ampersand")
		}

	})

}

func TestEditPost(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(26, user)
	discussion := discussionCache.Get(130, user)

	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {

		postSpec := model.Post{
			Text: "This is an test post",
		}
		post := CreatePost(folder, discussion, user, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}

		time.Sleep(2 * time.Second) // delay to  make sure last update date is different from created date

		updateSpec := model.Post{
			ModelBase: model.ModelBase{Id: post.Id},
			Text:      "This is an edited test post",
		}
		updated := EditPost(folder, discussion, user, &updateSpec, db)

		if updated.Id != post.Id {
			t.Error("Wrong post returned")
		}

		if updated.Text != "This is an edited test post" || !updated.LastEditDate.After(post.CreatedDate) {
			t.Error("Update failed")
		}

		var count int64
		db.Table("post_edit").Where("post_id = ?", post.Id).Count(&count)
		if count == 0 {
			t.Error("No row in edit table")
		}

	})

}

func TestDeletePost(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(26, user)
	discussion := discussionCache.Get(130, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		postSpec := model.Post{
			Text: "This is an test post",
		}
		post := CreatePost(folder, discussion, user, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}
		t.Logf("PostId: %d", post.Id)

		updated := DeletePost(folder, discussion, user, post.Id, db)
		if !updated.Deleted {
			t.Error("Post not flagged as deleted")
		}

	})

}

func TestDeletePostFailsForOtherPeople(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrForbidden):
				t.Log("Got forbidden")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(1, user)
	discussion := discussionCache.Get(1, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		post := GetPost(1, db)
		DeletePost(folder, discussion, user, post.Id, db)
	})

	t.Error("Should have panicked")

}

func TestLockedAccountCannotPost(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrForbidden):
				t.Log("Got forbidden")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(730)
	folder := folderCache.Get(26, user)
	discussion := discussionCache.Get(130, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		postSpec := model.Post{
			Text: "This is an test post",
		}
		post := CreatePost(folder, discussion, user, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}
		t.Logf("PostId: %d", post.Id)
	})

	t.Error("Should have panicked")

}

func TestLockedDiscussionCannotPost(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrForbidden):
				t.Log("Got forbidden")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(730)
	folder := folderCache.Get(16, user)
	discussion := discussionCache.Get(479, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		postSpec := model.Post{
			Text: "This is an test post",
		}
		post := CreatePost(folder, discussion, user, &postSpec, discussionCache, userCache, db)
		if post.Id == 0 {
			t.Error("Failed to create post")
		}
		t.Logf("PostId: %d", post.Id)
	})

	t.Error("Should have panicked")

}

func TestGetPosts(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	folder := folderCache.Get(26, user)
	discussion := discussionCache.Get(130, user)

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {
		posts := GetPosts(folder, discussion, user, 1, 20, db)
		if len(posts) != 20 {
			t.Errorf("Not enough posts, got: %d", len(posts))
		}
	})

}

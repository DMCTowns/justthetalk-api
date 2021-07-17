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

//docker run -d -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.12.1

import (
	"justthetalk/connections"
	"justthetalk/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreationAndTeardown(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	p := NewPostProcessor(userCache, folderCache, discussionCache)
	p.Run()

	if !p.IsRunning() {
		t.Error("Failed to start")
	}

	p.Close()
	if p.IsRunning() {
		t.Error("Failed to stop")
	}

}

func TestPublishPostToSearchIndex(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	p := NewPostProcessor(userCache, folderCache, discussionCache)
	p.Run()
	if !p.IsRunning() {
		t.Error("Failed to start")
	}

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {

		var post model.Post
		// normal
		db.Raw("call get_post(?)", 446).First(&post)
		success := p.DispatchToElasticsearch(&post)
		assert.True(t, success)

		// deleted
		db.Raw("call get_post(?)", 90).First(&post)
		success = p.DispatchToElasticsearch(&post)
		assert.True(t, success)

	})

	p.Close()
}

func TestDeletePostFromSearchIndex(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	p := NewPostProcessor(userCache, folderCache, discussionCache)
	p.Run()
	if !p.IsRunning() {
		t.Error("Failed to start")
	}

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {

		var post model.Post
		db.Raw("call get_post(?)", 446).First(&post)
		post.Status = model.PostStatusDeletedByAdmin
		p.DispatchToElasticsearch(&post)

	})

	p.Close()

}

func TestIndexAllPosts(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	p := NewPostProcessor(userCache, folderCache, discussionCache)
	p.IndexAllPosts()

}

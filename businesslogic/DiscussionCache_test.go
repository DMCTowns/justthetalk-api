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
	"justthetalk/utils"
	"testing"
)

func TestGetDiscussionGet(t *testing.T) {

	userCache := NewUserCache()

	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	userId := uint(5540)
	user := userCache.Get(userId)

	discussionId := uint(2271)

	if discussion := discussionCache.Get(discussionId, user); discussion.Id != discussionId {
		t.Error("Failed to get")
	}

	if discussion := discussionCache.Get(discussionId, user); discussion.Id != discussionId {
		t.Error("Failed to get cached version")
	}

}

func TestGetDiscussionGetLockedNonAdmin(t *testing.T) {

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

	userId := uint(5540)
	user := userCache.Get(userId)

	discussionId := uint(21)

	discussionCache.Get(discussionId, user)

	t.Error("Should have panicked")

}

func TestGetDiscussionGetLockedAsAdmin(t *testing.T) {

	userCache := NewUserCache()

	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	userId := uint(50)
	user := userCache.Get(userId)

	discussionId := uint(21)

	if discussion := discussionCache.Get(discussionId, user); discussion.Id != discussionId {
		t.Error("Failed to get")
	}

}

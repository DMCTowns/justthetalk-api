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
	"testing"
)

func TestFetchBlockedUsers(t *testing.T) {

	discussionId := uint(25876)

	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	discussion := discussionCache.UnsafeGet(discussionId)
	blockedUsers := discussionCache.BlockedUsers(discussion)
	if _, exists := blockedUsers[uint(2994)]; !exists {
		t.Fail()
	}

}

func TestBlockUser(t *testing.T) {

	discussionId := uint(25876)

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	discussion := discussionCache.UnsafeGet(discussionId)
	blockedUsers := discussionCache.BlockUser(discussion, user)
	if _, exists := blockedUsers[uint(5540)]; !exists {
		t.Fail()
	}

}

func TestUnblockUser(t *testing.T) {

	discussionId := uint(25876)

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(5540)
	discussion := discussionCache.UnsafeGet(discussionId)
	blockedUsers := discussionCache.UnblockUser(discussion, user)
	if _, exists := blockedUsers[uint(5540)]; exists {
		t.Fail()
	}

}

func TestAdminDeleteUndeletePost(t *testing.T) {
	t.Fail()
}

func TestAdminGetReports(t *testing.T) {
	t.Fail()
}

func TestAdminCreateAndGetComments(t *testing.T) {
	t.Fail()
}

func TestLockUnlockDiscussion(t *testing.T) {
	t.Fail()
}

func TestAdminPremodDiscussion(t *testing.T) {
	t.Fail()
}

func TestAdminDeleteDiscussion(t *testing.T) {
	t.Fail()
}

func TestAdminMoveDiscussion(t *testing.T) {
	t.Fail()
}

func TestAdminEraseDiscussion(t *testing.T) {
	t.Fail()
}

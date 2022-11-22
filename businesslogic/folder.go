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
	"html"
	"justthetalk/model"
	"justthetalk/utils"
	"time"

	"errors"

	"sync"

	"gorm.io/gorm"
)

var postFormatterOnce sync.Once
var postFormatter *utils.PostFormatter

var bannedWordsOnce sync.Once
var bannedWords *BannedWordsList

func PostFormatter() *utils.PostFormatter {
	postFormatterOnce.Do(func() {
		postFormatter = utils.NewPostFormatter()
	})
	return postFormatter
}

func BannedWords() *BannedWordsList {
	bannedWordsOnce.Do(func() {
		bannedWords = NewBannedWordsList()
	})
	return bannedWords
}

func GetDiscussions(folder *model.Folder, pageStart int, pageSize int, user *model.User, db *gorm.DB) []*model.FrontPageEntry {

	userId := 0
	if user != nil {
		userId = int(user.Id)
	}

	discussions := make([]*model.FrontPageEntry, 0)
	if result := db.Raw("call get_folder_discussions(?, ?, ?, ?)", folder.Id, userId, pageStart, pageSize).Scan(&discussions); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	utils.FormatFrontPageEntries(discussions)

	return discussions

}

func GetDiscussionsBefore(folder *model.Folder, beforeDate time.Time, pageSize int, user *model.User, db *gorm.DB) []*model.FrontPageEntry {

	userId := 0
	if user != nil {
		userId = int(user.Id)
	}

	discussions := make([]*model.FrontPageEntry, 0)
	if result := db.Raw("call get_folder_discussions_before(?, ?, ?, ?)", folder.Id, userId, beforeDate, pageSize).Scan(&discussions); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	utils.FormatFrontPageEntries(discussions)

	return discussions

}

func validateDiscussion(folder *model.Folder, discussion *model.Discussion, db *gorm.DB) {

	discussion.Title = html.EscapeString(discussion.Title)
	discussion.Header = html.EscapeString(discussion.Header)

	if len(discussion.Title) > 128 {
		utils.PanicWithWrapper(errors.New("Title too long"), utils.ErrBadRequest)
	}

	if len(discussion.Header) > 1024 {
		utils.PanicWithWrapper(errors.New("Header too long"), utils.ErrBadRequest)
	}

	var duplicateDiscussion model.Discussion
	if result := db.Raw("call find_discussion_by_title(?, ?)", folder.Id, discussion.Title).First(&duplicateDiscussion); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	}

	if duplicateDiscussion.Id != 0 && duplicateDiscussion.Id != discussion.Id {
		utils.PanicWithWrapper(errors.New("A discussion with that title already exists"), utils.ErrBadRequest)
	}

}

func CreateDiscussion(folder *model.Folder, discussion *model.Discussion, user *model.User, userCache *UserCache, discussionCache *DiscussionCache, db *gorm.DB) *model.Discussion {

	if user.IsPremoderate || user.AccountExpired || user.AccountLocked || !user.Enabled {
		panic(utils.ErrForbidden)
	}

	validateDiscussion(folder, discussion, db)

	lockedParam := 0
	if BannedWords().CheckForBannedWords(discussion.Title) || BannedWords().CheckForBannedWords(discussion.Header) {
		lockedParam = 1
	}

	var created model.Discussion
	if result := db.Raw("call create_discussion(?, ?, ?, ?, ?)", folder.Id, discussion.Title, discussion.Header, user.Id, lockedParam).First(&created); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	created.HeaderMarkup = PostFormatter().ApplyPostFormatting(created.Header, &created)
	created.Url = utils.UrlForDiscussion(folder, &created)

	if discussion.IsSubscribed {
		SetDiscussionSubscriptionStatus(&created, user, db, userCache)
	}

	discussionCache.Put(&created)

	return &created

}

func EditDiscussion(folder *model.Folder, discussion *model.Discussion, user *model.User, discussionCache *DiscussionCache, db *gorm.DB) *model.Discussion {

	if user.IsPremoderate || user.AccountExpired || user.AccountLocked || !user.Enabled {
		panic(utils.ErrForbidden)
	}

	validateDiscussion(folder, discussion, db)

	lockedParam := 0
	if BannedWords().CheckForBannedWords(discussion.Title) || BannedWords().CheckForBannedWords(discussion.Header) {
		lockedParam = 1
	}

	var edited model.Discussion
	if result := db.Raw("call edit_discussion(?, ?, ?, ?, ?, ?)", folder.Id, discussion.Id, discussion.Title, discussion.Header, user.Id, lockedParam).Scan(&edited); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	edited.HeaderMarkup = PostFormatter().ApplyPostFormatting(edited.Header, &edited)
	edited.Url = utils.UrlForDiscussion(folder, discussion)

	discussionCache.Put(&edited)

	return &edited

}

func DeleteDiscussion(folder *model.Folder, discussion *model.Discussion, user *model.User, db *gorm.DB) *model.Discussion {

	if discussion.FolderId != folder.Id {
		panic(utils.ErrBadRequest)
	}

	if !(discussion.CreatedByUserId == user.Id) {
		panic(utils.ErrForbidden)
	}

	var deleted model.Discussion
	if result := db.Raw("call delete_discussion(?, ?)", discussion.Id, model.DiscussionStatusDeletedByUser).Take(&deleted); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return &deleted

}

func GetPosts(folder *model.Folder, discussion *model.Discussion, user *model.User, pageStart int64, pageSize int, db *gorm.DB) []*model.Post {

	posts := make([]*model.Post, 0)

	userId := 0
	if user != nil {
		userId = int(user.Id)
	}

	if result := db.Raw("call get_discussion_posts(?, ?, ?, ?, ?)", userId, folder.Id, discussion.Id, pageStart, pageSize).Scan(&posts); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	for _, post := range posts {

		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
		post.Url = utils.UrlForPost(folder, discussion, post)

		if user == nil || !user.IsAdmin {
			switch post.Status {
			case model.PostStatusPostedByAdmin:
				post.CreatedByUserId = 1
				post.CreatedByUsername = "JUSTtheTalk"

			case model.PostStatusSuspendedByAdmin:
			case model.PostStatusDeletedByAdmin:
			case model.PostStatusDeletedByUser:
			case model.PostStatusInvisible:
				post.CreatedByUserId = 0
				post.CreatedByUsername = ""
				post.Text = ""

			}
		}
	}

	return posts

}

func CreatePost(folder *model.Folder, discussion *model.Discussion, user *model.User, post *model.Post, discussionCache *DiscussionCache, userCache *UserCache, db *gorm.DB) *model.Post {

	if user.AccountExpired || user.AccountLocked || !user.Enabled {
		panic(utils.ErrForbidden)
	}

	if discussion.IsBlocked {
		panic(utils.ErrForbidden)
	}

	if post.PostAsAdmin && !user.IsAdmin {
		panic(utils.ErrForbidden)
	}

	status := model.PostStatusOK
	if post.PostAsAdmin && user.IsAdmin {
		status = model.PostStatusPostedByAdmin
	} else if user.IsPremoderate || discussion.IsPremoderate {
		status = model.PostStatusSuspendedByAdmin
	} else if user.IsWatch {
		status = model.PostStatusWatch
	} else if BannedWords().CheckForBannedWords(post.Text) {
		status = model.PostStatusSuspendedByAdmin
	}

	post.Text = html.EscapeString(post.Text)

	var created model.Post
	if result := db.Raw("call create_discussion_post(?, ?, ?, ?, ?)", folder.Id, discussion.Id, post.Text, status, user.Id).First(&created); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	if created.Id == 0 {
		panic(utils.ErrInternalError)
	}

	discussion.LastPostDate = created.CreatedDate
	discussion.PostCount = created.PostNum
	discussionCache.Put(discussion)

	created.Markup = PostFormatter().ApplyPostFormatting(created.Text, discussion)
	created.Url = utils.UrlForPost(folder, discussion, &created)

	if post.SubscribeToDiscussion {
		SetDiscussionSubscriptionStatus(discussion, user, db, userCache)
	}

	if post.DiscussionId == 52283 {
		db.Table("user").Where("id = ?", user.Id).Update("account_locked", 1)
	}

	return &created

}

func EditPost(folder *model.Folder, discussion *model.Discussion, user *model.User, update *model.Post, db *gorm.DB) *model.Post {

	var post model.Post

	if discussion.IsBlocked {
		panic(utils.ErrForbidden)
	}

	update.Text = html.EscapeString(update.Text)

	if result := db.Raw("call edit_discussion_post(?, ?, ?, ?, ?)", folder.Id, discussion.Id, update.Id, update.Text, user.Id).Scan(&post); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	if post.Id == 0 {
		panic(utils.ErrNotModified)
	}

	post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
	post.Url = utils.UrlForPost(folder, discussion, &post)

	return &post

}

func DeletePost(folder *model.Folder, discussion *model.Discussion, user *model.User, postId uint, db *gorm.DB) *model.Post {

	var post model.Post
	if result := db.Raw("call get_post(?)", postId).First(&post); result.Error != nil {
		panic(utils.ErrBadRequest)
	}

	if !(user.Id == post.CreatedByUserId || user.IsAdmin) {
		panic(utils.ErrForbidden)
	}

	if result := db.Raw("call delete_discussion_post(?, ?, ?, ?)", folder.Id, discussion.Id, postId, user.Id).First(&post); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	if post.Id == 0 {
		panic(utils.ErrNotModified)
	}

	if user.IsAdmin {
		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
	} else {
		post.Text = ""
		post.Markup = ""
	}

	post.Url = utils.UrlForPost(folder, discussion, &post)

	return &post

}

func GetPost(postId uint, db *gorm.DB) *model.Post {
	var post model.Post
	if result := db.Raw("call get_post(?)", postId).First(&post); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}
	return &post
}

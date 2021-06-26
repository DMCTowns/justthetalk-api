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
	"justthetalk/model"
	"justthetalk/utils"

	"gorm.io/gorm"
)

func FetchBlockedUsers(discussion *model.Discussion, db *gorm.DB) map[uint]*model.BlockedDiscussionUser {

	var blockedUsersList []*model.BlockedDiscussionUser
	if result := db.Raw("call get_blocked_discussion_users(?)", discussion.Id).Scan(&blockedUsersList); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return mapBlockedUsers(blockedUsersList)

}

func mapBlockedUsers(blockedUsersList []*model.BlockedDiscussionUser) map[uint]*model.BlockedDiscussionUser {

	var blockedUserMap map[uint]*model.BlockedDiscussionUser

	blockedUserMap = make(map[uint]*model.BlockedDiscussionUser)
	for _, b := range blockedUsersList {
		blockedUserMap[b.UserId] = b
	}

	return blockedUserMap

}

func BlockUser(discussion *model.Discussion, user *model.User, db *gorm.DB) map[uint]*model.BlockedDiscussionUser {

	var blockedUsersList []*model.BlockedDiscussionUser
	if result := db.Raw("call block_discussion_user(?, ?, ?)", discussion.Id, user.Id, 1).Scan(&blockedUsersList); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}
	return mapBlockedUsers(blockedUsersList)

}

func UnblockUser(discussion *model.Discussion, user *model.User, db *gorm.DB) map[uint]*model.BlockedDiscussionUser {

	var blockedUsersList []*model.BlockedDiscussionUser
	if result := db.Raw("call block_discussion_user(?, ?, ?)", discussion.Id, user.Id, 0).Scan(&blockedUsersList); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}
	return mapBlockedUsers(blockedUsersList)

}

func AdminDeleteNoUndeletePost(postId uint, discussion *model.Discussion, deleteNotUndelete bool, db *gorm.DB) *model.Post {

	var post model.Post

	postStatus := model.PostStatusDeletedByAdmin
	if !deleteNotUndelete {
		postStatus = model.PostStatusOK
	}

	if result := db.Raw("call set_post_status(?, ?, ?, ?)", discussion.Id, postId, postStatus, 0).First(&post); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)

	return &post

}

func GetModerationQueue(folderCache *FolderCache, discussionCache *DiscussionCache, db *gorm.DB) []*model.Post {

	results := make([]*model.Post, 0)

	if result := db.Raw("call get_moderation_queue()").Find(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	// ugh!
	for _, post := range results {

		if !(post.Status == model.PostStatusSuspendedByAdmin || post.Status == model.PostStatusWatch) {
			post.UserReports = make([]*model.PostReport, 0)
			db.Table("post_report").Where("post_id = ?", post.Id).Find(&post.UserReports)
		}

		post.ModeratorComments = make([]*model.ModeratorComment, 0)
		db.Table("moderator_comment").Where("post_id = ?", post.Id).Find(&post.ModeratorComments)

		discussion := discussionCache.UnsafeGet(post.DiscussionId)
		folder := folderCache.UnsafeGet(discussion.FolderId)

		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
		post.Url = utils.UrlForPost(folder, discussion, post)

	}

	return results

}

func GetDiscussionReports(discussion *model.Discussion, db *gorm.DB) []*model.PostReport {

	results := make([]*model.PostReport, 0)
	if result := db.Raw("call get_reports_by_discussion(?)", discussion.Id).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func GetComments(discussion *model.Discussion, db *gorm.DB) []*model.ModeratorComment {

	results := make([]*model.ModeratorComment, 0)
	if result := db.Raw("call get_comments_by_discussion(?)", discussion.Id).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func CreateComment(comment *model.ModeratorComment, discussion *model.Discussion, post *model.Post, user *model.User, db *gorm.DB) ([]*model.ModeratorComment, *model.Post) {

	results := make([]*model.ModeratorComment, 0)
	if result := db.Raw("call create_admin_coment(?, ?, ?, ?)", post.Id, user.Id, comment.Body, comment.Vote).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	totalVote := 0
	for _, comment := range results {
		totalVote += comment.Vote
	}

	moderationThreshold := 2
	if post.Status == model.PostStatusSuspendedByAdmin || post.Status == model.PostStatusWatch {
		moderationThreshold = 1
	}

	if utils.Abs(totalVote) >= moderationThreshold {

		if totalVote < 0 {
			post.Status = model.PostStatusDeletedByAdmin
		} else {
			post.Status = model.PostStatusOK
		}

		var post model.Post
		if result := db.Raw("call set_post_status(?, ?)", post.Id, post.Status, totalVote).First(&post); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}

		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)

	}

	return results, post

}

func LockDiscussion(discussion *model.Discussion, lockState int, discussionCache *DiscussionCache, db *gorm.DB) {

	if result := db.Raw("call lock_discussion(?, ?)", discussion.Id, lockState).Take(discussion); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	discussionCache.Put(discussion)

}

func PremoderateDiscussion(discussion *model.Discussion, premodState int, discussionCache *DiscussionCache, db *gorm.DB) {

	if result := db.Raw("call lock_premoderate_discussiondiscussion(?, ?)", discussion.Id, premodState).Take(discussion); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	discussionCache.Put(discussion)

}

func AdminDeleteDiscussion(discussion *model.Discussion, deleteState int, discussionCache *DiscussionCache, db *gorm.DB) {

	var status = model.DiscussionStatusOk
	if deleteState == 1 {
		status = model.DiscussionStatusDeletedByAdmin
	}

	if result := db.Raw("call delete_discussion(?, ?)", discussion.Id, status).Take(discussion); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	discussionCache.Put(discussion)

}

func MoveDiscussion(discussion *model.Discussion, targetFolder *model.Folder, discussionCache *DiscussionCache, db *gorm.DB) {

	if result := db.Exec("call move_discussion(?, ?)", discussion.Id, targetFolder.Id); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	discussionCache.Put(discussion)

}

func EraseDiscussion(discussion *model.Discussion, discussionCache *DiscussionCache, db *gorm.DB) {

	if result := db.Exec("call erase_discussion(?)", discussion.Id); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	discussionCache.Flush(discussion.Id)

}

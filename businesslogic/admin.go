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

func AdminDeleteNoUndeletePost(postId uint, folder *model.Folder, discussion *model.Discussion, deleteNotUndelete bool, db *gorm.DB) *model.Post {

	var post model.Post

	postStatus := model.PostStatusDeletedByAdmin
	if !deleteNotUndelete {
		postStatus = model.PostStatusOK
	}

	if result := db.Raw("call set_post_status(?, ?, ?, ?)", discussion.Id, postId, postStatus, 0).First(&post); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
	post.Url = utils.UrlForPost(folder, discussion, &post)

	return &post

}

func GetModerationHistory(pageStart int, pageSize int, folderCache *FolderCache, discussionCache *DiscussionCache, db *gorm.DB) []*model.Post {

	posts := make([]*model.Post, 0)

	if result := db.Raw("call get_moderated_posts(?, ?)", pageStart*pageSize, pageSize).Find(&posts); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	for _, post := range posts {
		discussion := discussionCache.UnsafeGet(post.DiscussionId)
		folder := folderCache.UnsafeGet(discussion.FolderId)
		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
		post.Url = utils.UrlForPost(folder, discussion, post)
	}

	return posts

}

func GetModerationQueue(folderCache *FolderCache, discussionCache *DiscussionCache, db *gorm.DB) []*model.Post {

	posts := make([]*model.Post, 0)

	if result := db.Raw("call get_moderation_queue()").Find(&posts); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	for _, post := range posts {
		discussion := discussionCache.UnsafeGet(post.DiscussionId)
		folder := folderCache.UnsafeGet(discussion.FolderId)
		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
		post.Url = utils.UrlForPost(folder, discussion, post)
	}

	return posts

}

func GetReportsByPost(postId uint, db *gorm.DB) []*model.PostReport {

	results := make([]*model.PostReport, 0)
	if result := db.Raw("call get_reports_by_post(?)", postId).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func GetCommentsByPost(postId uint, db *gorm.DB) []*model.ModeratorComment {

	results := make([]*model.ModeratorComment, 0)
	if result := db.Raw("call get_comments_by_post(?)", postId).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func GetReportsByDiscussion(discussion *model.Discussion, db *gorm.DB) []*model.PostReport {

	results := make([]*model.PostReport, 0)
	if result := db.Raw("call get_reports_by_discussion(?)", discussion.Id).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func GetCommentsByDiscussion(discussion *model.Discussion, db *gorm.DB) []*model.ModeratorComment {

	results := make([]*model.ModeratorComment, 0)
	if result := db.Raw("call get_comments_by_discussion(?)", discussion.Id).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func CreateComment(comment *model.ModeratorComment, folder *model.Folder, discussion *model.Discussion, post *model.Post, user *model.User, db *gorm.DB) ([]*model.ModeratorComment, *model.Post) {

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
		if result := db.Raw("call set_post_status(?, ?, ?, ?)", discussion.Id, post.Id, post.Status, totalVote).First(&post); result.Error != nil {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}

		post.Markup = PostFormatter().ApplyPostFormatting(post.Text, discussion)
		post.Url = utils.UrlForPost(folder, discussion, &post)

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

func SearchUsers(searchTerm string, db *gorm.DB) []*model.UserSearchResults {

	results := make([]*model.UserSearchResults, 0)

	if result := db.Raw("call search_users(?)", fmt.Sprintf("%%%s%%", searchTerm)).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results
}

func FilterUsers(filterKey string, db *gorm.DB) []*model.UserSearchResults {

	results := make([]*model.UserSearchResults, 0)

	command := ""
	switch filterKey {
	case "premod":
		command = "search_users_premod"
	case "watch":
		command = "search_users_watch"
	case "locked":
		command = "search_users_locked"
	case "recent":
		command = "search_users_recent"
	default:
		panic(utils.ErrBadRequest)
	}

	if result := db.Raw(fmt.Sprintf("call %s", command)).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

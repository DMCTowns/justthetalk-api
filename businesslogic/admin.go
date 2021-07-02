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
	"time"

	"github.com/gosimple/slug"
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

func BlockUnblockUser(discussion *model.Discussion, targetUser *model.User, blockNotUnblock bool, adminUser *model.User, db *gorm.DB) map[uint]*model.BlockedDiscussionUser {

	state := 0
	eventType := model.UserHistoryAdminDiscussionUnblocked
	if blockNotUnblock {
		state = 1
		eventType = model.UserHistoryAdminDiscussionBlocked
	}

	var blockedUsersList []*model.BlockedDiscussionUser
	if result := db.Raw("call block_discussion_user(?, ?, ?)", discussion.Id, targetUser.Id, state).Scan(&blockedUsersList); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	CreateUserHistory(eventType, fmt.Sprintf("DiscussionId: %d, Actioned by: %s", discussion.Id, adminUser.Username), targetUser, db)

	return mapBlockedUsers(blockedUsersList)

}

func AdminDeleteNoUndeletePost(postId uint, folder *model.Folder, discussion *model.Discussion, deleteNotUndelete bool, adminUser *model.User, userCache *UserCache, db *gorm.DB) *model.Post {

	var post model.Post

	postStatus := model.PostStatusDeletedByAdmin
	if !deleteNotUndelete {
		postStatus = model.PostStatusOK
	}

	// TODO put this in a transaction
	if result := db.Raw("call set_post_status(?, ?, ?, ?)", discussion.Id, postId, postStatus, 0).First(&post); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	targetUser := userCache.Get(post.CreatedByUserId)
	var eventType string
	if deleteNotUndelete {
		eventType = model.UserHistoryAdminPostDelete
	} else {
		eventType = model.UserHistoryAdminPostUndelete
	}
	CreateUserHistory(eventType, fmt.Sprintf("Actioned by: %s", adminUser.Username), targetUser, db)

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

func CreateComment(comment *model.ModeratorComment, folder *model.Folder, discussion *model.Discussion, post *model.Post, user *model.User, userCache *UserCache, db *gorm.DB) ([]*model.ModeratorComment, *model.Post) {

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

		var result string
		if totalVote < 0 {
			post.Status = model.PostStatusDeletedByAdmin
			result = "DELETE"
		} else {
			post.Status = model.PostStatusOK
			result = "KEEP"
		}

		targetUser := userCache.Get(post.CreatedByUserId)
		CreateUserHistory(model.UserHistoryAdminPostModerated, fmt.Sprintf("PostId: %d, %s", post.Id, result), targetUser, db)

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

func SetUserStatus(targetUser *model.User, fieldMap map[string]interface{}, adminUser *model.User, userCache *UserCache, db *gorm.DB) *model.User {

	db.Transaction(func(tx *gorm.DB) error {

		var result *gorm.DB
		for k, v := range fieldMap {

			var eventType string
			eventData := fmt.Sprintf("Actioned by: %s", adminUser.Username)

			switch k {
			case "enabled":
				result = tx.Table("user").Where("id = ?", targetUser.Id).Update("enabled", v)
				if v.(bool) {
					eventType = model.UserHistoryAdminAccountDeleteEnabled
				} else {
					eventType = model.UserHistoryAdminAccountDeleteDisabled
				}

			case "accountLocked":
				result = tx.Table("user").Where("id = ?", targetUser.Id).Update("account_locked", v)
				if v.(bool) {
					eventType = model.UserHistoryAdminAccountLockedEnabled
				} else {
					eventType = model.UserHistoryAdminAccountLockedDisabled
				}

			case "isPremoderate":
				result = tx.Table("user_options").Where("user_id = ?", targetUser.Id).Update("premoderate", v)
				if v.(bool) {
					eventType = model.UserHistoryAdminPremodEnabled
				} else {
					eventType = model.UserHistoryAdminPremodDisabled
				}

			case "isWatch":
				result = tx.Table("user_options").Where("user_id = ?", targetUser.Id).Update("watch", v)
				if v.(bool) {
					eventType = model.UserHistoryAdminWatchEnabled
				} else {
					eventType = model.UserHistoryAdminWatchDisabled
				}

			}

			if result.Error != nil {
				break
			}

			result = tx.Table("user").Where("id = ?", targetUser.Id).Update("last_updated", time.Now())
			if result.Error != nil {
				break
			}

			CreateUserHistory(eventType, eventData, targetUser, tx)

		}

		if result.Error != nil {
			return result.Error
		} else {
			return nil
		}

	})

	userCache.Flush(targetUser)

	return userCache.Get(targetUser.Id)

}

func GetUserHistory(targetUser *model.User, db *gorm.DB) []*model.UserHistory {

	results := make([]*model.UserHistory, 0)
	if result := db.Raw("call get_user_history(?)", targetUser.Id).Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return results

}

func GetUserDiscussionBlocks(db *gorm.DB) []*model.DiscussionBlock {

	results := make([]*model.DiscussionBlock, 0)
	if result := db.Raw("call get_user_discussion_blocks()").Scan(&results); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	for _, item := range results {
		slugText := slug.Make(item.DiscussionTitle)
		item.Url = fmt.Sprintf("/%s/%d/%s/1", item.FolderKey, item.DiscussionId, slugText)
	}

	return results

}

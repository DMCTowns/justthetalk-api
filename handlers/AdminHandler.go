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

package handlers

import (
	"encoding/json"
	"net/http"

	"justthetalk/businesslogic"
	"justthetalk/model"
	"justthetalk/utils"

	"gorm.io/gorm"
)

type AdminHandler struct {
	userCache       *businesslogic.UserCache
	folderCache     *businesslogic.FolderCache
	discussionCache *businesslogic.DiscussionCache
	postProcessor   *businesslogic.PostProcessor
	postFormatter   *utils.PostFormatter
}

func NewAdminHandler(userCache *businesslogic.UserCache, folderCache *businesslogic.FolderCache, discussionCache *businesslogic.DiscussionCache, postProcessor *businesslogic.PostProcessor) *AdminHandler {

	return &AdminHandler{
		userCache:       userCache,
		folderCache:     folderCache,
		discussionCache: discussionCache,
		postProcessor:   postProcessor,
		postFormatter:   utils.NewPostFormatter(),
	}

}

func (h *AdminHandler) GetModerationHistory(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		pageStart := utils.ExtractQueryInt("start", req)
		pageSize := utils.ExtractQueryInt("size", req)

		results := businesslogic.GetModerationHistory(pageStart, pageSize, h.folderCache, h.discussionCache, db)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) GetModerationQueue(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		results := businesslogic.GetModerationQueue(h.folderCache, h.discussionCache, db)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) GetReportsByPost(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		h.discussionCache.Get(discussionId, user)

		postId := utils.ExtractVarInt("postId", req)

		results := businesslogic.GetReportsByPost(postId, db)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) GetCommentsByPost(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		h.discussionCache.Get(discussionId, user)

		postId := utils.ExtractVarInt("postId", req)

		results := businesslogic.GetCommentsByPost(postId, db)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) GetReportsByDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		discussion := h.discussionCache.Get(discussionId, user)

		results := businesslogic.GetReportsByDiscussion(discussion, db)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) GetCommentsByDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		discussion := h.discussionCache.Get(discussionId, user)

		results := businesslogic.GetCommentsByDiscussion(discussion, db)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) CreateComment(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		postId := utils.ExtractVarInt("postId", req)

		var comment model.ModeratorComment
		if err := json.NewDecoder(req.Body).Decode(&comment); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		discussion := h.discussionCache.Get(discussionId, user)
		folder := h.folderCache.Get(discussion.FolderId, user)
		post := businesslogic.GetPost(postId, db)

		if post.DiscussionId != discussionId {
			panic(utils.ErrBadRequest)
		}

		results, post := businesslogic.CreateComment(&comment, folder, discussion, post, user, db)

		h.postProcessor.PublishPost(post)

		return http.StatusOK, results, ""

	})
}

func (h *AdminHandler) LockDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		lockState := utils.ExtractQueryInt("state", req)

		discussion := h.discussionCache.Get(discussionId, user)

		businesslogic.LockDiscussion(discussion, lockState, h.discussionCache, db)

		return http.StatusOK, discussion, ""

	})
}

func (h *AdminHandler) PremoderateDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		premodState := utils.ExtractQueryInt("state", req)

		discussion := h.discussionCache.Get(discussionId, user)
		businesslogic.PremoderateDiscussion(discussion, premodState, h.discussionCache, db)

		return http.StatusOK, discussion, ""

	})
}

func (h *AdminHandler) DeleteDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		deleteState := utils.ExtractQueryInt("state", req)

		discussion := h.discussionCache.Get(discussionId, user)
		businesslogic.AdminDeleteDiscussion(discussion, deleteState, h.discussionCache, db)

		return http.StatusOK, discussion, ""

	})
}

func (h *AdminHandler) MoveDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		targetFolderId := utils.ExtractQueryInt("targetFolderId", req)

		discussion := h.discussionCache.Get(discussionId, user)
		targetFolder := h.folderCache.Get(uint(targetFolderId), user)

		businesslogic.MoveDiscussion(discussion, targetFolder, h.discussionCache, db)

		return http.StatusOK, discussion, ""

	})
}

func (h *AdminHandler) EraseDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		discussion := h.discussionCache.Get(discussionId, user)

		businesslogic.EraseDiscussion(discussion, h.discussionCache, db)

		return http.StatusOK, nil, "Discussion erased"

	})
}

func (h *AdminHandler) GetBlockedUsers(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)

		discussion := h.discussionCache.Get(discussionId, user)
		blockedUsers := h.discussionCache.BlockedUsers(discussion)

		return http.StatusOK, blockedUsers, ""

	})
}

func (h *AdminHandler) BlockUserDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		blockUserId := utils.ExtractVarInt("userId", req)

		discussion := h.discussionCache.Get(discussionId, user)
		blockUser := h.userCache.Get(blockUserId)

		blockedUsers := h.discussionCache.BlockUser(discussion, blockUser)

		return http.StatusOK, blockedUsers, ""

	})
}

func (h *AdminHandler) UnblockUserDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		unblockUserId := utils.ExtractVarInt("userId", req)

		discussion := h.discussionCache.Get(discussionId, user)
		unblockUser := h.userCache.Get(unblockUserId)

		blockedUsers := h.discussionCache.UnblockUser(discussion, unblockUser)

		return http.StatusOK, blockedUsers, ""

	})
}

func (h *AdminHandler) DeletePost(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		postId := utils.ExtractVarInt("postId", req)

		discussion := h.discussionCache.Get(discussionId, user)

		post := businesslogic.AdminDeleteNoUndeletePost(postId, discussion, true, db)

		post.Markup = h.postFormatter.ApplyPostFormatting(post.Text, discussion)
		h.postProcessor.PublishPost(post)

		return http.StatusOK, post, ""

	})
}

func (h *AdminHandler) UndeletePost(res http.ResponseWriter, req *http.Request) {
	utils.AdminOnlyHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		postId := utils.ExtractVarInt("postId", req)

		discussion := h.discussionCache.Get(discussionId, user)

		post := businesslogic.AdminDeleteNoUndeletePost(postId, discussion, false, db)

		post.Markup = h.postFormatter.ApplyPostFormatting(post.Text, discussion)
		h.postProcessor.PublishPost(post)

		return http.StatusOK, post, ""

	})

}

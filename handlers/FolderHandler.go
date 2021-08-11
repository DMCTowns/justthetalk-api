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
	"justthetalk/businesslogic"
	"justthetalk/model"
	"justthetalk/utils"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

var (
	postCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "justthetalk_post_count",
		Help: "Count of new posts",
	}, []string{"folder"})

	discussionCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "justthetalk_discussion_count",
		Help: "Count of new discussions",
	}, []string{"folder"})
)

type FolderHandler struct {
	userCache       *businesslogic.UserCache
	folderCache     *businesslogic.FolderCache
	discussionCache *businesslogic.DiscussionCache
	postProcessor   *businesslogic.PostProcessor
}

func NewFolderHandler(userCache *businesslogic.UserCache, folderCache *businesslogic.FolderCache, discussionCache *businesslogic.DiscussionCache, postProcessor *businesslogic.PostProcessor) *FolderHandler {

	return &FolderHandler{
		userCache:       userCache,
		folderCache:     folderCache,
		discussionCache: discussionCache,
		postProcessor:   postProcessor,
	}

}

func (h *FolderHandler) GetFolders(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		subsMap := make(map[uint]*model.UserFolderSubscription)
		if user != nil {
			subsList := businesslogic.GetFolderSubscriptions(user, db)
			for _, sub := range subsList {
				subsMap[sub.FolderId] = sub
			}
		}

		var data []*model.Folder
		for _, folder := range h.folderCache.Entries() {
			shouldAdd := folder.Type == model.FolderTypeNormal || (user != nil && user.IsAdmin)
			if shouldAdd {

				var folderCopy model.Folder
				if err := copier.Copy(&folderCopy, &folder); err != nil {
					panic(err)
				}

				_, folderCopy.IsSubscribed = subsMap[folderCopy.Id]

				data = append(data, &folderCopy)

			}
		}

		return http.StatusOK, data, ""

	})
}

func (h *FolderHandler) GetFolder(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		folder := h.folderCache.Get(folderId, user)

		var folderCopy model.Folder
		if err := copier.Copy(&folderCopy, &folder); err != nil {
			panic(err)
		}

		folderCopy.IsSubscribed = businesslogic.GetFolderSubscriptionStatus(&folderCopy, user, db)

		if folderCopy.Type == model.FolderTypeNormal {
			return http.StatusOK, folderCopy, ""
		} else {
			if user != nil && user.IsAdmin {
				return http.StatusOK, folderCopy, ""
			} else {
				panic(utils.ErrForbidden)
			}
		}

	})
}

func (h *FolderHandler) GetDiscussions(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		pageSize, pageStart := utils.ExtractPageSizeAndStart(req)

		folder := h.folderCache.Get(folderId, user)

		discussions := businesslogic.GetDiscussions(folder, pageStart, pageSize, user, db)

		return http.StatusOK, discussions, ""

	})
}

func (h *FolderHandler) CreateDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		folder := h.folderCache.Get(folderId, user)

		var discussion model.Discussion
		if err := json.NewDecoder(req.Body).Decode(&discussion); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		created := businesslogic.CreateDiscussion(folder, &discussion, user, h.userCache, h.discussionCache, db)

		discussionCount.WithLabelValues(folder.Key).Inc()

		return http.StatusOK, created, ""

	})
}

func (h *FolderHandler) GetDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)

		folder := h.folderCache.Get(folderId, user)
		discussion := h.discussionCache.Get(discussionId, user)

		if discussion == nil {
			panic(utils.ErrNotFound)
		}

		if discussion.FolderId != folder.Id {
			panic(utils.ErrBadRequest)
		}

		if user != nil {
			discussion.IsSubscribed = businesslogic.GetDiscussionSubscriptionStatus(discussion, user, db)
			discussion.IsBlocked = h.discussionCache.IsBlocked(discussion, user)
		}

		return http.StatusOK, discussion, ""

	})
}

func (h *FolderHandler) EditDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)

		folder := h.folderCache.Get(folderId, user)

		var discussion model.Discussion
		if err := json.NewDecoder(req.Body).Decode(&discussion); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}
		discussion.Id = discussionId

		edited := businesslogic.EditDiscussion(folder, &discussion, user, h.discussionCache, db)

		return http.StatusOK, edited, ""

	})
}

func (h *FolderHandler) DeleteDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)

		folder := h.folderCache.Get(folderId, user)

		discussion := h.discussionCache.Get(discussionId, user)
		if discussion == nil {
			panic(utils.ErrNotFound)
		}

		deleted := businesslogic.DeleteDiscussion(folder, discussion, user, db)

		h.discussionCache.Put(deleted)

		return http.StatusOK, deleted, ""

	})
}

func (h *FolderHandler) GetPosts(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)

		folder := h.folderCache.Get(folderId, user)
		discussion := h.discussionCache.Get(discussionId, user)

		pageStart := utils.ExtractQueryInt64("start", req)

		pageSize := utils.ExtractQueryInt("size", req)
		if pageSize == 0 {
			pageSize = 20
		}

		posts := businesslogic.GetPosts(folder, discussion, user, pageStart, pageSize, db)

		return http.StatusOK, posts, ""

	})
}

func (h *FolderHandler) CreatePost(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)

		var post model.Post
		if err := json.NewDecoder(req.Body).Decode(&post); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		folder := h.folderCache.Get(folderId, user)
		discussion := h.discussionCache.Get(discussionId, user)

		created := businesslogic.CreatePost(folder, discussion, user, &post, h.discussionCache, h.userCache, db)
		h.postProcessor.PublishPost(created)

		returnPostsFromPostNum := created.PostNum

		lastBookmark := businesslogic.GetDiscussionBookmark(user, discussion, db)
		if lastBookmark != nil {
			returnPostsFromPostNum = lastBookmark.LastPostCount + 1
		}

		posts := businesslogic.GetPosts(folder, discussion, user, returnPostsFromPostNum, 20, db)

		postCount.WithLabelValues(folder.Key).Inc()

		return http.StatusOK, posts, ""

	})
}

func (h *FolderHandler) EditPost(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)
		postId := utils.ExtractVarInt("postId", req)

		folder := h.folderCache.Get(folderId, user)
		discussion := h.discussionCache.Get(discussionId, user)

		var post model.Post
		if err := json.NewDecoder(req.Body).Decode(&post); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		post.Id = postId

		updated := businesslogic.EditPost(folder, discussion, user, &post, db)

		h.postProcessor.PublishPost(updated)

		return http.StatusOK, updated, ""

	})
}

func (h *FolderHandler) DeletePost(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		discussionId := utils.ExtractVarInt("discussionId", req)
		postId := utils.ExtractVarInt("postId", req)

		folder := h.folderCache.Get(folderId, user)
		discussion := h.discussionCache.Get(discussionId, user)

		updated := businesslogic.DeletePost(folder, discussion, user, postId, db)

		h.postProcessor.PublishPost(updated)

		return http.StatusOK, updated, ""

	})
}

func (h *FolderHandler) SubscribeToDiscussion(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		discussion := h.discussionCache.Get(discussionId, user)

		if req.Method == http.MethodPost {
			businesslogic.SetDiscussionSubscriptionStatus(discussion, user, db, h.userCache)
			discussion.IsSubscribed = true
		} else {
			businesslogic.UnsetDiscussionSubscriptionStatus(discussion, user, db, h.userCache)
			discussion.IsSubscribed = false
		}

		return http.StatusOK, discussion, ""

	})
}

func (h *FolderHandler) SubscribeToFolder(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		folderId := utils.ExtractVarInt("folderId", req)
		folder := h.folderCache.Get(folderId, user)

		var folderCopy model.Folder
		if err := copier.Copy(&folderCopy, &folder); err != nil {
			panic(err)
		}

		if req.Method == http.MethodPost {
			businesslogic.SetFolderSubscriptionStatus(folder, user, db, h.userCache)
			folderCopy.IsSubscribed = true
		} else {
			businesslogic.UnsetFolderSubscriptionStatus(folder, user, db, h.userCache)
			folderCopy.IsSubscribed = false
		}

		return http.StatusOK, folderCopy, ""

	})
}

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
	"justthetalk/businesslogic"
	"justthetalk/model"
	"justthetalk/utils"
	"net/http"

	"gorm.io/gorm"
)

type FrontPageHandler struct {
	userCache       *businesslogic.UserCache
	discussionCache *businesslogic.DiscussionCache
}

func NewFrontPageHandler(userCache *businesslogic.UserCache, discussionCache *businesslogic.DiscussionCache) *FrontPageHandler {
	return &FrontPageHandler{
		userCache:       userCache,
		discussionCache: discussionCache,
	}
}

func (h *FrontPageHandler) GetFrontPage(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		pageSize := 0
		pageStart := 0

		startParam := req.URL.Query().Get("start")
		if len(startParam) > 0 {
			pageStart = utils.ExtractQueryInt("start", req)
		}

		pageSize = utils.ExtractQueryInt("size", req)
		viewType := utils.ExtractVarString("viewType", req)

		discussions := businesslogic.GetFrontPage(user, viewType, pageSize, pageStart, h.userCache, h.discussionCache, db)

		return http.StatusOK, discussions, ""

	})
}

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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

var (
	frontPageCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "justthetalk_front_page_count",
		Help: "Count of front page requests",
	}, []string{"authenticated"})
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

		if user == nil {
			frontPageCount.WithLabelValues("anon").Inc()
		} else {
			frontPageCount.WithLabelValues("auth").Inc()
		}

		return http.StatusOK, discussions, ""

	})
}

func (h *FrontPageHandler) GetFrontPageSince(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var err error
		dateSince := time.Now()
		pageSize := 0

		dateParam := req.URL.Query().Get("dt")
		if len(dateParam) > 0 {
			dateSince, err = time.Parse(time.RFC3339, dateParam)
			if err != nil {
				panic(utils.ErrBadRequest)
			}
		}

		pageSize = utils.ExtractQueryInt("size", req)
		viewType := utils.ExtractVarString("viewType", req)

		discussions := businesslogic.GetFrontPageSince(user, viewType, pageSize, dateSince, h.userCache, h.discussionCache, db)

		if user == nil {
			frontPageCount.WithLabelValues("anon").Inc()
		} else {
			frontPageCount.WithLabelValues("auth").Inc()
		}

		return http.StatusOK, discussions, ""

	})
}

func (h *FrontPageHandler) GetFrontPageBefore(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var err error
		dateSince := time.Now()
		pageSize := 0

		dateParam := req.URL.Query().Get("dt")
		if len(dateParam) > 0 {
			dateSince, err = time.Parse(time.RFC3339, dateParam)
			if err != nil {
				panic(utils.ErrBadRequest)
			}
		}

		pageSize = utils.ExtractQueryInt("size", req)
		viewType := utils.ExtractVarString("viewType", req)

		discussions := businesslogic.GetFrontPageBefore(user, viewType, pageSize, dateSince, h.userCache, h.discussionCache, db)

		if user == nil {
			frontPageCount.WithLabelValues("anon").Inc()
		} else {
			frontPageCount.WithLabelValues("auth").Inc()
		}

		return http.StatusOK, discussions, ""

	})
}

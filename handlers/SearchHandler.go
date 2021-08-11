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
	"strconv"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SearchHandler struct {
	folderCache     *businesslogic.FolderCache
	discussionCache *businesslogic.DiscussionCache
}

func NewSearchHandler(folderCache *businesslogic.FolderCache, discussionCache *businesslogic.DiscussionCache) *SearchHandler {

	return &SearchHandler{
		folderCache:     folderCache,
		discussionCache: discussionCache,
	}

}

func (h *SearchHandler) SearchPosts(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var err error

		query := req.URL.Query().Get("q")
		sizeParam := req.URL.Query().Get("size")
		startParam := req.URL.Query().Get("start")

		size := 50
		start := 0

		if len(sizeParam) > 0 {
			if size, err = strconv.Atoi(sizeParam); err != nil {
				log.Errorf("%v", err)
				panic(utils.ErrBadRequest)
			}
		}

		if len(startParam) > 0 {
			if start, err = strconv.Atoi(startParam); err != nil {
				log.Errorf("%v", err)
				panic(utils.ErrBadRequest)
			}
		}

		if len(query) > 0 {
			results := businesslogic.SearchPosts(query, size, start, user, utils.ExtractIPAdress(req), h.folderCache, h.discussionCache, db, req.Context())
			return http.StatusOK, results, ""
		} else {
			panic(utils.ErrBadRequest)
		}

	})
}

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

func GetFrontPage(user *model.User, viewType string, pageSize int, pageStart int, userCache *UserCache, discussionCache *DiscussionCache, db *gorm.DB) []*model.FrontPageEntry {

	userId := 0
	isAdmin := 0

	if user != nil {
		userId = int(user.Id)
		if user.IsAdmin {
			isAdmin = 1
		}
	}

	if userId == 0 && (viewType == "subs" || viewType == "startedbyme") {
		panic(utils.ErrForbidden)
	}

	var discussions []*model.FrontPageEntry
	var result *gorm.DB
	switch viewType {
	case "latest":
		result = db.Raw("call get_frontpage_latest(?, ?, ?, ?)", userId, isAdmin, pageStart, pageSize).Scan(&discussions)
	case "mostactive":
		result = db.Raw("call get_frontpage_mostactive(?, ?, ?, ?)", userId, isAdmin, pageStart, pageSize).Scan(&discussions)
	case "subs":
		result = db.Raw("call get_frontpage_subscriptions(?, ?, ?, ?)", userId, isAdmin, pageStart, pageSize).Scan(&discussions)
	case "startedbyme":
		result = db.Raw("call get_frontpage_startedbyme(?, ?, ?, ?)", userId, isAdmin, pageStart, pageSize).Scan(&discussions)
	default:
		panic(utils.ErrBadRequest)
	}

	if result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	utils.FormatFrontPageEntries(discussions)

	return discussions

}

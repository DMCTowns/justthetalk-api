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

package model

import (
	"time"
)

const (
	DiscussionStatusOk             = 0
	DiscussionStatusDeletedByUser  = 1
	DiscussionStatusDeletedByAdmin = 2
	DiscussionStatusArchived       = 1024
	DiscussionStatusAged           = 1025
)

type BlockedDiscussionUser struct {
	ModelBase
	DiscussionId uint `json:"discussionId" gorm:"column:discussion_id"`
	UserId       uint `json:"userId" gorm:"column:user_id"`
	Status       bool `jdson:"status" gorm:"column:user_status"`
}

type Discussion struct {
	ModelBase
	FolderId          uint      `json:"folderId" gorm:"column:folder_id"`
	Title             string    `json:"title" gorm:"column:title"`
	TitleMarkup       string    `json:"titleMarkup" gorm:"-"`
	Header            string    `json:"header" gorm:"header"`
	HeaderMarkup      string    `json:"headerMarkup" gorm:"-"`
	LastPostId        uint      `json:"lastPostId" gorm:"column:last_post_id"`
	LastPostDate      time.Time `json:"lastPostDate" gorm:"column:last_post"`
	CreatedByUserId   uint      `json:"createdByUserId" gorm:"column:user_id"`
	CreatedByUsername string    `json:"createdByUsername" gorm:"column:username"`
	PostCount         int64     `json:"postCount" gorm:"column:post_count"`
	ZOrder            int       `json:"zOrder" gorm:"column:zorder"`
	Status            int       `json:"status" gorm:"column:status"`
	IsPremoderate     bool      `json:"isPremoderate" gorm:"column:premoderate"`
	IsDeleted         bool      `json:"isDeleted" gorm:"column:deleted"`
	IsLocked          bool      `json:"isLocked" gorm:"column:locked"`

	//BlockedUsers      map[uint]bool `json:"blockedUsers" gorm:"-"`
	Url          string `json:"url" gorm:"-"`
	IsBlocked    bool   `json:"isBlocked" gorm:"-"`
	IsSubscribed bool   `json:"isSubscribed" gorm:"-"`
}

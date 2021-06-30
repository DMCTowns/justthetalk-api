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

import "time"

const (
	PostStatusOK               = 0
	PostStatusSuspendedByAdmin = 1
	PostStatusDeletedByAdmin   = 2
	PostStatusPostedByAdmin    = 3
	PostStatusWatch            = 4
	PostStatusMaxDisplay       = 255
	PostStatusDeletedByUser    = 256
	PostStatusInvisible        = 257
)

type Post struct {
	ModelBase
	DiscussionId          uint      `json:"discussionId" gorm:"column:discussion_id"`
	DiscussionStatus      uint      `json:"discussionStatus" gorm:"column:discussion_status"`
	CreatedByUserId       uint      `json:"createdByUserId" gorm:"column:user_id"`
	CreatedByUsername     string    `json:"createdByUsername" gorm:"column:username"`
	CreatedByEnabled      bool      `json:"createdByEnabled" gorm:"column:user_enabled"`
	CreatedByLocked       bool      `json:"createdByLocked" gorm:"column:user_locked"`
	CreatedByExpired      bool      `json:"createdByExpired" gorm:"column:user_expired"`
	CreatedByWatch        bool      `json:"createdByWatch" gorm:"column:user_watch"`
	CreatedByPremoderate  bool      `json:"createdByPremoderate" gorm:"column:user_premod"`
	Deleted               bool      `json:"deleted" gorm:"column:deleted"`
	Status                int       `json:"status" gorm:"column:status"`
	Text                  string    `json:"text" gorm:"column:text"`
	Markdown              bool      `json:"markdown" gorm:"column:markdown"`
	Markup                string    `json:"markup" gorm:"-"`
	LastEditDate          time.Time `json:"lastEditDate" gorm:"column:last_edit_date"`
	PostNum               int64     `json:"postNum" gorm:"column:post_num"`
	ModerationScore       float64   `json:"moderationScore" gorm:"column:moderation_score"`
	ModerationResult      int       `json:"moderationResult" gorm:"column:moderation_result"`
	PostAsAdmin           bool      `json:"postAsAdmin,omitempty" gorm:"-"`
	SubscribeToDiscussion bool      `json:"subscribeToDiscussion,omitempty" gorm:"-"`
	Url                   string    `json:"url" gorm:"-"`
}

type IndexablePost struct {
	Id               uint      `json:"id" gorm:"column:id;primaryKey"`
	CreatedDate      time.Time `json:"date" gorm:"column:created_date"`
	FolderName       string    `json:"folder" gorm:"column:folder_name"`
	DiscussionTitle  string    `json:"thread" gorm:"column:discussion_title"`
	DiscussionHeader string    `json:"threadHeader" gorm:"column:discussion_header"`
	Text             string    `json:"text" gorm:"column:text"`
	Username         string    `json:"username" gorm:"column:username"`
}
type PostReport struct {
	ModelBase
	PostId         uint    `json:"postId" gorm:"column:post_id"`
	ReporterUserId uint    `json:"userId" gorm:"column:user_id"`
	ReporterName   string  `json:"name" gorm:"column:name"`
	ReporterEmail  string  `json:"email" gorm:"column:email"`
	Body           string  `json:"body" gorm:"column:comment"`
	IPAddress      string  `json:"ipaddress" gorm:"column:ipaddress"`
	Score          float64 `json:"score" gorm:"column:score"`
}

type ModeratorComment struct {
	ModelBase
	Name   string `json:"name" gorm:"column:username"`
	Body   string `json:"body" gorm:"column:comment"`
	PostId uint   `json:"postId" gorm:"column:post_id"`
	UserId uint   `json:"userId" gorm:"column:user_id"`
	Vote   int    `json:"vote" gorm:"column:result"`
}

type ModeratedPostRecord struct {
	Post           Post
	Report         PostReport
	Comment        ModeratorComment
	FrontPageEntry FrontPageEntry
}

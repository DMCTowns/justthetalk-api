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

type FrontPageEntry struct {
	DiscussionId      uint      `json:"discussionId" gorm:"column:discussion_id"`
	DiscussionTitle   string    `json:"discussionTitle" gorm:"column:discussion_name"`
	FolderId          uint      `json:"folderId" gorm:"column:folder_id"`
	FolderKey         string    `json:"folderKey" gorm:"column:folder_key"`
	FolderTitle       string    `json:"folderTitle" gorm:"column:folder_name"`
	LastPostId        uint      `json:"lastPostId" gorm:"column:last_post_id"`
	LastPostDate      time.Time `json:"lastPostDate" gorm:"column:last_post"`
	PostCount         int64     `json:"postCount" gorm:"column:post_count"`
	IsAdmin           *bool     `json:"-" gorm:"column:admin_only"`
	LastPostReadCount int64     `json:"lastPostReadCount" gorm:"last_post_read_count"`
	LastPostReadDate  time.Time `json:"lastPostReadDate" gorm:"last_post_read_date"`
	LastPostReadId    uint      `json:"lastPostReadId" gorm:"last_post_read_id"`
	Url               string    `json:"url" gorm:"-"`
}

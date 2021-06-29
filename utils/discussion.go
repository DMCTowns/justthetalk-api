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

package utils

import (
	"fmt"
	"justthetalk/model"

	"github.com/gosimple/slug"
)

func UrlForFrontPageEntry(entry *model.FrontPageEntry) string {
	slugText := slug.Make(entry.DiscussionTitle)
	return fmt.Sprintf("/%s/%d/%s", entry.FolderKey, entry.DiscussionId, slugText)
}

func UrlForDiscussion(folder *model.Folder, discussion *model.Discussion) string {
	slugText := slug.Make(discussion.Title)
	return fmt.Sprintf("/%s/%d/%s", folder.Key, discussion.Id, slugText)
}

func UrlForPost(folder *model.Folder, discussion *model.Discussion, post *model.Post) string {
	slugText := slug.Make(discussion.Title)
	return fmt.Sprintf("/%s/%d/%s/%d", folder.Key, discussion.Id, slugText, post.PostNum)
}

func FormatFrontPageEntry(entry *model.FrontPageEntry) {
	//entry.DiscussionTitle = html.EscapeString(entry.DiscussionTitle)
	entry.Url = UrlForFrontPageEntry(entry)
}

func FormatFrontPageEntries(entries []*model.FrontPageEntry) {
	for _, entry := range entries {
		FormatFrontPageEntry(entry)
	}
}

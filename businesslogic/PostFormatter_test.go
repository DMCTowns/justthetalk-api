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
	"justthetalk/utils"
	"testing"
)

func TestFormatting1(t *testing.T) {

	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)
	discussion := discussionCache.UnsafeGet(47)

	formatter := utils.NewPostFormatter()

	result := formatter.ApplyPostFormatting(discussion.Header, discussion)
	t.Log(result)

	expected := `<div><p>See post <a href="/userspace/47/thread-to-practise-formatting/1">#1</a>...</p><p></p><p>	There&#39;s a new version of the site to test on <a href="https://beta.justthetalk.com" rel='nofollow'>https://beta.justthetalk.com</a></p></div>`

	if len(result) != len(expected) {
		t.Errorf("Length mismatch %d vs %d", len(result), len(expected))
	}

	for i := 0; i < len(result); i++ {
		if result[:i] != expected[:i] {
			t.Errorf("Result mismatch:\n%s\n%s", result[:i], expected[:i])
			break
		}
	}

}

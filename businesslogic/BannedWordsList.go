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
	"justthetalk/connections"
	"regexp"
	"time"

	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

type BannedWordsEntry struct {
	Id      int            `gorm:"column:id"`
	Version int            `gorm:"column:version"`
	Pattern string         `gorm:"column:word"`
	re      *regexp.Regexp `gorm:"-"`
}

type BannedWordsList struct {
	bannedWords []*BannedWordsEntry
}

func NewBannedWordsList() *BannedWordsList {

	wordList := BannedWordsList{
		bannedWords: make([]*BannedWordsEntry, 0),
	}

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		if result := db.Raw("select * from banned_word").Scan(&wordList.bannedWords); result.Error != nil {
			panic(result.Error)
		}
	})

	for _, entry := range wordList.bannedWords {
		entry.re = regexp.MustCompile(entry.Pattern)
	}

	return &wordList

}

func (list *BannedWordsList) CheckForBannedWords(text string) bool {

	var found bool

	for _, entry := range list.bannedWords {
		if entry.re.MatchString(text) {
			log.Warnf("%s=%s\n", text, entry.Pattern)
			found = true
			break
		}
	}

	return found

}

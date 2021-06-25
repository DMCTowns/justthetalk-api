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

type ModelBase struct {
	Id              uint      `json:"id" gorm:"column:id;primaryKey"`
	Version         uint      `json:"version" gorm:"column:version"`
	CreatedDate     time.Time `json:"createdDate" gorm:"column:created_date"`
	LastUpdatedDate time.Time `json:"lastUpdatedDate" gorm:"column:last_updated"`
}

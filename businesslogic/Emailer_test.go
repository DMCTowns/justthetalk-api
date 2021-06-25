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
	"testing"
	"time"

	"gorm.io/gorm"

	"justthetalk/connections"
	"justthetalk/model"
)

func TestSendPasswordResetEmail(t *testing.T) {

	userCache := NewUserCache()

	requestId := 1274
	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		var request model.PasswordResetRequest
		if result := db.Table("password_reset").Where("id = ?", requestId).Take(&request); result.Error != nil {
			t.Error("failed to get request")
		}

		user := userCache.Get(request.UserId)

		SendEmailToUser(user, &request, PasswordResetRequestTemplate)

	})

}

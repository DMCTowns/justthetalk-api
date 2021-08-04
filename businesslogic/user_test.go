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
	"errors"
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestValidUserLogin(t *testing.T) {

	userCache := NewUserCache()

	credentials := model.LoginCredentials{
		Username: "testuser1",
		Password: "1234567890",
	}

	connections.WithDatabase(10*time.Second, func(db *gorm.DB) {

		var count1 int64
		var count2 int64

		db.Table("login_history").Count(&count1)

		user := ValidateUserLogin(credentials, "8.8.8.8", db, userCache)
		if user == nil || user.Id != 5540 {
			t.Error("User not found")
		}

		db.Table("login_history").Count(&count2)
		if count2-count1 != 1 {
			t.Error("No login history")
		}

	})

}

func TestInvalidUserLogin(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrUnauthorised):
				t.Log("Got unauthorised")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()

	credentials := model.LoginCredentials{
		Username: "testuser1",
		Password: "xxxxxx",
	}

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		ValidateUserLogin(credentials, "8.8.8.8", db, userCache)
	})

	t.Error("Expecting a panic here")

}

func TestCreateUser(t *testing.T) {
	t.Fail()
}

func TestConfirmUser(t *testing.T) {

	key := "50926866-aa8b-4751-b173-ae57b3d9eb7f"
	userCache := NewUserCache()
	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {
		ValidateSignupConfirmationKey(key, "8.8.8.8", userCache, db)
	})

}

func TestConfirmUserExpired(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrExpired):
				t.Log("Got expired")
			default:
				t.Error(err)
			}
		}
	}()

	key := "58ffca03-3f5c-4e64-bfbe-ba22357b68a4"
	userCache := NewUserCache()
	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {
		ValidateSignupConfirmationKey(key, "8.8.8.8", userCache, db)
		t.Error("Should have failed")
	})

}

func TestConfirmUserAlreadyUsed(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrBadRequest):
				t.Log("Got bad request")
			default:
				t.Error(err)
			}
		}
	}()

	key := "50926866-aa8b-4751-b173-ae57b3d9eb7f"
	userCache := NewUserCache()
	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {
		ValidateSignupConfirmationKey(key, "8.8.8.8", userCache, db)
	})

}

func TestCreateUserFailsWithGarbageParams(t *testing.T) {
	t.Fail()
}

func TestCreateReport(t *testing.T) {
	t.Fail()
}

func TestCreateReportFailsIfPostAlreadyReported(t *testing.T) {
	t.Fail()
}

func TestCreateReportCreatesUserHistory(t *testing.T) {
	t.Fail()
}

func TestGetOtherUser(t *testing.T) {

	userCache := NewUserCache()
	userId := uint(50)

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		user := GetOtherUser(userId, db, userCache)
		if user == nil {
			t.Error("Failed to get user")
		} else if !(user.UserId == 50 && user.Username == "johnnythesailor") {
			t.Error("Details invalid")
		}
	})

}

func TestSetUnsetFolderSubscription(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	userId := uint(5540)
	user := userCache.Get(userId)

	folderId := uint(1)
	folder := folderCache.Get(folderId, user)

	discussionId := uint(1)
	discussion := discussionCache.Get(discussionId, user)

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {

		SetDiscussionSubscriptionStatus(discussion, user, db, userCache)
		subscribed := GetDiscussionSubscriptionStatus(discussion, user, db)
		if !subscribed {
			t.Error("Subscription not set")
		}

		UnsetDiscussionSubscriptionStatus(discussion, user, db, userCache)
		subscribed = GetDiscussionSubscriptionStatus(discussion, user, db)
		if subscribed {
			t.Error("Subscription set")
		}

		SetFolderSubscriptionStatus(folder, user, db, userCache)
		subscribed = GetFolderSubscriptionStatus(folder, user, db)
		if !subscribed {
			t.Error("Subscription not set")
		}

		UnsetFolderSubscriptionStatus(folder, user, db, userCache)
		subscribed = GetFolderSubscriptionStatus(folder, user, db)
		if !subscribed {
			t.Error("Subscription not unset")
		}

	})

}
func TestSetUnsetDiscussionSubscription(t *testing.T) {

	folderCache := NewFolderCache()

	userCache := NewUserCache()

	discussionCache := NewDiscussionCache(folderCache)

	userId := uint(5540)
	user := userCache.Get(userId)

	discussionId := uint(2801)
	discussion := discussionCache.Get(discussionId, user)

	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {

		SetDiscussionSubscriptionStatus(discussion, user, db, userCache)
		UnsetDiscussionSubscriptionStatus(discussion, user, db, userCache)

	})

}
func TestGetUserSubscriptionStatus(t *testing.T) {

	userCache := NewUserCache()
	folderCache := NewFolderCache()
	discussionCache := NewDiscussionCache(folderCache)

	user := userCache.Get(50)

	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {
		discussion := discussionCache.Get(2494, user)
		subscribed := GetDiscussionSubscriptionStatus(discussion, user, db)
		if !subscribed {
			t.Error("Not subscribed")
		}
	})

	connections.WithDatabase(60*time.Second, func(db *gorm.DB) {
		discussion := discussionCache.Get(2495, user)
		subscribed := GetDiscussionSubscriptionStatus(discussion, user, db)
		if subscribed {
			t.Error("Subscribed")
		}
	})

}

func TestGetFolderSubscriptions(t *testing.T) {

	userCache := NewUserCache()

	userId := uint(50)
	user := userCache.Get(userId)
	//	var subs []*model.FolderSubscription

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		results := GetFolderSubscriptions(user, db)
		if len(results) == 0 {
			t.Error("No subs found")
		}
	})

}

func TestGetFolderSubscriptionExceptions(t *testing.T) {

	userCache := NewUserCache()

	userId := uint(50)
	user := userCache.Get(userId)

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		results := GetFolderSubscriptionExcepions(user, db)
		if len(results) == 0 {
			t.Error("No subs exceptions found")
		}
	})

}

func TestUpdateFolderSubscriptions(t *testing.T) {

	folderCache := NewFolderCache()
	userCache := NewUserCache()

	userId := uint(5540)
	user := userCache.Get(userId)

	subscriptions := []uint{1, 2, 3}

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		result := UpdateFolderSubscriptions(subscriptions, user, db, userCache, folderCache)
		if len(result) != 3 {
			t.Error("Not all subscriptions created")
		}

		subscriptions = []uint{}
		result = UpdateFolderSubscriptions(subscriptions, user, db, userCache, folderCache)
		if len(result) != 0 {
			t.Error("Not all subscriptions deleted")
		}
	})

}

func TestUpdateIgnore(t *testing.T) {

	userCache := NewUserCache()
	userId := uint(50)
	user := userCache.Get(userId)
	otherUserId := uint(5540)

	connections.WithDatabase(1*time.Second, func(db *gorm.DB) {
		UpdateIgnore(user, otherUserId, 0, db, userCache)
		if _, exists := user.IgnoredUsers[otherUserId]; exists {
			t.Error("Should not have user in ignore state")
		}

		UpdateIgnore(user, otherUserId, 1, db, userCache)
		if _, exists := user.IgnoredUsers[otherUserId]; !exists {
			t.Error("Should have user in ignore state")
		}

		UpdateIgnore(user, otherUserId, 0, db, userCache)
		if _, exists := user.IgnoredUsers[otherUserId]; exists {
			t.Error("Should not have user in ignore state after add + remove")
		}
	})

}

func TestForgotPasswordWithValidEmail(t *testing.T) {

	credentials := model.LoginCredentials{
		Email: "john@johndudmesh.com",
	}

	userCache := NewUserCache()

	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		var count1 int64
		var count2 int64

		db.Table("password_reset").Count(&count1)

		request := ForgotPassword(&credentials, "8.8.8.8", userCache, db)
		if request == nil {
			t.Error("user not found")
		}

		db.Table("password_reset").Count(&count2)
		if count2-count1 != 1 {
			t.Error("No reset record")
		}

	})

}

func TestForgotPasswordWithInvalidEmail(t *testing.T) {

	credentials := model.LoginCredentials{
		Email: "john@nobody.com",
	}

	userCache := NewUserCache()

	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		var count1 int64
		var count2 int64

		db.Table("password_reset").Count(&count1)

		request := ForgotPassword(&credentials, "8.8.8.8", userCache, db)
		if request != nil {
			t.Error("Unexpected user")
		}

		db.Table("password_reset").Count(&count2)
		if count2-count1 != 0 {
			t.Error("No reset record")
		}

	})

}

func TestValidatePasswordResetKeySuccess(t *testing.T) {

	credentials := model.LoginCredentials{
		Email: "john@johndudmesh.com",
	}

	userCache := NewUserCache()

	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		request := ForgotPassword(&credentials, "8.8.8.8", userCache, db)
		if request == nil {
			t.Error("user not found")
			return
		}

		user, _ := ValidatePasswordResetKey(request.ResetKey, userCache, db)
		if user == nil || user.Id != request.UserId {
			t.Error("User not returned or invalid")
		}

	})
}

func TestValidatePasswordResetKeyFail(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrBadRequest):
				t.Log("Got bad request")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()

	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		user, _ := ValidatePasswordResetKey("11111111-1111-1111-1111-111111111111", userCache, db)
		if user != nil {
			t.Error("Unexpected user")
		}

	})

}

func TestValidatePasswordResetKeyExpired(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrExpired):
				t.Log("Got expired")
			default:
				t.Error(err)
			}
		}
	}()

	userCache := NewUserCache()

	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		user, _ := ValidatePasswordResetKey("696f4adf-39ab-4e9a-905d-8c5412962c09", userCache, db)
		if user != nil {
			t.Error("Unexpected user")
		}

	})

}

func TestUpdateUserPasswordSuccess(t *testing.T) {

	updateData := model.UserOptionsUpdateData{
		OldPassword: "password",
		NewPassword: "password",
	}

	userCache := NewUserCache()
	user := userCache.Get(1777)
	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {
		UpdatePassword(user, &updateData, userCache, db)
	})

}

func TestUpdateUserPasswordFailure(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			switch {
			case errors.Is(err, utils.ErrUnauthorised):
				t.Log("Got unauthorised")
			default:
				t.Error(err)
			}
		}
	}()

	updateData := model.UserOptionsUpdateData{
		OldPassword: "wrong_password",
		NewPassword: "password",
	}

	userCache := NewUserCache()
	user := userCache.Get(1777)
	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {
		UpdatePassword(user, &updateData, userCache, db)
	})

}

func TestUpdateUserPasswordWithKeySuccess(t *testing.T) {

	credentials := model.LoginCredentials{
		Email: "john@johndudmesh.com",
	}

	userCache := NewUserCache()
	user := userCache.Get(1777)
	connections.WithDatabase(30*time.Second, func(db *gorm.DB) {

		request := ForgotPassword(&credentials, "8.8.8.8", userCache, db)
		if request == nil {
			t.Error("user not found")
			return
		}
		t.Logf("Reset key: %s", request.ResetKey)

		updateData := model.UserOptionsUpdateData{
			ResetKey:    request.ResetKey,
			NewPassword: "password",
		}

		UpdatePassword(user, &updateData, userCache, db)

		var deletedRequest model.PasswordResetRequest
		if result := db.Table("password_reset").Where("reset_key = ?", request.ResetKey).Take(&deletedRequest); result.Error == nil || !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			t.Error("Reset key not deleted")
		}
		if deletedRequest.Id > 0 {
			t.Error("Reset key not deleted")
		}

	})

}

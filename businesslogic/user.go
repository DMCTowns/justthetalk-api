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
	"gorm.io/gorm"

	"crypto/sha256"
	"errors"
	"fmt"
	"html"
	"justthetalk/model"
	"justthetalk/utils"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func CreateLoginHistory(status string, user *model.User, ipAddress string, db *gorm.DB) {

	history := model.LoginHistory{
		CreatedDate: time.Now(),
		UserId:      user.Id,
		IPAddress:   ipAddress,
		Status:      status,
	}

	if result := db.Table("login_history").Create(&history); result.Error != nil {
		log.Errorf("%v", result.Error)
		panic(utils.ErrInternalError)
	}

}

func ValidateUserLogin(credentials model.LoginCredentials, ipAddress string, db *gorm.DB, userCache *UserCache) *model.User {

	username := html.EscapeString(credentials.Username)
	passwordHashBytes := sha256.Sum256([]byte(credentials.Password))
	passwordHash := fmt.Sprintf("%x", passwordHashBytes)

	var userLookup model.User
	if result := db.Raw("call find_user(?, ?)", username, passwordHash).Take(&userLookup); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Errorf("Failed login for user: %s", username)
			utils.PanicWithWrapper(errors.New("Unknown username or incorrect password"), utils.ErrUnauthorised)
		} else {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	} else if userLookup.ModelBase.Id == 0 {
		log.Errorf("Failed login for user: %s", username)
		panic(utils.ErrUnauthorised)
	}

	user := userCache.Get(userLookup.Id)

	if user.AccountExpired || !user.Enabled {
		utils.PanicWithWrapper(errors.New("This account has been deleted"), utils.ErrUnauthorised)
	}

	CreateLoginHistory("login", user, ipAddress, db)

	return user

}

func GetDiscussionSubscriptionStatus(discussion *model.Discussion, user *model.User, db *gorm.DB) bool {

	var isSubscribed int64
	if result := db.Raw("call get_discussion_subscription_status(?, ?)", user.Id, discussion.Id).First(&isSubscribed); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	if isSubscribed == 0 {
		return false
	} else {
		return true
	}

}

func GetFolderSubscriptionStatus(folder *model.Folder, user *model.User, db *gorm.DB) bool {

	var isSubscribed int64
	if result := db.Raw("call get_folder_subscription_status(?, ?)", user.Id, folder.Id).First(&isSubscribed); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	if isSubscribed == 0 {
		return false
	} else {
		return true
	}

}

func MarkFolderSubscriptionsRead(subsList []uint, user *model.User, db *gorm.DB, userCache *UserCache) []*model.UserFolderSubscription {

	err := db.Transaction(func(tx *gorm.DB) error {

		var err error
		for _, subsId := range subsList {
			err = tx.Exec("call mark_folder_subscription_read(?, ?)", user.Id, subsId).Error
			if err != nil {
				break
			}
		}
		return err
	})

	if err != nil {
		utils.PanicWithWrapper(err, utils.ErrInternalError)
	}

	var entries []*model.UserFolderSubscription
	if result := db.Raw("call get_folder_subscriptions(?)", user.Id).Scan(&entries); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return entries

}

func MarkDiscussionSubscriptionsRead(subsList []uint, user *model.User, db *gorm.DB, userCache *UserCache) []*model.FrontPageEntry {

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		for _, subsId := range subsList {
			err = tx.Exec("call mark_discussion_read(?, ?)", user.Id, subsId).Error
			if err != nil {
				break
			}
		}
		return err
	})

	if err != nil {
		utils.PanicWithWrapper(err, utils.ErrInternalError)
	}

	var entries []*model.FrontPageEntry
	if result := db.Raw("call get_discussion_subscriptions(?)", user.Id).Scan(&entries); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return entries

}

func DeleteFolderSubscriptions(subsList []uint, user *model.User, db *gorm.DB, userCache *UserCache) []*model.UserFolderSubscription {

	err := db.Transaction(func(tx *gorm.DB) error {

		var err error
		for _, subsId := range subsList {
			err = tx.Exec("call delete_folder_subscription(?, ?)", user.Id, subsId).Error
			if err != nil {
				break
			}
		}
		return err
	})

	if err != nil {
		utils.PanicWithWrapper(err, utils.ErrInternalError)
	}

	entries := make([]*model.UserFolderSubscription, 0)
	if result := db.Raw("call get_folder_subscriptions(?)", user.Id).Scan(&entries); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return entries

}

func DeleteDiscussionSubscriptions(subsList []uint, user *model.User, db *gorm.DB, userCache *UserCache) []*model.FrontPageEntry {

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		for _, subsId := range subsList {
			err = tx.Exec("call delete_discussion_subscription(?, ?)", user.Id, subsId).Error
			if err != nil {
				break
			}
		}
		return err
	})

	if err != nil {
		utils.PanicWithWrapper(err, utils.ErrInternalError)
	}

	subscriptions := make([]*model.FrontPageEntry, 0)
	if result := db.Raw("call get_discussion_subscriptions(?)", user.Id).Scan(&subscriptions); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	for _, sub := range subscriptions {
		sub.Url = utils.UrlForFrontPageEntry(sub)
	}

	return subscriptions

}

func updateFolderSubscription(userData *model.UserSidebandData, folderId uint, subscriptionState int, db *gorm.DB) {

	var folderSubs []*model.UserFolderSubscription
	if result := db.Raw("call update_user_folder_subscription(?, ?, ?)", userData.UserId, folderId, subscriptionState).Scan(&folderSubs); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	userData.FolderSubscriptions = make(map[uint]*model.UserFolderSubscription)
	for _, sub := range folderSubs {
		userData.FolderSubscriptions[sub.FolderId] = sub
	}

	if subscriptionState == 0 {
		userData.FolderSubscriptionExceptions = make(map[uint]*model.UserFolderSubscriptionException)
	}

}

func SetDiscussionSubscriptionStatus(discussion *model.Discussion, user *model.User, db *gorm.DB, userCache *UserCache) {

	if result := db.Exec("call update_user_discussion_subscription(?, ?, ?)", user.Id, discussion.Id, 1); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

}

func UnsetDiscussionSubscriptionStatus(discussion *model.Discussion, user *model.User, db *gorm.DB, userCache *UserCache) {

	if result := db.Exec("call update_user_discussion_subscription(?, ?, ?)", user.Id, discussion.Id, 0); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

}

func SetFolderSubscriptionStatus(folder *model.Folder, user *model.User, db *gorm.DB, userCache *UserCache) {

	if result := db.Exec("call update_user_folder_subscription(?, ?, ?)", user.Id, folder.Id, 1); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

}

func UnsetFolderSubscriptionStatus(folder *model.Folder, user *model.User, db *gorm.DB, userCache *UserCache) {

	if result := db.Exec("call update_user_folder_subscription(?, ?, ?)", user.Id, folder.Id, 0); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

}

func GetDiscussionSubscriptions(user *model.User, db *gorm.DB) []*model.FrontPageEntry {

	subscriptions := make([]*model.FrontPageEntry, 0)
	if result := db.Raw("call get_discussion_subscriptions(?)", user.Id).Scan(&subscriptions); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	for _, sub := range subscriptions {
		sub.Url = utils.UrlForFrontPageEntry(sub)
	}

	return subscriptions
}

func GetFolderSubscriptions(user *model.User, db *gorm.DB) []*model.UserFolderSubscription {

	subscriptions := make([]*model.UserFolderSubscription, 0)
	if result := db.Raw("call get_folder_subscriptions(?)", user.Id).Scan(&subscriptions); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return subscriptions

}

func GetFolderSubscriptionExcepions(user *model.User, db *gorm.DB) []*model.UserFolderSubscriptionException {

	exceptions := make([]*model.UserFolderSubscriptionException, 0)
	if result := db.Raw("call get_folder_subscription_exceptions(?)", user.Id).Scan(&exceptions); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return exceptions

}

func UpdateFolderSubscriptions(subsList []uint, user *model.User, db *gorm.DB, userCache *UserCache, folderCache *FolderCache) []*model.UserFolderSubscription {

	subscriptions := make(map[uint]bool)

	for _, folder := range folderCache.Entries() {
		subscriptions[folder.Id] = false
	}

	for _, folderId := range subsList {
		subscriptions[folderId] = true
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		for folderId, subscribed := range subscriptions {
			subscriptionState := 0
			if subscribed {
				subscriptionState = 1
			}
			if result := db.Exec("call update_user_folder_subscription(?, ?, ?)", user.Id, folderId, subscriptionState); result.Error != nil {
				err = result.Error
				break
			}

		}
		return err
	})

	if err != nil {
		utils.PanicWithWrapper(err, utils.ErrInternalError)
	}

	results := GetFolderSubscriptions(user, db)

	return results

}

func GetOtherUser(userId uint, db *gorm.DB, userCache *UserCache) *model.OtherUser {
	user := userCache.Get(userId)
	return &model.OtherUser{
		UserId:      user.Id,
		Username:    user.Username,
		Bio:         user.Bio,
		CreatedDate: user.CreatedDate,
	}
}

func UpdateIgnore(user *model.User, ignoreUserId uint, ignoreState int, db *gorm.DB, userCache *UserCache) {

	var ignored []*model.IgnoredUser
	if result := db.Raw("call update_user_ignore(?, ?, ?)", user.Id, ignoreUserId, ignoreState).Scan(&ignored); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	user.IgnoredUsers = make(map[uint]*model.IgnoredUser)
	for _, item := range ignored {
		user.IgnoredUsers[item.IgnoredUserId] = item
	}

	userCache.Put(user)

}

func CreateUser(credentials *model.LoginCredentials, ipAddress string, db *gorm.DB) *model.User {

	username := html.EscapeString(credentials.Username)

	var countOfExisting int64
	if result := db.Table("user").Where("username = ? or email = ?", username, credentials.Email).Count(&countOfExisting); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	} else if countOfExisting > 0 {
		utils.PanicWithWrapper(utils.ErrBadRequest, errors.New("This username is already taken or e-mail address has already been used"))
	}

	passwordHashBytes := sha256.Sum256([]byte(credentials.Password))
	passwordHash := fmt.Sprintf("%x", passwordHashBytes)

	var user model.User
	if result := db.Raw("call create_user(?, ?, ?)", credentials.Email, username, passwordHash).Take(&user); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	CreateLoginHistory("new", &user, ipAddress, db)

	CreateNewSignupConfirmation(&user, db)

	return &user

}

func CreateNewSignupConfirmation(user *model.User, db *gorm.DB) {

	var confirmation model.SignupConfirmation
	if result := db.Raw("call create_signup_confirmation(?, ?)", user.Id, uuid.NewString()).Take(&confirmation); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	SendEmailToUser(user, confirmation, NewSignupTemplate)

}

func ForgotPassword(credentials *model.LoginCredentials, ipAddress string, userCache *UserCache, db *gorm.DB) *model.PasswordResetRequest {

	var foundUser model.User
	if result := db.Raw("call find_user_by_email(?)", credentials.Email).Take(&foundUser); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	user := userCache.Get(foundUser.Id)

	var request model.PasswordResetRequest
	if result := db.Raw("call create_password_reset_request(?, ?, ?)", user.Id, ipAddress, uuid.NewString()).Take(&request); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	SendEmailToUser(user, request, PasswordResetRequestTemplate)

	return &request

}

func ValidatePasswordResetKey(key string, userCache *UserCache, db *gorm.DB) (*model.User, *model.PasswordResetRequest) {

	if _, err := uuid.Parse(key); err != nil {
		panic(utils.ErrBadRequest)
	}

	var request model.PasswordResetRequest
	if result := db.Raw("call find_password_reset_request(?)", key).Take(&request); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			panic(utils.ErrBadRequest)
		}
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	if request.CreatedDate.Add(time.Hour).Before(time.Now()) {
		utils.PanicWithWrapper(errors.New("Reset token not valid"), utils.ErrExpired)
	}

	if result := db.Table("user").Where("id = ?", request.UserId).Update("password_expired", 1); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	user := userCache.Get(request.UserId)
	user.PasswordExpired = true

	return user, &request

}

func UpdatePassword(user *model.User, updateData *model.UserOptionsUpdateData, userCache *UserCache, db *gorm.DB) {

	if len(updateData.NewPassword) < 8 {
		utils.PanicWithWrapper(errors.New("Passwords must be at least 8 characters long"), utils.ErrBadRequest)
	}

	err := db.Transaction(func(tx *gorm.DB) error {

		var err error

		if len(updateData.OldPassword) > 0 {

			passwordHashBytes := sha256.Sum256([]byte(updateData.OldPassword))
			passwordHash := fmt.Sprintf("%x", passwordHashBytes)

			if result := tx.Raw("call find_user(?, ?)", user.Username, passwordHash).Take(user); result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					err = utils.ErrUnauthorised
				} else {
					err = result.Error
				}
			} else if user.ModelBase.Id == 0 {
				err = utils.ErrUnauthorised
			}

		} else if len(updateData.ResetKey) > 0 {
			matchingUser, resetRequest := ValidatePasswordResetKey(updateData.ResetKey, userCache, tx)
			if matchingUser.Id != user.Id {
				err = utils.ErrBadRequest
			}
			if result := tx.Table("password_reset").Delete(&resetRequest); result.Error != nil {
				err = result.Error
			}
		} else {
			err = utils.ErrBadRequest
		}

		if err == nil {

			passwordHashBytes := sha256.Sum256([]byte(updateData.NewPassword))
			passwordHash := fmt.Sprintf("%x", passwordHashBytes)
			if result := tx.Raw("call update_user_password(?, ?)", user.Id, passwordHash).Scan(user); result.Error != nil {
				utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
			}

			userCache.Flush(user)

		}

		return err

	})

	if err != nil {
		panic(err)
	}

}

func ValidateSignupConfirmationKey(key string, ipAddress string, userCache *UserCache, db *gorm.DB) *model.User {

	if _, err := uuid.Parse(key); err != nil {
		panic(utils.ErrBadRequest)
	}

	var request model.SignupConfirmation

	if result := db.Raw("call find_signup_confirmation_request(?)", key).Take(&request); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			panic(utils.ErrBadRequest)
		}
		panic(result.Error)
	}

	if request.CreatedDate.Add(72 * time.Hour).Before(time.Now()) {
		panic(utils.ErrExpired)
	}

	var user model.User
	if result := db.Raw("call accept_signup_confirmation_request(?, ?)", request.Id, ipAddress).Take(&request); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			panic(utils.ErrBadRequest)
		}
		panic(result.Error)
	}

	updatedUser := userCache.Reload(user.Id)

	return updatedUser

}

func UpdateAutoSubscribe(user *model.User, subscribeState int, userCache *UserCache, db *gorm.DB) *model.User {

	var u model.User
	if result := db.Raw("call update_user_autosubscribe(?, ?)", user.Id, subscribeState).Scan(&u); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	updatedUser := userCache.Reload(user.Id)

	return updatedUser

}

func UpdateSortFoldersByActivity(user *model.User, sortState int, userCache *UserCache, db *gorm.DB) *model.User {

	var u model.User
	if result := db.Raw("call update_user_foldersort(?, ?)", user.Id, sortState).Scan(&u); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	updatedUser := userCache.Reload(user.Id)

	return updatedUser

}

func UpdateSubscriptionFetchOrder(user *model.User, fetchOrder int, userCache *UserCache, db *gorm.DB) *model.User {

	var u model.User
	if result := db.Raw("call update_user_subsfetchorder(?, ?)", user.Id, fetchOrder).Scan(&u); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	updatedUser := userCache.Reload(user.Id)

	return updatedUser

}

func UpdateBio(user *model.User, bio string, userCache *UserCache, db *gorm.DB) *model.User {

	var u model.User
	if result := db.Raw("call update_user_bio(?, ?)", user.Id, bio).Scan(&u); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	updatedUser := userCache.Reload(user.Id)

	return updatedUser

}

func DeleteDiscussionBookmark(user *model.User, discussion *model.Discussion, userCache *UserCache, db *gorm.DB) {

	if result := db.Exec("call delete_user_bookmark(?, ?)", user.Id, discussion.Id); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	userCache.FlushBookmark(user, discussion)

}

func CreateReport(reportData *model.PostReport, db *gorm.DB) {

	if result := db.Exec("call create_report(?, ?, ?, ?, ?, ?)", reportData.PostId, reportData.ReporterUserId, reportData.ReporterName, reportData.ReporterEmail, reportData.Body, reportData.IPAddress); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	SendEmail(reportData.ReporterEmail, reportData, ReportSubmittedTemplate)

}

func UpdateViewType(user *model.User, viewType string, userCache *UserCache, db *gorm.DB) *model.User {

	var updatedUser model.User
	if result := db.Raw("call update_user_viewtype(?, ?)", user.Id, viewType).Scan(&updatedUser); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	user.ViewType = updatedUser.ViewType
	userCache.Put(user)

	return user

}

func GetIgnoredUsers(user *model.User, db *gorm.DB) []*model.IgnoredUser {

	var ignoredUserList []*model.IgnoredUser
	if result := db.Raw("call get_user_ignored_users(?)", user.Id).Scan(&ignoredUserList); result.Error != nil {
		utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
	}

	return ignoredUserList

}

func CheckSubscriptions(user *model.User, db *gorm.DB) []*model.FrontPageEntry {

	var subscriptions []*model.FrontPageEntry

	isAdmin := 0
	if user.IsAdmin {
		isAdmin = 1
	}

	if result := db.Raw("call get_frontpage_subscriptions(?, ?, ?, ?)", user.Id, isAdmin, 0, 1).Scan(&subscriptions); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.PanicWithWrapper(result.Error, utils.ErrInternalError)
		}
	}

	unreadSubs := make([]*model.FrontPageEntry, 0)
	for _, s := range subscriptions {
		if s.PostCount-s.LastPostReadCount > 0 {
			unreadSubs = append(unreadSubs, s)
		}
	}

	return unreadSubs

}

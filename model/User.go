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

	"github.com/dgrijalva/jwt-go"
)

type UserOptionsUpdateData struct {
	ViewType               string `json:"viewType"`
	AutoSubscribe          bool   `json:"autoSubscribe"`
	SortFoldersByActivity  bool   `json:"sortFoldersByActivity"`
	SubscriptionFetchOrder int    `json:"subscriptionFetchOrder"`
	Bio                    string `json:"bio"`
	OldPassword            string `json:"oldPassword"`
	NewPassword            string `json:"newPassword"`
	ResetKey               string `json:"resetKey"`
	RecaptchaResponse      string `json:"recaptchaResponse"`
}

type UserFolderBookmark struct {
	ModelBase
	FolderId     uint      `json:"folderId" gorm:"column:folder_id"`
	UserId       uint      `json:"userId" gorm:"column:user_id"`
	LastPostDate time.Time `json:"lastPostDate" gorm:"column:last_read"`
}

type UserDiscussionBookmark struct {
	ModelBase
	DiscussionId  uint      `json:"discussionId" gorm:"column:discussion_id"`
	UserId        uint      `json:"userId" gorm:"column:user_id"`
	LastPostId    uint      `json:"lastPostId" gorm:"column:last_post_id"`
	LastPostCount int64     `json:"lastPostCount" gorm:"column:last_post_count"`
	LastPostDate  time.Time `json:"lastPostDate" gorm:"column:last_post_date"`
}
type UserFolderSubscription struct {
	ModelBase
	UserId       uint      `json:"userId" gorm:"column:user_id"`
	FolderId     uint      `json:"folderId" gorm:"column:folder_id"`
	LastReadDate time.Time `json:"lastPostDate" gorm:"column:last_read_date"`
	// TODO Url          string    `json:"url" gorm:"-"`
}

type UserFolderSubscriptionException struct {
	ModelBase
	SubscriptionId uint `json:"subscriptionId" gorm:"column:subscription_id"`
	DiscussionId   uint `json:"discussionId" gorm:"column:discussion_id"`
}

type UserDiscussionSubscription struct {
	UserId       uint      `json:"userId" gorm:"column:user_id"`
	DiscussionId uint      `json:"discussionId" gorm:"column:discussion_id"`
	LastReadDate time.Time `json:"lastPostDate" gorm:"column:last_read_date"`
}

type SubscriptionUpdate struct {
	DiscussionId uint       `gorm:"column:id"`
	LastPost     *time.Time `gorm:"column:last_post"`
}

const (
	UserDiscussionStatusBlocked  = -1
	UserDiscussionStatusPending  = 0
	UserDiscussionStatusAccepted = 1
)

type UserDiscussionStatus struct {
	DiscussionId uint `json:"discussionId" gorm:"column:discussion_id"`
	UserId       uint `json:"userId" gorm:"column:user_id"`
	Status       int  `json:"status" gorm:"column:user_status"`
}

type IgnoredUser struct {
	ModelBase
	UserId          uint   `json:"userId" gorm:"column:user_id"`
	IgnoredUserId   uint   `json:"ignoredUserId" gorm:"column:ignored_user_id"`
	IgnoredUserName string `json:"ignoredUserName" gorm:"column:ignored_username"`
}

type User struct {
	ModelBase
	LastLoginDate          time.Time             `json:"lastLoginDate" gorm:"column:last_login_date"`
	Username               string                `json:"username" gorm:"column:username"`
	Email                  string                `json:"email" gorm:"column:email"`
	Bio                    string                `json:"bio" gorm:"column:bio"`
	Password               string                `json:"-" gorm:"column:password"`
	AccountExpired         bool                  `json:"accountExpired" gorm:"column:account_expired"`
	AccountLocked          bool                  `json:"accountLocked" gorm:"column:account_locked"`
	Enabled                bool                  `json:"enabled" gorm:"column:enabled"`
	PasswordExpired        bool                  `json:"passwordExpired" gorm:"column:password_expired"`
	DisplayEmail           bool                  `json:"displayEmail" gorm:"column:display_email"`
	IsAdmin                bool                  `json:"isAdmin" gorm:"column:is_admin"`
	IsPremoderate          bool                  `json:"isPremoderate" gorm:"column:is_premoderate"`
	IsWatch                bool                  `json:"isWatch" gorm:"column:is_watch"`
	IsEmailVerified        bool                  `json:"isEmailVerified" gorm:"column:email_verified"`
	SortFoldersByActivity  bool                  `json:"sortFoldersByActivity" gorm:"column:sort_folders_by_activity"`
	AutoSubscribe          bool                  `json:"autoSubscribe" gorm:"column:auto_subs"`
	SubscriptionFetchOrder int                   `json:"subscriptionFetchOrder" gorm:"column:subs_fetch_order"`
	ViewType               string                `json:"viewType" gorm:"column:view_type"`
	IgnoredUsers           map[uint]*IgnoredUser `json:"ignoredUsers" gorm:"-"`
}

type UserSidebandData struct {
	UserId                       uint                                      `json:"userId"`
	DiscussionBookmarks          map[uint]*UserDiscussionBookmark          `json:"discussionBookmarks" gorm:"-"`
	FolderSubscriptions          map[uint]*UserFolderSubscription          `json:"folderSubscriptions" gorm:"-"`
	FolderSubscriptionExceptions map[uint]*UserFolderSubscriptionException `json:"folderSubscriptionExceptions" gorm:"-"`
	DiscussionSubscriptions      map[uint]*UserDiscussionSubscription      `json:"discussionSubscriptions" gorm:"-"`
}

const UserClaimPurposeAccessToken = "a"
const UserClaimPurposeRefreshToken = "p"

type UserClaims struct {
	UserId  uint   `json:"u"`
	Purpose string `json:"p"`
	jwt.StandardClaims
}

type OtherUser struct {
	UserId      uint      `json:"userId"`
	Username    string    `json:"username" gorm:"column:username"`
	Bio         string    `json:"bio" gorm:"column:bio"`
	CreatedDate time.Time `json:"createdDate" gorm:"column:created_date"`
}

type LoginCredentials struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	Email             string `json:"email"`
	AgreeTerms        bool   `json:"agreeTerms"`
	RecaptchaResponse string `json:"recaptchaResponse"`
	PasswordResetKey  string `json:"key"`
}

type LoginHistory struct {
	Id          uint      `json:"id" gorm:"column:id;primaryKey"`
	Version     int       `json:"version" gorm:"column:version"`
	CreatedDate time.Time `json:"date" gorm:"column:logged_in_date"`
	UserId      uint      `json:"userId" gorm:"column:user_id"`
	IPAddress   string    `json:"ipAddress" gorm:"column:ip_address"`
	Status      string    `json:"status" gorm:"column:session_id"`
}

type SignupConfirmation struct {
	Id              uint      `json:"id" gorm:"column:id;primaryKey"`
	Version         int       `json:"version" gorm:"column:version"`
	CreatedDate     time.Time `json:"createdDate" gorm:"column:created_date"`
	LastUpdatedDate time.Time `json:"lastUpdatedDate" gorm:"column:last_updated"`
	UserId          uint      `json:"userId" gorm:"column:user_id"`
	ConfirmationKey string    `json:"confirmationKey" gorm:"column:confirmation_key"`
	IPAddress       string    `json:"ipAddress" gorm:"column:ip_address"`
	Username        string    `json:"username" gorm:"username"`
}

type PasswordResetRequest struct {
	Id          uint      `json:"id" gorm:"column:id;primaryKey"`
	Version     int       `json:"version" gorm:"column:version"`
	CreatedDate time.Time `json:"createdDate" gorm:"column:created_date"`
	UserId      uint      `json:"userId" gorm:"column:user_id"`
	ResetKey    string    `json:"resetKey" gorm:"column:reset_key"`
	IPAddress   string    `json:"ipAddress" gorm:"column:ip_address"`
	Username    string    `json:"username" gorm:"username"`
}

type UserSearchResults struct {
	ModelBase
	LastLoginDate   time.Time `json:"lastLoginDate" gorm:"column:last_login_date"`
	Username        string    `json:"username" gorm:"column:username"`
	Email           string    `json:"email" gorm:"column:email"`
	AccountExpired  bool      `json:"accountExpired" gorm:"column:account_expired"`
	AccountLocked   bool      `json:"accountLocked" gorm:"column:account_locked"`
	Enabled         bool      `json:"enabled" gorm:"column:enabled"`
	IsAdmin         bool      `json:"isAdmin" gorm:"column:is_admin"`
	IsPremoderate   bool      `json:"isPremoderate" gorm:"column:is_premoderate"`
	IsWatch         bool      `json:"isWatch" gorm:"column:is_watch"`
	IsEmailVerified bool      `json:"isEmailVerified" gorm:"column:email_verified"`
}

type UserHistory struct {
	Id          uint      `json:"id" gorm:"column:id;primaryKey"`
	Version     int       `json:"version" gorm:"column:version"`
	CreatedDate time.Time `json:"createdDate" gorm:"column:created_date"`
	EventType   string    `json:"eventType" gorm:"column:event_type"`
	EventData   string    `json:"eventData" gorm:"column:event_data"`
	UserId      uint      `json:"userId" gorm:"column:user_id"`
}

const UserHistoryAdminPostDelete = "ADMINDELETE"
const UserHistoryAdminPostUndelete = "ADMINUNDELETE"
const UserHistoryAdminDiscussionBlocked = "BLOCK"
const UserHistoryAdminDiscussionUnblocked = "UNBLOCK"
const UserHistoryAdminAccountDeleteEnabled = "DELETE"
const UserHistoryAdminAccountDeleteDisabled = "UNDELETE"
const UserHistoryAdminPostModerated = "POST MODERATED"
const UserHistoryUserPostReported = "POST REPORTED"
const UserHistoryUserReportedPost = "REPORTED POST"
const UserHistoryAdminSignup = "SIGNUP"
const UserHistoryAdminSignupConfirmed = "SIGNUP CONFIRMED"
const UserHistoryAdminAccountLockedEnabled = "LOCK"
const UserHistoryAdminAccountLockedDisabled = "UNLOCK"
const UserHistoryAdminPremodEnabled = "PREMOD"
const UserHistoryAdminPremodDisabled = "UNPREMOD"
const UserHistoryAdminWatchDisabled = "UNWATCH"
const UserHistoryAdminWatchEnabled = "WATCH"

type DiscussionBlock struct {
	Id              uint   `json:"id" gorm:"column:id;primaryKey"`
	DiscussionId    uint   `json:"discussionId" gorm:"column:discussion_id"`
	DiscussionTitle string `json:"discussionTitle" gorm:"column:discussion_name"`
	FolderId        uint   `json:"folderId" gorm:"column:folder_id"`
	FolderKey       string `json:"folderKey" gorm:"column:folder_key"`
	FolderTitle     string `json:"folderTitle" gorm:"column:folder_name"`
	UserId          uint   `json:"userId" gorm:"column:user_id"`
	Username        string `json:"username" gorm:"username"`
	Url             string `json:"url" gorm:"-"`
}

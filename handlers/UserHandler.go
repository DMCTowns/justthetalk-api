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
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"justthetalk/businesslogic"
	"justthetalk/model"
	"justthetalk/utils"
)

type UserHandler struct {
	userCache        *businesslogic.UserCache
	folderCache      *businesslogic.FolderCache
	discussionCache  *businesslogic.DiscussionCache
	emailRegex       *regexp.Regexp
	useSecureCookies bool
}

func NewUserHandler(platform string, userCache *businesslogic.UserCache, folderCache *businesslogic.FolderCache, discussionCache *businesslogic.DiscussionCache) *UserHandler {

	return &UserHandler{
		userCache:        userCache,
		folderCache:      folderCache,
		discussionCache:  discussionCache,
		emailRegex:       regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
		useSecureCookies: (platform == utils.Production),
	}

}

func (h *UserHandler) GetUser(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {
		// just return the user straight from the cache
		return http.StatusOK, user, ""
	})
}

func (h *UserHandler) sendUserWithNewAccessToken(user *model.User) (map[string]interface{}, *http.Cookie) {

	cookie := h.createRefreshTokenCookie(user)

	responseData := make(map[string]interface{})
	responseData["user"] = user
	responseData["accessToken"] = utils.CreateJWT(user, time.Now().Add(15*time.Minute), model.UserClaimPurposeAccessToken)

	return responseData, cookie

}

func (h *UserHandler) Login(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		var credentials model.LoginCredentials
		if err := json.NewDecoder(req.Body).Decode(&credentials); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		if len(credentials.Username) == 0 || len(credentials.Password) == 0 {
			panic(utils.ErrBadRequest)
		}

		user := businesslogic.ValidateUserLogin(credentials, utils.ExtractIPAdress(req), db, h.userCache)

		responseData, cookie := h.sendUserWithNewAccessToken(user)
		http.SetCookie(res, cookie)

		return http.StatusOK, responseData, "Login successful"

	})
}

func (h *UserHandler) Logout(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		if refreshToken, err := req.Cookie("refresh-token"); err == nil {
			h.userCache.ClearRefreshToken(refreshToken.Value)
			h.userCache.Flush(user)
		}

		cookie := h.expiredRefreshTokenCookie()
		http.SetCookie(res, cookie)

		businesslogic.CreateLoginHistory("logout", user, utils.ExtractIPAdress(req), db)

		return http.StatusOK, nil, "User logged out"

	})
}

func extractCookieHack(cookieHeader string) string {

	refreshTokenCookies := []string{}

	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		if strings.HasPrefix(strings.TrimSpace(cookie), "refresh-token") {
			refreshTokenCookies = append(refreshTokenCookies, strings.Split(cookie, "=")[1])
		}
	}

	if len(refreshTokenCookies) == 0 {
		panic(utils.ErrBadRequest)
	}

	sort.Slice(refreshTokenCookies, func(i, j int) bool {
		return len(refreshTokenCookies[i]) > len(refreshTokenCookies[j])
	})

	return refreshTokenCookies[0]

}

func (h *UserHandler) RefreshToken(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		defer func() {
			if r := recover(); r != nil {
				err := r.(error)
				log.Error(err)
				if data, err := json.Marshal(req.Header); err == nil {
					log.Info(string(data))
				}
				panic(err)
			}
		}()

		cookieHeader := req.Header.Get("Cookie")
		if len(cookieHeader) == 0 {
			panic(utils.ErrBadRequest)
		}
		refreshToken := extractCookieHack(cookieHeader)

		token, err := jwt.ParseWithClaims(refreshToken, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(utils.SigningKey), nil
		})

		if err != nil {
			panic(utils.ErrBadRequest)
		}

		claims, ok := token.Claims.(*model.UserClaims)
		if !ok {
			panic(utils.ErrBadRequest)
		}

		if claims.Purpose != model.UserClaimPurposeRefreshToken {
			panic(utils.ErrBadRequest)
		}

		user := h.userCache.Get(claims.UserId)

		cookie := h.createRefreshTokenCookie(user)
		http.SetCookie(res, cookie)

		responseData := make(map[string]interface{})
		responseData["accessToken"] = utils.CreateJWT(user, time.Now().Add(15*time.Minute), model.UserClaimPurposeAccessToken)

		return http.StatusOK, responseData, ""

	})
}

func (h *UserHandler) UpdateAutoSubscribe(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var updateData model.UserOptionsUpdateData
		if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		state := 0
		if updateData.AutoSubscribe {
			state = 1
		}

		updatedUser := businesslogic.UpdateAutoSubscribe(user, state, h.userCache, db)

		return http.StatusOK, updatedUser, ""

	})
}
func (h *UserHandler) UpdateSortFolders(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var updateData model.UserOptionsUpdateData
		if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		state := 0
		if updateData.SortFoldersByActivity {
			state = 1
		}

		updatedUser := businesslogic.UpdateSortFoldersByActivity(user, state, h.userCache, db)

		return http.StatusOK, updatedUser, ""

	})
}

func (h *UserHandler) UpdateSubscriptionFetchOrder(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var updateData model.UserOptionsUpdateData
		if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		updatedUser := businesslogic.UpdateSubscriptionFetchOrder(user, updateData.SubscriptionFetchOrder, h.userCache, db)

		return http.StatusOK, updatedUser, ""

	})
}

func (h *UserHandler) UpdateBio(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var updateData model.UserOptionsUpdateData
		if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		updatedUser := businesslogic.UpdateBio(user, updateData.Bio, h.userCache, db)

		return http.StatusOK, updatedUser, ""

	})
}

func (h *UserHandler) UpdatePassword(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var updateData model.UserOptionsUpdateData
		if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		if err := utils.ValidateRecaptchaResponse(updateData.RecaptchaResponse); err != nil {
			utils.PanicWithWrapper(utils.ErrBadRequest, err)
		}

		businesslogic.UpdatePassword(user, &updateData, h.userCache, db)

		responseData, cookie := h.sendUserWithNewAccessToken(user)
		http.SetCookie(res, cookie)

		return http.StatusOK, responseData, "Password updated"

	})
}

func (h *UserHandler) UpdateViewType(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var updateData model.UserOptionsUpdateData
		if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		if len(updateData.ViewType) == 0 {
			panic(utils.ErrBadRequest)
		}

		updatedUser := businesslogic.UpdateViewType(user, updateData.ViewType, h.userCache, db)
		return http.StatusOK, updatedUser, ""

	})
}

func (h *UserHandler) GetIgnoredUsers(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		ignoredUserList := businesslogic.GetIgnoredUsers(user, db)
		return http.StatusOK, ignoredUserList, ""

	})
}

func (h *UserHandler) UpdateIgnore(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		ignoreUserId := utils.ExtractVarInt("userId", req)
		ignoreState := utils.ExtractQueryInt("state", req)

		businesslogic.UpdateIgnore(user, ignoreUserId, ignoreState, db, h.userCache)

		return http.StatusOK, user, ""

	})
}

func (h *UserHandler) UpdateDiscussionBookmark(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		discussion := h.discussionCache.Get(discussionId, user)

		var lastPost model.Post
		if err := json.NewDecoder(req.Body).Decode(&lastPost); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		currentBookmark := businesslogic.UpdateDiscussionBookmark(user, discussion, &lastPost, db)
		return http.StatusOK, currentBookmark, ""

	})
}

func (h *UserHandler) DeleteDiscussionBookmark(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		discussionId := utils.ExtractVarInt("discussionId", req)
		discussion := h.discussionCache.Get(discussionId, user)

		businesslogic.DeleteDiscussionBookmark(user, discussion, h.userCache, db)

		return http.StatusOK, nil, "Bookmark deleted"

	})
}

func (h *UserHandler) CreateReport(res http.ResponseWriter, req *http.Request) {
	utils.HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var reportData model.PostReport
		if err := json.NewDecoder(req.Body).Decode(&reportData); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		if user != nil {
			reportData.ReporterUserId = user.Id
		}

		if len(reportData.ReporterName) == 0 || len(reportData.ReporterEmail) == 0 || len(reportData.Body) == 0 {
			panic(utils.ErrBadRequest)
		}

		reportData.IPAddress = utils.ExtractIPAdress(req)

		businesslogic.CreateReport(&reportData, h.userCache, db)

		return http.StatusOK, nil, "Report submitted"

	})
}

func (h *UserHandler) CreateUser(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		var credentials model.LoginCredentials
		if err := json.NewDecoder(req.Body).Decode(&credentials); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		if len(credentials.Username) < 4 || len(credentials.Password) < 8 {
			utils.PanicWithWrapper(utils.ErrBadRequest, errors.New("Invalid username or password"))
		}

		if !h.emailRegex.MatchString(credentials.Email) {
			utils.PanicWithWrapper(utils.ErrBadRequest, errors.New("Invalid e-mail address"))
		}

		if err := utils.ValidateRecaptchaResponse(credentials.RecaptchaResponse); err != nil {
			utils.PanicWithWrapper(utils.ErrBadRequest, err)
		}

		user := businesslogic.CreateUser(&credentials, utils.ExtractIPAdress(req), db)

		responseData, cookie := h.sendUserWithNewAccessToken(user)
		http.SetCookie(res, cookie)

		return http.StatusOK, responseData, ""

	})
}

func (h *UserHandler) CheckSubscriptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		subscriptions := businesslogic.CheckSubscriptions(user, db)
		if len(subscriptions) > 0 {
			url := utils.UrlForFrontPageEntry(subscriptions[0])
			return http.StatusOK, url, ""
		} else {
			return http.StatusNoContent, nil, "Subscriptions up to date"
		}

	})
}

func (h *UserHandler) GetDiscussionSubscriptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {
		entries := businesslogic.GetDiscussionSubscriptions(user, db)
		return http.StatusOK, entries, ""
	})
}

func (h *UserHandler) GetFolderSubscriptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {
		folderSubs := businesslogic.GetFolderSubscriptions(user, db)
		return http.StatusOK, folderSubs, ""
	})
}

func (h *UserHandler) GetFolderSubscriptionExceptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {
		exceptions := businesslogic.GetFolderSubscriptionExcepions(user, db)
		return http.StatusOK, exceptions, ""
	})
}

func (h *UserHandler) MarkFolderSubscriptionsRead(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var subsList []uint
		if err := json.NewDecoder(req.Body).Decode(&subsList); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		subscriptons := businesslogic.MarkFolderSubscriptionsRead(subsList, user, db, h.userCache)
		return http.StatusOK, subscriptons, ""

	})
}

func (h *UserHandler) MarkDiscussionSubscriptionsRead(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var subsList []uint
		if err := json.NewDecoder(req.Body).Decode(&subsList); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		subscriptons := businesslogic.MarkDiscussionSubscriptionsRead(subsList, user, db, h.userCache)
		return http.StatusOK, subscriptons, ""

	})
}

func (h *UserHandler) UpdateFolderSubscriptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var subsList []uint
		if err := json.NewDecoder(req.Body).Decode(&subsList); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		subscriptons := businesslogic.UpdateFolderSubscriptions(subsList, user, db, h.userCache, h.folderCache)
		return http.StatusOK, subscriptons, ""

	})
}

func (h *UserHandler) DeleteFolderSubscriptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var subsList []uint
		if err := json.NewDecoder(req.Body).Decode(&subsList); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		result := businesslogic.DeleteFolderSubscriptions(subsList, user, db, h.userCache)

		return http.StatusOK, result, ""

	})
}

func (h *UserHandler) DeleteDiscussionSubscriptions(res http.ResponseWriter, req *http.Request) {
	utils.AuthenticatedHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		var subsList []uint
		if err := json.NewDecoder(req.Body).Decode(&subsList); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		result := businesslogic.DeleteDiscussionSubscriptions(subsList, user, db, h.userCache)

		return http.StatusOK, result, ""

	})
}

func (h *UserHandler) GetOtherUser(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		userId := utils.ExtractVarInt("userId", req)

		user := businesslogic.GetOtherUser(userId, db, h.userCache)

		return http.StatusOK, user, ""

	})
}

func (h *UserHandler) ForgotPassword(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		var credentials model.LoginCredentials
		if err := json.NewDecoder(req.Body).Decode(&credentials); err != nil {
			utils.PanicWithWrapper(err, utils.ErrBadRequest)
		}

		if !h.emailRegex.MatchString(credentials.Email) {
			utils.PanicWithWrapper(utils.ErrBadRequest, errors.New("Invalid e-mail address"))
		}

		if err := utils.ValidateRecaptchaResponse(credentials.RecaptchaResponse); err != nil {
			utils.PanicWithWrapper(utils.ErrBadRequest, err)
		}

		request := businesslogic.ForgotPassword(&credentials, utils.ExtractIPAdress(req), h.userCache, db)

		if request != nil {
			return http.StatusOK, nil, "Request accepted"
		} else {
			return http.StatusBadRequest, nil, "Request not accepted"
		}

	})
}

func (h *UserHandler) ValidatePasswordResetKey(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		resetKey := req.URL.Query().Get("key")
		if len(resetKey) == 0 {
			panic(utils.ErrBadRequest)
		}

		user, _ := businesslogic.ValidatePasswordResetKey(resetKey, h.userCache, db)

		responseData, _ := h.sendUserWithNewAccessToken(user)

		return http.StatusOK, responseData, "Login successful"

	})
}

func (h *UserHandler) ValidateSignupConfirmationKey(res http.ResponseWriter, req *http.Request) {
	utils.AnonymousHandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (int, interface{}, string) {

		confirmationKey := utils.ExtractQueryString("key", req)
		if len(confirmationKey) == 0 {
			panic(utils.ErrBadRequest)
		}

		user, err := businesslogic.ValidateSignupConfirmationKey(confirmationKey, utils.ExtractIPAdress(req), h.userCache, db)
		if err != nil {
			panic(err)
		}

		responseData, cookie := h.sendUserWithNewAccessToken(user)
		http.SetCookie(res, cookie)

		return http.StatusOK, responseData, "Login successful"

	})
}

func (h *UserHandler) createRefreshTokenCookie(user *model.User) *http.Cookie {

	expiryTime := time.Now().Add(time.Hour * 720)

	sameSiteMode := http.SameSiteNoneMode
	if !h.useSecureCookies {
		sameSiteMode = http.SameSiteLaxMode
	}

	return &http.Cookie{
		Name:     "refresh-token",
		Domain:   "justthetalk.com",
		Path:     "/",
		Value:    utils.CreateJWT(user, time.Now().Add(time.Hour*720), model.UserClaimPurposeRefreshToken),
		HttpOnly: true,
		Secure:   h.useSecureCookies,
		SameSite: sameSiteMode,
		Expires:  expiryTime,
	}

}

func (h *UserHandler) expiredRefreshTokenCookie() *http.Cookie {

	sameSiteMode := http.SameSiteNoneMode
	if !h.useSecureCookies {
		sameSiteMode = http.SameSiteLaxMode
	}

	return &http.Cookie{
		Name:     "refresh-token",
		Domain:   "justthetalk.com",
		Path:     "/",
		Value:    "",
		HttpOnly: true,
		Secure:   h.useSecureCookies,
		SameSite: sameSiteMode,
		MaxAge:   0,
	}

}

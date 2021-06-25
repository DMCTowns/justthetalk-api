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

package middleware

import (
	"context"
	"errors"
	"justthetalk/businesslogic"
	"justthetalk/model"
	"justthetalk/utils"
	"net/http"
	"strings"

	"runtime/debug"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type SessionMiddleware struct {
	userCache *businesslogic.UserCache
}

func NewSessionMiddleware(userCache *businesslogic.UserCache) *SessionMiddleware {
	return &SessionMiddleware{
		userCache: userCache,
	}
}

func (m *SessionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		defer func() {
			if r := recover(); r != nil {
				err := r.(error)

				var statusCode int

				switch {
				case errors.Is(err, utils.ErrBadRequest):
					statusCode = http.StatusBadRequest
				case errors.Is(err, utils.ErrUnauthorised):
					statusCode = http.StatusUnauthorized
				case errors.Is(err, utils.ErrForbidden):
					statusCode = http.StatusForbidden
				case errors.Is(err, utils.ErrInternalError):
					statusCode = http.StatusInternalServerError
				case errors.Is(err, utils.ErrNoContent):
					statusCode = http.StatusNoContent
				case errors.Is(err, utils.ErrNotModified):
					statusCode = http.StatusNotModified
				default:
					log.Error(err)
					debug.PrintStack()
					statusCode = http.StatusInternalServerError
				}

				utils.SendRespsonse(statusCode, nil, err.Error(), res)

			}
		}()

		var accessToken string
		authHeader := req.Header.Get(utils.HeaderAuthorization)
		if strings.HasPrefix(authHeader, utils.Bearer) {
			f := strings.Split(authHeader, " ")
			if len(f) == 2 {
				accessToken = f[1]
			}
		}

		if len(accessToken) > 0 {

			token, err := jwt.ParseWithClaims(accessToken, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(utils.SigningKey), nil
			})

			if err != nil {
				panic(utils.ErrBadRequest)
			}

			claims, ok := token.Claims.(*model.UserClaims)
			if !ok {
				panic(utils.ErrBadRequest)
			}

			user := m.userCache.Get(claims.UserId)
			if !(user != nil && user.Id == claims.UserId && !user.AccountLocked && user.Enabled) {
				panic(utils.ErrForbidden)
			}

			ctx := context.WithValue(req.Context(), utils.ContextUserKey, user)
			nextRequest := req.WithContext(ctx)
			next.ServeHTTP(res, nextRequest)

		} else {
			next.ServeHTTP(res, req)
		}

	})
}

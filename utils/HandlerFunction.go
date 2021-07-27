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
	"encoding/json"
	"errors"
	"justthetalk/model"
	"net/http"

	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HandlerFunctionTarget func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (statusCode int, responseData interface{}, message string)
type AuthenticatedHandlerFunctionTarget func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (statusCode int, responseData interface{}, message string)
type AnonymousHandlerFunctionTarget func(res http.ResponseWriter, req *http.Request, db *gorm.DB) (statusCode int, responseData interface{}, message string)

func AdminOnlyHandlerFunction(res http.ResponseWriter, req *http.Request, targetFunc AuthenticatedHandlerFunctionTarget) {
	HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		if user == nil {
			panic(ErrUnauthorised)
		}

		if !user.IsAdmin {
			panic(ErrForbidden)
		}

		return targetFunc(res, req, user, db)

	})
}

func AuthenticatedHandlerFunction(res http.ResponseWriter, req *http.Request, targetFunc AuthenticatedHandlerFunctionTarget) {
	HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {

		if user == nil {
			panic(ErrUnauthorised)
		}

		return targetFunc(res, req, user, db)

	})
}

func AnonymousHandlerFunction(res http.ResponseWriter, req *http.Request, targetFunc AnonymousHandlerFunctionTarget) {
	HandlerFunction(res, req, func(res http.ResponseWriter, req *http.Request, user *model.User, db *gorm.DB) (int, interface{}, string) {
		return targetFunc(res, req, db)
	})
}

func HandlerFunction(res http.ResponseWriter, req *http.Request, targetFunc HandlerFunctionTarget) {

	res.Header().Set(HeaderAccessControlAllowOrigin, req.Header.Get("Origin"))
	res.Header().Set(HeaderVary, "Origin")
	res.Header().Set(HeaderAccessControlAllowCredentials, "true")
	res.Header().Set(HeaderAccessControlAllowHeaders, "Authorization,Content-Type")
	res.Header().Set(HeaderAccessControlAllowMethods, "PUT,POST,GET,DELETE,OPTIONS")

	res.Header().Set(HeaderContentType, ContentTypeJson)

	if req.Method == http.MethodOptions {
		res.WriteHeader(http.StatusOK)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)

			var statusCode int

			switch {
			case errors.Is(err, ErrBadRequest):
				statusCode = http.StatusBadRequest
			case errors.Is(err, ErrUnauthorised):
				statusCode = http.StatusUnauthorized
			case errors.Is(err, ErrForbidden):
				statusCode = http.StatusForbidden
			case errors.Is(err, ErrInternalError):
				statusCode = http.StatusInternalServerError
			case errors.Is(err, ErrNoContent):
				statusCode = http.StatusNoContent
			case errors.Is(err, ErrNotModified):
				statusCode = http.StatusNotModified
			default:
				log.Error(err)
				debug.PrintStack()
				statusCode = http.StatusInternalServerError
			}

			SendRespsonse(statusCode, nil, err.Error(), res)

		}
	}()

	var db *gorm.DB
	var user *model.User
	var ok bool

	db, ok = req.Context().Value(ContextDbKey).(*gorm.DB)
	if !ok {
		panic(ErrInternalError)
	}

	user, _ = req.Context().Value(ContextUserKey).(*model.User)

	statusCode, responseData, message := targetFunc(res, req, user, db)

	res.Header().Set(HeaderCacheControl, "no-store")
	res.Header().Set(HeaderConnection, "Keep-Alive")
	res.Header().Set(HeaderKeepAlive, "timeout=60")

	SendRespsonse(statusCode, responseData, message, res)

}

func SendRespsonse(statusCode int, responseData interface{}, message string, res http.ResponseWriter) {

	result := make(map[string]interface{})

	result["status"] = statusCode
	if responseData != nil {
		result["data"] = responseData
	}

	if len(message) > 0 {
		result["message"] = message
	}

	if data, err := json.Marshal(result); err == nil {
		res.WriteHeader(statusCode)
		_, err := res.Write([]byte(data))
		if err != nil {
			PanicWithWrapper(err, ErrInternalError)
		}
	} else {
		PanicWithWrapper(err, ErrInternalError)
	}

}

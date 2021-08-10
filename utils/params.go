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
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func ExtractPageSizeAndStart(req *http.Request) (int, int) {

	pageSize := ExtractQueryInt("size", req)
	if pageSize == 0 {
		pageSize = 50
	}

	pageStart := ExtractQueryInt("start", req)

	return pageSize, pageStart

}

func ExtractQueryInt64(param string, req *http.Request) int64 {
	return int64(ExtractQueryInt(param, req))
}

func ExtractQueryInt(param string, req *http.Request) int {
	value := 0
	paramValue := req.URL.Query().Get(param)
	if len(paramValue) > 0 {
		if i, err := strconv.Atoi(paramValue); err != nil {
			PanicWithWrapper(err, ErrBadRequest)
		} else {
			value = i
		}
	}
	return value
}

func ExtractVarString(param string, req *http.Request) string {

	value := ""
	vars := mux.Vars(req)

	if param, exists := vars[param]; exists {
		value = param

	} else {
		panic(ErrBadRequest)
	}

	return value

}

func ExtractVarInt(param string, req *http.Request) uint {

	value := 0

	paramValue := ExtractVarString(param, req)
	if i, err := strconv.Atoi(paramValue); err == nil {
		value = i
	} else {
		PanicWithWrapper(err, ErrBadRequest)
	}

	return uint(value)

}

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

// func JsonResponse(status int, message string, res http.ResponseWriter) {

// 	result := make(map[string]interface{})

// 	result["status"] = status
// 	result["message"] = message

// 	if jsonData, err := json.Marshal(result); err == nil {
// 		res.WriteHeader(status)
// 		res.Write([]byte(jsonData))
// 	} else {
// 		PanicWithWrapper(err, ErrInternalError)
// 	}

// }

// func JsonResponseWithData(status int, message string, responseData interface{}, res http.ResponseWriter) {

// 	result := make(map[string]interface{})

// 	result["status"] = status
// 	if len(message) > 0 {
// 		result["message"] = message
// 	}
// 	result["data"] = responseData

// 	if jsonData, err := json.Marshal(result); err == nil {
// 		res.WriteHeader(status)
// 		res.Write([]byte(jsonData))
// 	} else {
// 		PanicWithWrapper(err, ErrInternalError)
// 	}

// }

// func JsonResponseOkWithData(responseData interface{}, res http.ResponseWriter) {
// 	JsonResponseOkWithStatusAndData(http.StatusOK, responseData, res)
// }

// func JsonResponseOkWithStatusAndData(status int, responseData interface{}, res http.ResponseWriter) {

// 	result := make(map[string]interface{})

// 	result["status"] = status
// 	result["data"] = responseData

// 	if jsonData, err := json.Marshal(result); err == nil {
// 		res.WriteHeader(status)
// 		res.Write([]byte(jsonData))
// 	} else {
// 		PanicWithWrapper(err, ErrInternalError)
// 	}

// }

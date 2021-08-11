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

const PlatformEnvVar = "PLATFORM"
const Production = "PRODUCTION"
const Development = "DEVELOPMENT"

const ContextDbKey = "DB"
const ContextUserKey = "User"
const ContextRedisKey = "Redis"

const HeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
const HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
const HeaderAccessControlAllowHeaders = "Access-Control-Allow-Headers"
const HeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"
const HeaderVary = "Vary"

const HeaderContentType = "Content-Type"
const HeaderCacheControl = "Cache-Control"
const HeaderAuthorization = "Authorization"
const HeaderConnection = "Connection"
const HeaderKeepAlive = "Keep-Alive"

const Bearer = "Bearer"

const ContentTypeJson = "application/json; charset=utf-8"

const SigningKey = "xDHk3yWp@$Q8b5x!aPj7=6ekkcwcF#PxgFWENE@X!Vb*XU%_EYgS*ZE@UsnApj*b"

func Abs(val int) int {
	if val >= 0 {
		return val
	} else {
		return -val
	}
}

func Max(val1 int, val2 int) int {
	if val1 > val2 {
		return val1
	} else {
		return val2
	}
}

func Min(val1 int, val2 int) int {
	if val1 < val2 {
		return val1
	} else {
		return val2
	}
}

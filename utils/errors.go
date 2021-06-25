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
	"errors"
	"fmt"
)

var (
	ErrBadRequest    = errors.New("Bad request")
	ErrUnauthorised  = errors.New("Unauthorised")
	ErrForbidden     = errors.New("Forbidden")
	ErrInternalError = errors.New("Internal error")
	ErrNotFound      = errors.New("Not found")
	ErrNoContent     = errors.New("No content")
	ErrNotModified   = errors.New("Not modified")
	ErrExpired       = errors.New("Expired")
)

func PanicWithWrapper(outer error, wrapped error) {
	panic(fmt.Errorf("%v: %w", outer, wrapped))
}

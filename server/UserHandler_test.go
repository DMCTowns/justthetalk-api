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

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestUserEndpointsNeedAuthentication(t *testing.T) {

	testApp := NewApp()

	endpoints := []MethodCall{
		{"POST", "/user/logout"},
		{"PUT", "/user/autosubscribe"},
		{"PUT", "/user/bio"},
		{"PUT", "/user/password"},
		{"PUT", "/user/viewtype"},
		{"PUT", "/user/ignore/{userId}"},
		{"PUT", "/user/folder/{folderId}/subscription"},
		{"PUT", "/user/discussion/{discussionId}/subscription"},
		{"DELETE", "/user/folder/{folderId}/subscription"},
		{"DELETE", "/user/discussion/{discussionId}/subscription"},
		{"DELETE", "/user/discussion/{discussionId}/bookmark"},
		{"GET", "/user/subscription/check"},
		{"GET", "/user/ignored-users"},
		{"DELETE", "/user/ignored-users/{ignoredUserId}"},
		{"GET", "/user/subscriptions/discussion"},
		{"GET", "/user/subscriptions/folder"},
		{"GET", "/user/subscriptions/folder/exceptions"},
	}

	for _, e := range endpoints {
		req, _ := http.NewRequest(e.m, e.u, nil)
		res := testApp.ExecuteTestRequest(req)
		if res.Code != http.StatusUnauthorized {
			t.Errorf("%s: %d", e.u, res.Code)
		}
	}

}

func TestLoginSucceeds(t *testing.T) {

	testApp := NewApp()

	accessToken, _ := GetAccessAndRefreshToken(t, testApp)
	ValidateAccessToken(t, accessToken)

}

func TestInvalidLoginFails(t *testing.T) {

	testApp := NewApp()

	var payload = []byte(`{"username":"testuser1", "password": "xxxxxxxx"}`)

	req, _ := http.NewRequest("POST", "/user/login", bytes.NewBuffer(payload))
	response := testApp.ExecuteTestRequest(req)

	CheckResponseCode(t, http.StatusUnauthorized, response.Code)

}

func TestLockedLoginFails(t *testing.T) {
	t.Fail()
}

func TestLogout(t *testing.T) {

	testApp := NewApp()

	accessToken, _ := GetAccessAndRefreshToken(t, testApp)

	req, _ := http.NewRequest("POST", "/user/logout", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	response := testApp.ExecuteTestRequest(req)
	CheckResponseCode(t, http.StatusOK, response.Code)

	found := false
	cookies := response.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refresh-token" && cookie.MaxAge <= 0 {
			found = true
		}
	}

	if !found {
		t.Error("refresh-token not deleted")
	}

}

func TestRefreshToken(t *testing.T) {

	testApp := NewApp()

	_, refreshToken := GetAccessAndRefreshToken(t, testApp)

	req, _ := http.NewRequest("POST", "/user/refresh-token", nil)
	req.Header.Set("Cookie", fmt.Sprintf("refresh-token=%s", refreshToken))

	response := testApp.ExecuteTestRequest(req)
	CheckResponseCode(t, http.StatusOK, response.Code)

	responseMap := GetResponseMap(t, response)
	if data, exists := responseMap["data"]; exists {
		responseData := data.(map[string]interface{})
		if accessToken, exists := responseData["accessToken"]; exists {
			ValidateAccessToken(t, accessToken.(string))
		}
	}

}

func TestGetUser(t *testing.T) {

	testApp := NewApp()

	accessToken, _ := GetAccessAndRefreshToken(t, testApp)

	req, _ := http.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res := testApp.ExecuteTestRequest(req)
	CheckResponseCode(t, http.StatusOK, res.Code)

	responseMap := make(map[string]interface{})
	err := json.Unmarshal(res.Body.Bytes(), &responseMap)
	if err != nil {
		t.Errorf("Invalid response: %v", err)
	}

}

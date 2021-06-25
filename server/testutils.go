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
	"net/http"
	"net/http/httptest"
	"testing"

	"justthetalk/model"
	"justthetalk/utils"

	"github.com/dgrijalva/jwt-go"
)

type MethodCall struct {
	m string
	u string
}

func DoValidLogin(t *testing.T, testApp *App) *httptest.ResponseRecorder {

	var payload = []byte(`{"username":"testuser1", "password": "1234567890"}`)

	req, _ := http.NewRequest("POST", "/user/login", bytes.NewBuffer(payload))

	res := testApp.ExecuteTestRequest(req)
	if res.Code != http.StatusOK {
		t.Log("Login failed")
		t.FailNow()
	}

	return res

}

func GetResponseMap(t *testing.T, res *httptest.ResponseRecorder) map[string]interface{} {

	responseMap := make(map[string]interface{})
	err := json.Unmarshal(res.Body.Bytes(), &responseMap)
	if err != nil {
		t.Logf("Invalid response: %v", err)
		t.FailNow()
	}

	return responseMap

}

func GetAccessAndRefreshToken(t *testing.T, testApp *App) (string, string) {

	accessToken := ""

	response := DoValidLogin(t, testApp)

	responseMap := GetResponseMap(t, response)
	if data, exists := responseMap["data"]; exists {
		responseData := data.(map[string]interface{})
		if t, exists := responseData["accessToken"]; exists {
			accessToken = t.(string)
		}
	}

	refreshToken := ""
	cookies := response.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refresh-token" {
			refreshToken = cookie.Value
		}
	}

	if len(accessToken) == 0 {
		t.Log("Missing access token")
		t.FailNow()
	}

	if len(refreshToken) == 0 {
		t.Log("Missing refresh token")
		t.FailNow()
	}

	return accessToken, refreshToken

}

func ValidateAccessToken(t *testing.T, accessToken string) {

	token, err := jwt.ParseWithClaims(accessToken, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.SigningKey), nil
	})

	if err != nil {
		t.Errorf("Unable to parse access token: %v", err)
	}

	_, ok := token.Claims.(*model.UserClaims)
	if !ok {
		t.Error("Unable to parse access token")
	}

}

func CheckResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

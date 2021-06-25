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
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const RECAPTCHA_API_ENDPOINT = "https://www.google.com/recaptcha/api/siteverify"

type siteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

func CreateJWT(user *model.User) string {

	claims := model.UserClaims{
		UserId: user.Id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			Issuer:    "justthetalk.com",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if signedToken, err := token.SignedString([]byte(SigningKey)); err == nil {
		return signedToken
	} else {
		panic(ErrInternalError)
	}

}

func ValidateRecaptchaResponse(recaptchaResponse string) error {

	req, err := http.NewRequest(http.MethodPost, RECAPTCHA_API_ENDPOINT, nil)
	if err != nil {
		return err
	}

	secret := os.Getenv("RECAPTCHA_API_KEY")

	// Add necessary request parameters.
	q := req.URL.Query()
	q.Add("secret", secret)
	q.Add("response", recaptchaResponse)
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode response.
	var body siteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	// Check recaptcha verification success.
	if !body.Success {
		log.Infof("%v", body)
		return errors.New("unsuccessful recaptcha verify request")
	}

	if body.ChallengeTS.Add(5 * time.Minute).Before(time.Now()) {
		return errors.New("Recaptcha expired")
	}

	return nil

}

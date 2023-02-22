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
	"context"
	"errors"
	"fmt"
	"justthetalk/businesslogic"
	"justthetalk/connections"
	"justthetalk/model"
	"justthetalk/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"runtime/debug"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	writeWait      = 10 * time.Second
	pingPeriod     = 60 * time.Second
	maxMessageSize = 1024
)

var (
	websocketGuage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "justthetalk_active_websocket_count",
		Help: "Count of active websockets",
	})
	validOrigins = map[string]bool{
		"https://beta.justthetalk.com.": true,
		"https://beta.justthetalk.com":  true,
		"https://justthetalk.com.":      true,
		"https://justthetalk.com":       true,
		"https://justthetalk.co.uk.":      true,
		"https://justthetalk.co.uk":       true,
		"https://www.justthetalk.co.uk.":      true,
		"https://www.justthetalk.co.uk":       true,
		"http://localhost.":      true,
		"http://localhost":       true,
	}
)

type websocketClient struct {
	handler    *WebsockerHandler
	user       *model.User
	connection *websocket.Conn
	writeQueue chan string
	quitFlag   chan bool
	hasQuit    bool
}

type WebsockerHandler struct {
	userCache    *businesslogic.UserCache
	upgrader     websocket.Upgrader
	isProduction bool
}

func NewWebsockerHandler(userCache *businesslogic.UserCache) *WebsockerHandler {

	platform := os.Getenv(utils.PlatformEnvVar)

	websockerHandler := &WebsockerHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		userCache:    userCache,
		isProduction: (platform == utils.Production),
	}

	websockerHandler.upgrader.CheckOrigin = websockerHandler.checkOrigin

	return websockerHandler

}

func (h *WebsockerHandler) Close() {

}

func (h *WebsockerHandler) checkOrigin(req *http.Request) bool {
	if !h.isProduction {
		return true
	}
	origin := req.Header.Get(utils.HeaderOrigin)
	_, exists := validOrigins[origin]
	if !exists {
		log.Errorf("invalid websocket origin: %s", origin)
	}
	return exists
}

func (h *WebsockerHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if conn, err := h.upgrader.Upgrade(res, req, nil); err != nil {
		log.Errorf("%v", err)
	} else {
		NewWebsocketClient(conn, h)
	}

}

func (h *WebsockerHandler) registerClient(client *websocketClient) {
	if client.user == nil {
		panic("no user")
	}
	websocketGuage.Inc()
	h.userCache.AddSubscriber(client.user)
}

func (h *WebsockerHandler) unregisterClient(client *websocketClient) {
	if client.user == nil {
		log.Error("unregisterClient: no user")
		return
	}
	h.userCache.RemoveSubscriber(client.user)
}

func (h *WebsockerHandler) findUser(userId uint) *model.User {
	return h.userCache.Get(userId)
}

func NewWebsocketClient(connection *websocket.Conn, handler *WebsockerHandler) *websocketClient {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("%v", err)
			debug.PrintStack()
		}
	}()

	client := &websocketClient{
		user:       nil,
		connection: connection,
		handler:    handler,
		writeQueue: make(chan string),
		quitFlag:   make(chan bool),
	}

	go client.readWorker()
	go client.writeWorker()

	go func() {

		<-client.quitFlag
		client.hasQuit = true

		log.Debug("Closing client")

		close(client.writeQueue)
		close(client.quitFlag)
		client.connection.Close()

		handler.unregisterClient(client)

	}()

	return client

}

func (client *websocketClient) readWorker() {

	log.Debug("Creating read worker")

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Debugf("%v", err)
			debug.PrintStack()
		}
		if !client.hasQuit {
			client.quitFlag <- true
		}
		log.Debug("Closing read worker")
	}()

	client.connection.SetReadLimit(maxMessageSize)
	err := client.connection.SetReadDeadline(time.Now().Add(pingPeriod * 3))
	if err != nil {
		return
	}

	quit := false
	for !quit {
		_, message, err := client.connection.ReadMessage()
		if err != nil {
			log.Debugf("error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			quit = true
		} else {
			client.processMessage(string(message))
			err := client.connection.SetReadDeadline(time.Now().Add(pingPeriod * 3))
			if err != nil {
				return
			}
		}
	}

}

func (client *websocketClient) writeWorker() {

	log.Debug("Creating write worker")

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Debugf("%v", err)
		}
		if !client.hasQuit {
			client.quitFlag <- true
		}
		log.Debug("Closing write worker")
	}()

	var err error

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	quit := false
	for !quit {
		select {

		case message, ok := <-client.writeQueue:

			if !ok {
				err = client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Debug(err)
				}
				panic(errors.New("channel closed"))
			} else if err = client.send(message); err != nil {
				panic(err)
			}

		case <-ticker.C:
			client.sendPing()

		case <-client.quitFlag:
			quit = true

		}
	}

}

func (client *websocketClient) sendPing() {
	err := client.send("ping!")
	if err != nil {
		log.Debug(err)
	}
}

func (client *websocketClient) sendPong() {
	err := client.send("pong!")
	if err != nil {
		log.Debug(err)
	}
}

func (client *websocketClient) send(message string) error {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("Sending to websocket: %v", err)
			debug.PrintStack()
		}
	}()

	//log.Debug("Sending: " + message)

	err := client.connection.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		return err
	}

	w, err := client.connection.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if nBytes, err := w.Write([]byte(message)); err != nil || nBytes != len(message) {
		return err
	}

	return nil

}

func (client *websocketClient) processMessage(msg string) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Error(err)
			debug.PrintStack()
		}
	}()

	log.Debug(msg)

	f := strings.Split(msg, "!")

	if len(f) != 2 {
		return
	}

	switch f[0] {
	case "hello":
		go client.hello(f[1])
	case "ping":
		//log.Debug("Got ping")
		client.sendPong()
	case "pong":
		//log.Debug("Got pong")
	}

}

func (client *websocketClient) hello(accessToken string) {

	log.Debug("Creating pubsub reader")

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("%v", err)
			debug.PrintStack()
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			if !client.hasQuit {
				client.writeQueue <- "nack!"
			}
		}
		log.Debug("Closing pubsub reader")
	}()

	if len(accessToken) == 0 {
		panic(errors.New("invalid access token"))
	}

	token, err := jwt.ParseWithClaims(accessToken, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.SigningKey), nil
	})

	if err != nil {
		panic(errors.New("invalid access token"))
	}

	claims, ok := token.Claims.(*model.UserClaims)
	if !ok {
		panic(errors.New("invalid access token"))
	}

	client.user = client.handler.findUser(claims.UserId)
	if client.user == nil {
		panic(errors.New("user not found"))
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()

	client.handler.registerClient(client)

	topic := fmt.Sprintf("user:%d", client.user.Id)
	subscription := connections.RedisConnection().Subscribe(ctx, topic)
	if subscription == nil {
		panic(errors.New("subscription is nil"))
	}

	defer func() {
		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second)
		defer cancelFn()
		subscription.Unsubscribe(ctx)
		subscription.Close()
	}()

	client.writeQueue <- "ack!"

	quit := false
	for !quit {
		select {
		case msg := <-subscription.Channel():
			client.writeQueue <- msg.Payload
		case <-client.quitFlag:
			quit = true
		}
	}

}

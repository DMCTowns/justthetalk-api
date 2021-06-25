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
	"errors"
	"justthetalk/businesslogic"
	"justthetalk/model"
	"justthetalk/utils"
	"net/http"
	"strings"
	"time"

	"runtime/debug"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type websocketClient struct {
	handler    *WebsockerHandler
	user       *model.User
	connection *websocket.Conn
	writeQueue chan string
	ticker     *time.Ticker
}

type WebsockerHandler struct {
	userCache *businesslogic.UserCache

	upgrader   websocket.Upgrader
	pingPeriod time.Duration
	writeWait  time.Duration
}

func NewWebsockerHandler(userCache *businesslogic.UserCache) *WebsockerHandler {

	websockerHandler := &WebsockerHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		userCache:  userCache,
		pingPeriod: 30 * time.Second,
		writeWait:  10 * time.Second,
	}

	websockerHandler.upgrader.CheckOrigin = websockerHandler.checkOrigin

	return websockerHandler

}

func (h *WebsockerHandler) Close() {

}

func (h *WebsockerHandler) checkOrigin(req *http.Request) bool {
	origin := req.Header.Get("Origin")
	return strings.HasSuffix(origin, ".justthetalk.com")
}

func (h *WebsockerHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if conn, err := h.upgrader.Upgrade(res, req, nil); err != nil {
		log.Errorf("%v", err)
	} else {
		NewWebsocketClient(conn, h)
	}

}

func (h *WebsockerHandler) registerClient(client *websocketClient) *redis.PubSub {
	return h.userCache.AddSubscriber(client.user)
}

func (h *WebsockerHandler) unregisterClient(client *websocketClient) {
	h.userCache.RemoveSubscriber(client.user)
}

func (h *WebsockerHandler) findUser(userId uint) *model.User {
	return h.userCache.Get(userId)
}

func NewWebsocketClient(connection *websocket.Conn, handler *WebsockerHandler) *websocketClient {

	client := &websocketClient{
		user:       nil,
		connection: connection,
		handler:    handler,
		writeQueue: make(chan string),
	}

	go client.readWorker()
	go client.writeWorker()

	return client

}

func (client *websocketClient) close() {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("%v", err)
			debug.PrintStack()
		}
	}()

	client.ticker.Stop()
	close(client.writeQueue)
	client.connection.Close()

	client.handler.unregisterClient(client)

}

func (client *websocketClient) readWorker() {

	client.connection.SetReadLimit(maxMessageSize)
	client.connection.SetReadDeadline(time.Now().Add(pongWait))
	client.connection.SetPongHandler(func(string) error { client.connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			break
		}
		client.processMessage(string(message))
	}

}

func (client *websocketClient) writeWorker() {

	client.ticker = time.NewTicker(pingPeriod)

	defer func() {
		client.close()
	}()

	for {
		select {

		case message, ok := <-client.writeQueue:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write([]byte(message))
			if err := w.Close(); err != nil {
				return
			}

		case <-client.ticker.C:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}

}

func (client *websocketClient) processMessage(msg string) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Error(err)
			debug.PrintStack()
		}
	}()

	f := strings.Split(msg, "!")

	if len(f) != 2 {
		return
	}

	switch f[0] {
	case "hello":
		client.hello(f[1])
	}

}

func (client *websocketClient) hello(accessToken string) {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("%v", err)
			debug.PrintStack()
			client.close()
		}
	}()

	if len(accessToken) == 0 {
		panic(errors.New("Invalid access token"))
	}

	token, err := jwt.ParseWithClaims(accessToken, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.SigningKey), nil
	})

	if err != nil {
		panic(errors.New("Invalid access token"))
	}

	claims, ok := token.Claims.(*model.UserClaims)
	if !ok {
		panic(errors.New("Invalid access token"))
	}

	client.user = client.handler.findUser(claims.UserId)

	subscription := client.handler.registerClient(client)

	defer func() {
		log.Infof("Exiting socket handler for: %d", client.user.Id)
		subscription.Close()
		client.close()
	}()

	for {
		select {
		case msg := <-subscription.Channel():
			client.writeQueue <- msg.Payload
		}
	}

}

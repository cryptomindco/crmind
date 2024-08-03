// Copyright 2013 Beego Samples authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package controllers

import (
	"crassets/logpack"
	"crassets/models"
	"crassets/utils"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketController handles WebSocket requests.
type WebSocketController struct {
	BaseController
}

// Join method handles WebSocket requests for WebSocketController.
func (this *WebSocketController) Connect() {
	var loginToken string
	if err := this.Ctx.Input.Bind(&loginToken, "token"); err != nil {
		logpack.Error("Can't get login user token", utils.GetFuncName(), err)
		return
	}
	authClaims, err := this.AuthTokenCheck(loginToken)
	if err != nil {
		logpack.Error("Login token failed", utils.GetFuncName(), err)
		return
	}
	loginId := authClaims.Id
	// Upgrade from http request to WebSocket.
	ws, err := websocket.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(this.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		logpack.Error("Cannot setup WebSocket connection", utils.GetFuncName(), err)
		return
	}
	NewConnection(loginId, ws)
	defer Leave(loginId)
	this.ResponseSuccessfully(loginId, "", utils.GetFuncName())
}

// broadcastWebSocket broadcasts messages to WebSocket users.
func broadcastWebSocket(event models.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logpack.Error("Fail to marshal event", utils.GetFuncName(), err)
		return
	}
	for conn := userConnections.Front(); conn != nil; conn = conn.Next() {
		// Immediately send event to WebSocket users.
		ws := conn.Value.(UserConnection).Conn
		if ws != nil {
			if ws.WriteMessage(websocket.TextMessage, data) != nil {
				disconnect <- conn.Value.(UserConnection).UserId
			}
		}
	}
}

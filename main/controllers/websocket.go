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
	"crmind/logpack"
	"crmind/models"
	"crmind/pb/chatpb"
	"crmind/services"
	"crmind/utils"
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
	authClaims, err := this.GetLoginUser()
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
	// Message receive loop.
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			return
		}
		var chatContent models.ChatContent
		parseErr := json.Unmarshal(p, &chatContent)
		if parseErr != nil {
			continue
		}

		res, err := services.GetChatMsgHandler(this.Ctx.Request.Context(), &chatpb.GetChatMsgRequest{
			Common: &chatpb.CommonRequest{
				LoginName: authClaims.Username,
			},
			ChatId: chatContent.ChatId,
		})
		if err != nil {
			return
		}
		var chatMsg models.ChatMsg
		chatParseErr := utils.JsonStringToObject(res.Data, &chatMsg)
		if chatParseErr != nil {
			return
		}
		chatMsgDisplay := models.ChatDisplay{
			ChatMsg:     &chatMsg,
			HasContent:  true,
			LastContent: chatContent,
		}

		msgBytes, marshalErr := json.Marshal(chatMsgDisplay)
		if marshalErr != nil {
			continue
		}
		//send new msg to socket
		publish <- NewEvent(models.EVENT_CHATMSG, loginId, string(msgBytes))
	}
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

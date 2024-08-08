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
	"container/list"
	"crassets/models"
	"crassets/utils"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func NewConnection(userId int64, ws *websocket.Conn) {
	userconnect <- UserConnection{UserId: userId, Conn: ws}
}

func Leave(userId int64) {
	disconnect <- userId
}

var (
	// Send events here to publish them.
	publish         = make(chan models.Event, 10)
	userconnect     = make(chan UserConnection, 10)
	disconnect      = make(chan int64, 10)
	userConnections = list.New()
)

type Subscription struct {
	Archive []models.Event      // All the events from the archive.
	New     <-chan models.Event // New events coming in.
}

type UserConnection struct {
	UserId int64
	Conn   *websocket.Conn // Only for WebSocket users; otherwise nil.
}

func NewEvent(ep models.EventType, userId int64, msg string) models.Event {
	return models.Event{ep, userId, time.Now().Unix(), msg}
}

func init() {
	go SocketEventArea()
	//Every 7 seconds. Exchange rates of all types are updated once and sent to the session
	//Create a connection to socket
	go func() {
		for {
			//Get rate info
			rateObject, err := utils.ReadRateFromDB()
			rateJsonStr := ""
			if err != nil {
				rateJsonStr = "{}"
			} else {
				rateJsonStr = utils.ObjectToJsonString(rateObject)
			}
			//retrive rate for socket
			publish <- NewEvent(models.EVENT_RATE, -1, rateJsonStr)
			time.Sleep(7 * time.Second)
		}
	}()
}

func isUserExist(userConnectionList *list.List, userId int64) bool {
	for conn := userConnectionList.Front(); conn != nil; conn = conn.Next() {
		if conn.Value.(UserConnection).UserId == userId {
			return true
		}
	}
	return false
}

func SocketEventArea() {
	for {
		select {
		case connect := <-userconnect:
			if !isUserExist(userConnections, connect.UserId) {
				userConnections.PushBack(connect)
				publish <- NewEvent(models.EVENT_JOIN, connect.UserId, "Connect with new Conn")
				fmt.Println("New connection for user: ", connect.UserId, ";WebSocket: ", connect.Conn != nil)
			} else {
				fmt.Println("Existed connection of user: ", connect.UserId, ";WebSocket: ", connect.Conn != nil)
			}
		case event := <-publish:
			broadcastWebSocket(event)
			models.NewArchive(event)
		case disconn := <-disconnect:
			for conn := userConnections.Front(); conn != nil; conn = conn.Next() {
				if conn.Value.(UserConnection).UserId == disconn {
					userConnections.Remove(conn)
					// Clone connection.
					ws := conn.Value.(UserConnection).Conn
					if ws != nil {
						ws.Close()
						fmt.Println("WebSocket closed:", disconn)
					}
					publish <- NewEvent(models.EVENT_LEAVE, disconn, "")
					break
				}
			}
		}
	}
}

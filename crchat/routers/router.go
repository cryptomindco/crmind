package routers

import (
	"crchat/controllers"

	beego "github.com/beego/beego/v2/adapter"
)

func init() {
	//Web socket routes
	beego.Router("/ws/connect", &controllers.WebSocketController{}, "get:Connect")

	beego.Router("/updateUnread", &controllers.ChatController{}, "post:UpdateUnreadForChat")
	beego.Router("/deleteChat", &controllers.ChatController{}, "post:DeleteChat")
	beego.Router("/sendChatMessage", &controllers.ChatController{}, "post:SendChatMessage")
	beego.Router("/get-chat-msg", &controllers.ChatController{}, "get:GetChatMsgDisplayList")
}

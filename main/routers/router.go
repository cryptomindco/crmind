package routers

import (
	"crmind/controllers"

	beego "github.com/beego/beego/v2/adapter"
)

// API Gateway initialization
func init() {
	beego.Router("/", &controllers.MainController{})

	//Auth service api gateway
	beego.Router("/passkey/registerStart", &controllers.AuthController{}, "post:BeginRegistration")
	beego.Router("/passkey/registerFinish", &controllers.AuthController{}, "post:FinishRegistration")
	beego.Router("/assertion/options", &controllers.AuthController{}, "post:AssertionOptions")
	beego.Router("/assertion/result", &controllers.AuthController{}, "post:AssertionResult")
	beego.Router("/passkey/updateStart", &controllers.AuthController{}, "post:BeginUpdatePasskey")
	beego.Router("/passkey/updateFinish", &controllers.AuthController{}, "post:FinishUpdatePasskey")
	beego.Router("/passkey/confirmStart", &controllers.AuthController{}, "post:BeginConfirmPasskey")
	beego.Router("/passkey/confirmFinish", &controllers.AuthController{}, "post:FinishConfirmPasskey")
	beego.Router("/passkey/cancelRegister", &controllers.AuthController{}, "post:CancelRegister")
	beego.Router("/passkey/changeUsernameFinish", &controllers.AuthController{}, "post:ChangeUsernameFinish")
	beego.Router("/exit", &controllers.AuthController{}, "get:Quit")
	beego.Router("/gen-random-username", &controllers.AuthController{}, "get:GenRandomUsername")
	beego.Router("/check-user", &controllers.AuthController{}, "get:CheckUser")
	beego.Router("/login", &controllers.AuthController{})
	//Profile router
	beego.Router("/profile", &controllers.ProfileController{})

	//Admin router
	beego.Router("/admin", &controllers.AdminController{})
	beego.Router("/admin/user", &controllers.AdminController{}, "get:UserDetail")
	beego.Router("/admin/ChangeUserStatus", &controllers.AdminController{}, "post:ChangeUserStatus")

	//chat router
	beego.Router("/updateUnread", &controllers.ChatController{}, "post:UpdateUnreadForChat")
	beego.Router("/deleteChat", &controllers.ChatController{}, "post:DeleteChat")
	beego.Router("/sendChatMessage", &controllers.ChatController{}, "post:SendChatMessage")
	//Configure URLs with and without login authentication
	InitSetFilterUrl()
	//Filter, intercept all requests
	beego.InsertFilter("/*", beego.BeforeRouter, FilterCryptomind)
}

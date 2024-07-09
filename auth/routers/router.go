package routers

import (
	"auth/controllers"

	beego "github.com/beego/beego/v2/adapter"
)

func init() {
	//Auth Controller
	beego.Router("/passkey/registerStart", &controllers.AuthController{}, "post:BeginRegistration")
	beego.Router("/passkey/registerFinish", &controllers.AuthController{}, "post:FinishRegistration")
	beego.Router("/assertion/options", &controllers.AuthController{}, "post:AssertionOptions")
	beego.Router("/assertion/result", &controllers.AuthController{}, "post:AssertionResult")
	beego.Router("/passkey/updateStart", &controllers.AuthController{}, "post:BeginUpdatePasskey")
	beego.Router("/passkey/updateFinish", &controllers.AuthController{}, "post:FinishUpdatePasskey")
	beego.Router("/passkey/confirmStart", &controllers.AuthController{}, "post:BeginConfirmPasskey")
	beego.Router("/passkey/confirmFinish", &controllers.AuthController{}, "post:FinishConfirmPasskey")
	beego.Router("/passkey/cancelRegister", &controllers.AuthController{}, "post:CancelRegister")

	//Login/Register routing
	beego.Router("/exit", &controllers.LoginController{}, "get:Quit")
	beego.Router("/gen-random-username", &controllers.LoginController{}, "get:GenRandomUsername")
}

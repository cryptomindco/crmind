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
	beego.Router("/passkey/changeUsernameFinish", &controllers.AuthController{}, "post:ChangeUsernameFinish")
	beego.Router("/auth/syncChangeUsername", &controllers.AuthController{}, "post:SyncUsernameDB")

	//data router
	beego.Router("/admin/get-users", &controllers.QueryController{}, "get:GetAdminUserList")
	beego.Router("/admin/user-info", &controllers.QueryController{}, "get:GetAdminUserInfo")
	beego.Router("/admin/change-user-status", &controllers.QueryController{}, "post:ChangeUserStatus")
	beego.Router("/username-list", &controllers.QueryController{}, "get:GetExcludeLoginUserNameList")
	beego.Router("/user-by-name", &controllers.QueryController{}, "get:GetUserInfoByUsername")

	beego.Router("/is-logging", &controllers.AuthController{}, "get:IsLoggingOn")
	beego.Router("/exit", &controllers.AuthController{}, "get:Quit")
	beego.Router("/gen-random-username", &controllers.AuthController{}, "get:GenRandomUsername")
	beego.Router("/check-user", &controllers.AuthController{}, "get:CheckUser")
}

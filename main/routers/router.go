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
	beego.Router("/settings", &controllers.AdminController{}, "get:GetSettings")
	beego.Router("/updateSettings", &controllers.AdminController{}, "post:UpdateSettings")
	beego.Router("/syncTransactions", &controllers.AdminController{}, "post:SyncTransactions")

	//chat router
	beego.Router("/updateUnread", &controllers.ChatController{}, "post:UpdateUnreadForChat")
	beego.Router("/deleteChat", &controllers.ChatController{}, "post:DeleteChat")
	beego.Router("/sendChatMessage", &controllers.ChatController{}, "post:SendChatMessage")
	//Transfer router
	beego.Router("/transfer/GetHistoryList", &controllers.TransferController{}, "get:FilterTxHistory")
	beego.Router("/check-contact-user", &controllers.TransferController{}, "get:CheckContactUser")
	beego.Router("/confirmAmount", &controllers.TransferController{}, "post:ConfirmAmount")
	beego.Router("/transfer-amount", &controllers.TransferController{}, "post:TransferAmount")
	beego.Router("/updateNewLabel", &controllers.TransferController{}, "post:UpdateNewLabel")
	//Trading router
	beego.Router("/send-trading-request", &controllers.TradingController{}, "post:SendTradingRequest")

	//Assets Router
	beego.Router("/fetch-rate", &controllers.AssetsController{}, "get:FetchRate")
	beego.Router("/assets/detail", &controllers.AssetsController{}, "get:AssetsDetail")
	beego.Router("/GetCodeListData", &controllers.AssetsController{}, "get:GetCodeListData")
	beego.Router("/GetAddressListData", &controllers.AssetsController{}, "get:GetAddressListData")
	beego.Router("/confirmAddressAction", &controllers.AssetsController{}, "post:ConfirmAddressAction")
	beego.Router("/cancelUrlCode", &controllers.AssetsController{}, "post:CancelUrlCode")

	beego.Router("/createNewAddress", &controllers.WalletController{}, "post:CreateNewAddress")

	//Configure URLs with and without login authentication
	InitSetFilterUrl()
	//Filter, intercept all requests
	beego.InsertFilter("/*", beego.BeforeRouter, FilterCryptomind)
}

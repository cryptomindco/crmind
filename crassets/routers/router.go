package routers

import (
	"crassets/controllers"

	beego "github.com/beego/beego/v2/adapter"
)

func init() {
	//Web socket routes
	beego.Router("/ws/connect", &controllers.WebSocketController{}, "get:Connect")

	//Wallet routes
	beego.Router("/createNewAddress", &controllers.WalletController{}, "post:CreateNewAddress")
	beego.Router("/walletSocket", &controllers.WalletController{}, "post:WalletSocket")
	//Admin routes
	beego.Router("/syncTransactions", &controllers.AdminController{}, "post:SyncTransactions")
	//Trading routes
	beego.Router("/send-trading-request", &controllers.TradingController{}, "post:SendTradingRequest")
	//transfer routes
	beego.Router("/transfer-amount", &controllers.TransferController{}, "post:TransferAmount")
	beego.Router("/transfer/GetHistoryList", &controllers.TransferController{}, "get:FilterTxHistory")
	beego.Router("/GetCodeListData", &controllers.TransferController{}, "get:GetCodeListData")
	beego.Router("/GetAddressListData", &controllers.TransferController{}, "get:GetAddressListData")
	beego.Router("/confirmAmount", &controllers.TransferController{}, "post:ConfirmAmount")
	beego.Router("/confirmWithdraw", &controllers.TransferController{}, "post:ConfirmWithdrawal")
	beego.Router("/cancelUrlCode", &controllers.TransferController{}, "post:CancelUrlCode")
	beego.Router("/confirmAddressAction", &controllers.TransferController{}, "post:ConfirmAddressAction")
	beego.Router("/updateNewLabel", &controllers.TransferController{}, "post:UpdateNewLabel")
	beego.Router("/transfer/GetLastTxs", &controllers.TransferController{}, "get:GetLastTxs")
}

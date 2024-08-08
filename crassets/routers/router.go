package routers

import (
	"crassets/controllers"

	beego "github.com/beego/beego/v2/adapter"
)

func init() {
	//Web socket routes
	beego.Router("/ws/connect", &controllers.WebSocketController{}, "get:Connect")

	//Wallet routes
	beego.Router("/wallet/create-new-address", &controllers.WalletController{}, "post:CreateNewAddress")
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
	beego.Router("/updateNewLabel", &controllers.TransferController{}, "post:UpdateNewLabel")
	beego.Router("/transfer/GetLastTxs", &controllers.TransferController{}, "get:GetLastTxs")
	beego.Router("/check-contact-user", &controllers.TransferController{}, "get:CheckContactUser")
	//Assets routes
	beego.Router("/assets/get-balance-summary", &controllers.AssetsController{}, "get:GetBalanceSummary")
	beego.Router("/assets/get-asset-list", &controllers.AssetsController{}, "get:GetAssetDBList")
	beego.Router("/assets/get-user-asset", &controllers.AssetsController{}, "get:GetUserAssetDB")
	beego.Router("/assets/get-address-list", &controllers.AssetsController{}, "get:GetAddressList")
	beego.Router("/assets/count-address", &controllers.AssetsController{}, "get:CountAddress")
	beego.Router("/assets/has-txcodes", &controllers.AssetsController{}, "get:CheckHasCodeList")
	beego.Router("/assets/get-contacts", &controllers.AssetsController{}, "get:GetContactList")
	beego.Router("/assets/filter-txcode", &controllers.AssetsController{}, "get:FilterTxCode")
	beego.Router("/assets/get-txhistory", &controllers.AssetsController{}, "get:GetTxHistory")
	beego.Router("/assets/filter-address-list", &controllers.AssetsController{}, "get:FilterAddressList")
	beego.Router("/assets/create-account-token", &controllers.AssetsController{}, "get:CheckAndCreateAccountToken")
	beego.Router("/fetch-rate", &controllers.AssetsController{}, "get:FetchRate")
	beego.Router("/assets/asset-match-user", &controllers.AssetsController{}, "get:CheckAssetMatchWithUser")
	beego.Router("/assets/address-match-user", &controllers.AssetsController{}, "get:CheckAddressMatchWithUser")
	beego.Router("/assets/get-address", &controllers.AssetsController{}, "get:GetAddress")
	beego.Router("/assets/confirm-address-action", &controllers.AssetsController{}, "post:ConfirmAddressAction")
	beego.Router("/assets/cancel-url-code", &controllers.AssetsController{}, "post:CancelUrlCode")
}

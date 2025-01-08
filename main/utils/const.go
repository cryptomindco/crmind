package utils

import "crmind/models"

type AssetType string
type ServiceType string

const (
	Tokenkey                       = "Token"
	LoginUserKey                   = "AuthClaims"
	UserListSessionKey             = "userList"
	NilAsset           AssetType   = ""
	BTCWalletAsset     AssetType   = "btc"
	DCRWalletAsset     AssetType   = "dcr"
	LTCWalletAsset     AssetType   = "ltc"
	USDWalletAsset     AssetType   = "usd"
	AuthService        ServiceType = "auth"
	ChatService        ServiceType = "chat"
	AssetsService      ServiceType = "assets"
)

var AuthHost, AssetsHost, ChatHost string

var ServiceList = []string{"auth", "chat", "assets"}
var LoginExcludeUrl = []string{"/404", "/exit", "/login", "/LoginSubmit", "/checkLogin", "/register", "/RegisterSubmit", "/walletSocket", "/withdrawl",
	"/confirmWithdraw", "/passkey/registerStart", "/passkey/registerFinish", "/assertion/options",
	"/assertion/result", "/passkey/cancelRegister", "/assertion/withdrawConfirmLoginResult", "/passkey/withdrawWithNewAccountFinish", "/gen-random-username",
	"/check-user", "/password/register", "/password/login"}

var AssetUrl = []string{"/walletSocket", "/withdrawl", "/confirmWithdraw", "/assertion/withdrawConfirmLoginResult", "/passkey/withdrawWithNewAccountFinish",
	"/adminUpdateBalance", "/transfer/GetHistoryList", "/check-contact-user", "/confirmAmount", "/transfer-amount", "/updateNewLabel", "/send-trading-request",
	"/fetch-rate", "/assets/detail", "/GetCodeListData", "/GetAddressListData", "/confirmAddressAction", "/cancelUrlCode", "/transaction/detail", "/createNewAddress"}

var ChatUrl = []string{"/updateUnread", "/deleteChat", "/sendChatMessage"}

type ResponseData struct {
	IsError   bool        `json:"error"`
	ErrorCode string      `json:"errorCode"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
}

type TempoRes struct {
	Exist bool          `json:"exist"`
	Asset *models.Asset `json:"asset"`
}

type UserRole int
type UrlCodeStatus int
type TransType int

const (
	UrlCodeStatusCreated UrlCodeStatus = iota
	UrlCodeStatusConfirmed
	UrlCodeStatusCancelled
)

const (
	RoleSuperAdmin UserRole = iota
	RoleRegular
)

const (
	TransTypeLocal TransType = iota
	TransTypeChainSend
	TransTypeChainReceive
)

func GetAssetColor(assetType string) string {
	switch assetType {
	case string(BTCWalletAsset):
		return "#ebf5ff"
	case string(DCRWalletAsset):
		return "#D4F3E1"
	case string(LTCWalletAsset):
		return "#FFD6F4"
	case string(USDWalletAsset):
		return "#fff2f2"
	default:
		return "#fff2f2"
	}
}

func AssetSortInt(assetType string) int {
	switch assetType {
	case string(BTCWalletAsset):
		return 2
	case string(DCRWalletAsset):
		return 3
	case string(LTCWalletAsset):
		return 4
	case string(USDWalletAsset):
		return 1
	default:
		return 1
	}
}

// ToFull returns the full network name of the provided asset.
func GetAssetFullName(assetType string) string {
	switch assetType {
	case string(BTCWalletAsset):
		return "Bitcoin"
	case string(DCRWalletAsset):
		return "Decred"
	case string(LTCWalletAsset):
		return "Litecoin"
	case string(USDWalletAsset):
		return "US Dollar"
	default:
		return "Unknown"
	}
}

func (urlCodeStatus UrlCodeStatus) ToString() string {
	switch urlCodeStatus {
	case UrlCodeStatusCreated:
		return "Unredeemed"
	case UrlCodeStatusConfirmed:
		return "Redeemed"
	case UrlCodeStatusCancelled:
		return "Cancelled"
	default:
		return ""
	}
}

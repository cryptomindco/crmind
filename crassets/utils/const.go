package utils

import "fmt"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

const (
	CurrencyUsd             = "usd"
	CurrencyBtc             = "btc"
	TransactionTypeSent     = "sent"
	TransactionTypeReceived = "received"
	TransactionTypeFees     = "fee"
	TradingTypeBuy          = "buy"
	TradingTypeSell         = "sell"
)

type UserRole int
type UserStatus int
type TransType int
type AssetStatus int
type UrlCodeStatus int

const (
	StatusDeactive UserStatus = iota
	StatusActive
)

type ResponseData struct {
	IsError bool        `json:"error"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

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
	AssetStatusDeactive AssetStatus = iota
	AssetStatusActive
)

const (
	TransTypeLocal TransType = iota
	TransTypeChainSend
	TransTypeChainReceive
)

func GetUrlCodeStatusFromValue(status int) (UrlCodeStatus, error) {
	switch status {
	case int(UrlCodeStatusCreated):
		return UrlCodeStatusCreated, nil
	case int(UrlCodeStatusConfirmed):
		return UrlCodeStatusConfirmed, nil
	case int(UrlCodeStatusCancelled):
		return UrlCodeStatusCancelled, nil
	default:
		return -1, fmt.Errorf("Get Url Code error with status: %d", status)
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

func (urlCodeStatus UrlCodeStatus) CodeStatusColor() string {
	switch urlCodeStatus {
	case UrlCodeStatusCreated:
		return "#dbc272"
	case UrlCodeStatusConfirmed:
		return "#008000"
	case UrlCodeStatusCancelled:
		return "#dc3545"
	default:
		return ""
	}
}

func (transType TransType) ToString() string {
	switch transType {
	case TransTypeLocal:
		return "Offchain Transaction"
	case TransTypeChainSend:
		return "Onchain Sending"
	case TransTypeChainReceive:
		return "Onchain Receiving"
	default:
		return ""
	}
}

func GetTransTypeFromValue(transType int) TransType {
	switch transType {
	case int(TransTypeLocal):
		return TransTypeLocal
	case int(TransTypeChainSend):
		return TransTypeChainSend
	case int(TransTypeChainReceive):
		return TransTypeChainReceive
	default:
		return TransTypeLocal
	}
}

package models

import "crassets/pkg/walletlib/assets"

type TxHistory struct {
	Id            int64   `json:"id" gorm:"primaryKey"`
	TransType     int     `json:"transType"`
	Sender        string  `json:"sender"`
	Receiver      string  `json:"receiver"`
	Currency      string  `json:"currency"`
	Amount        float64 `json:"amount"`
	Rate          float64 `json:"rate"`
	Status        int     `json:"status"`
	ToAddress     string  `json:"toAddress"`
	Txid          string  `json:"txid"`
	Fee           float64 `json:"fee"`
	IsTrading     bool    `json:"isTrading"`
	TradingType   string  `json:"tradingType"`
	PaymentType   string  `json:"paymentType"`
	Confirmed     bool    `json:"confirmed"`
	Confirmations int     `orm:"default(0)" json:"confirmations"`
	Description   string  `json:"description"`
	Createdt      int64   `orm:"size(10);default(0)" json:"createdt"`
}

type TxHistoryDisplay struct {
	TxHistory
	RateValue            float64                   `json:"rateValue"`
	IsSender             bool                      `json:"isSender"`
	TypeDisplay          string                    `json:"typeDisplay"`
	ConfirmationNeed     int                       `json:"confirmationNeed"`
	Transaction          *assets.TransactionResult `json:"transaction"`
	IsOffChain           bool                      `json:"isOffChain"`
	CreatedtDisp         string                    `json:"createdtDisp"`
	TradingPaymentAmount float64                   `json:"tradingPaymentAmount"`
}

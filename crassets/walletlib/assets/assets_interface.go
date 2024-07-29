package assets

import (
	"encoding/json"
)

type Asset interface {
	CreateRPCClient() error
	ShutdownRPCClient()
	GetNewAddress(username string) (string, error)
	CreateNewAddressWithLabel(username, label string) (string, error)
	GetAddressesByLabel(label string) ([]string, error)
	GetReceivedByAddress(address string) (float64, error)
	GetSendTransactionAmount(txhash string) (float64, error)
	GetReceiveTransactionAmount(txhash string) (float64, error)
	GetTransactionAmount(txhash string, category string) (float64, error)
	GetTransactionFee(txhash string) (float64, error)
	GetWalletPassphrase() string
	CreateRawTransaction(address string, amount float64) (json.RawMessage, error)
	FundRawTransaction(username, rawTransactionHex string) (*FundRawTransactionResult, error)
	SendToAddress(from, to, fromAddress, toAddress string, amount float64) (string, error)
	SendToAccountAddress(account, toAddress string, amount float64) (string, error)
	GetTransactionStatus(txhash string) (TransactionStatus, int64, error)
	GetTransactionByTxhash(txhash string) (*TransactionResult, error)
	CreateNewAccount(username string) error
	GetAccount(address string) (string, error)
	GetLabelList() []string
	GetAllTransactions() ([]ListTransactionsResult, error)
	CheckAndLoadWallet()
	GetBalanceMapByLabel() map[string]float64
	GetSystemLabel() string
	GetReceivedByAccount(account string) (float64, error)
	GetSystemBalance() (float64, error)
	EstimateFeeAndSize(target *TxTarget) (*TxFeeAndSize, error)
	SendToAddressObfuscateUtxos(target *TxTarget) (string, error)
	SetSystemAddress()
	GetSystemAddress() string
	GetSpendableAmount() float64
	UpdateLabel(address string, newLabel string) error
	MutexLock()
	MutexUnlock()
	IsValidAddress(address string) bool
}

package assets

import (
	"math"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/ltcsuite/ltcd/ltcutil"
)

const MainnetString = "mainnet"
const TestnetString = "testnet"
const CategorySend = "send"
const CategoryReceive = "receive"
const DefaultAccount = "default"

type TransactionStatus int

type FundRawTransactionResult struct {
	Hex       string  `json:"hex"`
	Fee       float64 `json:"fee"`
	Changepos int     `json:"changepos"`
}

type TransactionResult struct {
	Amount          float64                       `json:"amount"`
	Fee             float64                       `json:"fee,omitempty"`
	Confirmations   int64                         `json:"confirmations"`
	BlockHash       string                        `json:"blockhash"`
	BlockIndex      int64                         `json:"blockindex"`
	BlockTime       int64                         `json:"blocktime"`
	TxID            string                        `json:"txid"`
	WalletConflicts []string                      `json:"walletconflicts"`
	Time            int64                         `json:"time"`
	TimeReceived    int64                         `json:"timereceived"`
	Details         []GetTransactionDetailsResult `json:"details"`
	Hex             string                        `json:"hex"`
}

type GetTransactionDetailsResult struct {
	Account           string   `json:"account"`
	Address           string   `json:"address,omitempty"`
	Amount            float64  `json:"amount"`
	Category          string   `json:"category"`
	InvolvesWatchOnly bool     `json:"involveswatchonly,omitempty"`
	Fee               *float64 `json:"fee,omitempty"`
	Vout              uint32   `json:"vout"`
}

type TransationNote struct {
	From        string `json:"from"`
	FromAddress string `json:"fromAddress"`
	To          string `json:"to"`
	ToAddress   string `json:"toAddress"`
}

// ListTransactionsResult models the data from the listtransactions command.
type ListTransactionsResult struct {
	Account           string   `json:"account"`
	Address           string   `json:"address,omitempty"`
	Amount            float64  `json:"amount"`
	BlockHash         string   `json:"blockhash,omitempty"`
	BlockHeight       *int32   `json:"blockheight,omitempty"`
	BlockIndex        *int64   `json:"blockindex,omitempty"`
	BlockTime         int64    `json:"blocktime,omitempty"`
	Category          string   `json:"category"`
	Confirmations     int64    `json:"confirmations"`
	Fee               *float64 `json:"fee,omitempty"`
	Generated         bool     `json:"generated,omitempty"`
	InvolvesWatchOnly bool     `json:"involveswatchonly,omitempty"`
	Label             *string  `json:"label,omitempty"`
	Time              int64    `json:"time"`
	TimeReceived      int64    `json:"timereceived"`
	TxID              string   `json:"txid"`
	Vout              uint32   `json:"vout"`
	WalletConflicts   []string `json:"walletconflicts"`
	Comment           string   `json:"comment,omitempty"`
	OtherAccount      string   `json:"otheraccount,omitempty"`
}

type Amount struct {
	// UnitValue holds the base monetary unit value for a cryptocurrency
	UnitValue int64
	// CoinValue holds the monetary amount counted in a cryptocurrency base
	CoinValue float64
}

type TxFeeAndSize struct {
	Fee                 *Amount
	Change              *Amount
	FeeRate             int64 // calculated in Sat/kvB or Lit/kvB
	EstimatedSignedSize int
}

type TxTarget struct {
	FromAddresses []string
	ToAddress     string
	Amount        float64
	UnitAmount    int64
	Account       string
}

type AssetType string

const (
	TransactionUnconfirmed TransactionStatus = iota
	TransactionConfirmed
)

const (
	NilAsset       AssetType = ""
	BTCWalletAsset AssetType = "btc"
	DCRWalletAsset AssetType = "dcr"
	LTCWalletAsset AssetType = "ltc"
	USDWalletAsset AssetType = "usd"

	FullDateformat  = "2006-01-02 15:04:05"
	DateOnlyFormat  = "2006-01-02"
	TimeOnlyformat  = "15:04:05"
	ShortTimeformat = "2006-01-02 15:04"

	BTCFallbackFeeRatePerkvB        btcutil.Amount = 50 * 1000
	LTCFallbackFeeRatePerkvB        ltcutil.Amount = 50 * 1000
	DCRFallbackFeeRatePerkvB        dcrutil.Amount = 0.0001 * 1e8
	DefaultDCRRequiredConfirmations                = 2
	DefaultBTCRequiredConfirmations                = 6
	DefaultLTCRequiredConfirmations                = 6
	ListUnspentMinimumConf                         = 1
	ListUnspentMaximumConf                         = math.MaxInt
)

type UnspentOutput struct {
	TxID          string
	Vout          uint32
	Address       string
	ScriptPubKey  string
	RedeemScript  string
	Amount        AssetAmount
	Confirmations int32
	Spendable     bool
	ReceiveTime   time.Time
	Tree          int8
}

type BlockInfo struct {
	Height    int32
	Timestamp int64
}

type AssetAmount interface {
	// ToCoin returns an asset formatted amount in float64.
	ToCoin() float64
	// String returns an asset formatted amount in string.
	String() string
	// MulF64 multiplies an Amount by a floating point value.
	MulF64(f float64) AssetAmount
	// ToInt() returns the complete int64 value without formatting.
	ToInt() int64
}

func (str AssetType) ToStringUpper() string {
	return strings.ToUpper(string(str))
}

func (str AssetType) AssetSortInt() int {
	switch str {
	case BTCWalletAsset:
		return 2
	case DCRWalletAsset:
		return 3
	case LTCWalletAsset:
		return 4
	case USDWalletAsset:
		return 1
	default:
		return 1
	}
}

func (str AssetType) AssetColor() string {
	switch str {
	case BTCWalletAsset:
		return "#ebf5ff"
	case DCRWalletAsset:
		return "#D4F3E1"
	case LTCWalletAsset:
		return "#FFD6F4"
	case USDWalletAsset:
		return "#fff2f2"
	default:
		return "#fff2f2"
	}
}

// ToFull returns the full network name of the provided asset.
func (str AssetType) ToFullName() string {
	switch str {
	case BTCWalletAsset:
		return "Bitcoin"
	case DCRWalletAsset:
		return "Decred"
	case LTCWalletAsset:
		return "Litecoin"
	case USDWalletAsset:
		return "US Dollar"
	default:
		return "Unknown"
	}
}

func StringToAssetType(assetType string) AssetType {
	switch assetType {
	case "usd":
		return USDWalletAsset
	case "btc":
		return BTCWalletAsset
	case "dcr":
		return DCRWalletAsset
	case "ltc":
		return LTCWalletAsset
	default:
		return NilAsset
	}
}

func AssetTypeSymbol(assetType string) string {
	switch assetType {
	case "usd":
		return "$"
	case "btc":
		return "BTC"
	case "dcr":
		return "DCR"
	case "ltc":
		return "LTC"
	default:
		return ""
	}
}

func (str AssetType) String() string {
	return string(str)
}

package btc

import (
	"crassets/models"
	"crassets/walletlib/assets"
	"sync"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
)

var _ assets.Asset = (*Asset)(nil)

type Asset struct {
	*models.Asset
	AssetType      string
	ChainParam     *chaincfg.Params
	RpcClient      *rpcclient.Client
	ConnectCfg     *rpcclient.ConnConfig
	SystemAddress  string
	TxAuthoredInfo *TxAuthor
	sync.RWMutex
}

type TxAuthor struct {
	Address        string
	Inputs         []*wire.TxIn
	InputValues    []btcutil.Amount
	TxSpendAmount  btcutil.Amount // Equal to fee + send amount
	unsignedTx     *txauthor.AuthoredTx
	needsConstruct bool

	selectedUXTOs []*assets.UnspentOutput
}

// Amount implements the Asset amount interface for the BTC asset
type Amount btcutil.Amount

// ToCoin returns the float64 version of the BTC formatted asset amount.
func (a Amount) ToCoin() float64 {
	return btcutil.Amount(a).ToBTC()
}

// String returns the string version of the BTC formatted asset amount.
func (a Amount) String() string {
	return btcutil.Amount(a).String()
}

// MulF64 multiplys the Amount with the provided float64 value.
func (a Amount) MulF64(f float64) assets.AssetAmount {
	return Amount(btcutil.Amount(a).MulF64(f))
}

// ToInt return the original unformatted amount BTCs
func (a Amount) ToInt() int64 {
	return int64(btcutil.Amount(a))
}

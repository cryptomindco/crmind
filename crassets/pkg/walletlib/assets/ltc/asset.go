package ltc

import (
	"crassets/pkg/config"
	"crassets/pkg/db"
	"crassets/pkg/models"
	"crassets/pkg/walletlib/assets"
	"sync"

	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/ltcsuite/ltcd/rpcclient"
	"github.com/ltcsuite/ltcd/wire"
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
	Handler        db.Handler
	Config         config.Config
	sync.RWMutex
}

type TxAuthor struct {
	Address        string
	Inputs         []*wire.TxIn
	InputValues    []ltcutil.Amount
	TxSpendAmount  ltcutil.Amount // Equal to fee + send amount
	unsignedTx     *AuthoredTx
	needsConstruct bool

	selectedUXTOs []*assets.UnspentOutput
}

// Amount implements the Asset amount interface for the BTC asset
type Amount ltcutil.Amount

// ToCoin returns the float64 version of the BTC formatted asset amount.
func (a Amount) ToCoin() float64 {
	return ltcutil.Amount(a).ToBTC()
}

// String returns the string version of the BTC formatted asset amount.
func (a Amount) String() string {
	return ltcutil.Amount(a).String()
}

// MulF64 multiplys the Amount with the provided float64 value.
func (a Amount) MulF64(f float64) assets.AssetAmount {
	return Amount(ltcutil.Amount(a).MulF64(f))
}

// ToInt return the original unformatted amount BTCs
func (a Amount) ToInt() int64 {
	return int64(ltcutil.Amount(a))
}

package dcr

import (
	"context"
	"crassets/pkg/config"
	"crassets/pkg/db"
	"crassets/pkg/models"
	"crassets/pkg/walletlib/assets"
	"sync"

	"decred.org/dcrwallet/v3/rpc/client/dcrwallet"
	"decred.org/dcrwallet/v4/wallet/txauthor"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/dcrd/wire"
)

var _ assets.Asset = (*Asset)(nil)

type Asset struct {
	*models.Asset
	AssetType      string
	ChainParam     *chaincfg.Params
	RpcClient      *rpcclient.Client
	WalletClient   *dcrwallet.Client
	Ctx            context.Context
	ConnectCfg     *rpcclient.ConnConfig
	SystemAddress  string
	TxAuthoredInfo *TxAuthor
	cancelFuncs    []context.CancelFunc
	Handler        db.Handler
	Config         config.Config
	sync.RWMutex
}

type TxAuthor struct {
	Address        string
	Inputs         []*wire.TxIn
	InputValues    []dcrutil.Amount
	TxSpendAmount  dcrutil.Amount // Equal to fee + send amount
	unsignedTx     *txauthor.AuthoredTx
	needsConstruct bool

	selectedUXTOs []*assets.UnspentOutput
}

// Amount implements the Asset amount interface for the BTC asset
type Amount dcrutil.Amount

// ToCoin returns the float64 version of the BTC formatted asset amount.
func (a Amount) ToCoin() float64 {
	return dcrutil.Amount(a).ToCoin()
}

// String returns the string version of the BTC formatted asset amount.
func (a Amount) String() string {
	return dcrutil.Amount(a).String()
}

// MulF64 multiplys the Amount with the provided float64 value.
func (a Amount) MulF64(f float64) assets.AssetAmount {
	return Amount(dcrutil.Amount(a).MulF64(f))
}

// ToInt return the original unformatted amount BTCs
func (a Amount) ToInt() int64 {
	return int64(dcrutil.Amount(a))
}

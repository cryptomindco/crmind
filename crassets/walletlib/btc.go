package walletlib

import (
	"crassets/walletlib/assets"
	"crassets/walletlib/assets/btc"
)

func initBTCAsset() *btc.Asset {
	asset := &btc.Asset{
		AssetType: assets.BTCWalletAsset.String(),
	}
	return asset
}

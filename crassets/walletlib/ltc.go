package walletlib

import (
	"crassets/walletlib/assets"
	"crassets/walletlib/assets/ltc"
)

func initLTCAsset() *ltc.Asset {
	asset := &ltc.Asset{
		AssetType: assets.LTCWalletAsset.String(),
	}
	return asset
}

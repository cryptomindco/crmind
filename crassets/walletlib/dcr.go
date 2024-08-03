package walletlib

import (
	"crassets/walletlib/assets"
	"crassets/walletlib/assets/dcr"
)

func initDCRAsset() *dcr.Asset {
	asset := &dcr.Asset{
		AssetType: assets.DCRWalletAsset.String(),
	}
	return asset
}

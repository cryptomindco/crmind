package utils

import "crassets/walletlib/assets"

type Globals struct {
	CheckedNet  bool
	Testnet     bool
	AssetsAllow []string
	AssetMgrMap map[string]assets.Asset
}

var GlobalItem *Globals

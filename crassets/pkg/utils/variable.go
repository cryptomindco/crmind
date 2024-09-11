package utils

import "crassets/pkg/walletlib/assets"

type Globals struct {
	CheckedNet     bool
	Testnet        bool
	AssetsAllow    []string
	AssetMgrMap    map[string]assets.Asset
	ExchangeServer string
	PriceSpread    float64
}

var GlobalItem *Globals

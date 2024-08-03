package walletlib

import (
	"crassets/utils"
	"crassets/walletlib/assets"
	"fmt"
	"strings"
)

func NewAssets(assetType assets.AssetType) (assets.Asset, error) {
	var result assets.Asset
	var err error
	switch assetType {
	case assets.BTCWalletAsset:
		result = initBTCAsset()
	case assets.DCRWalletAsset:
		result = initDCRAsset()
	case assets.LTCWalletAsset:
		result = initLTCAsset()
	default:
		err = fmt.Errorf("Create asset from type failed")
		return nil, err
	}
	return result, nil
}

func CreateAssetAndConnectRPC(assetType assets.AssetType) (assets.Asset, error) {
	asset, err := NewAssets(assetType)
	if err != nil {
		return nil, err
	}
	rpcErr := asset.CreateRPCClient()
	if rpcErr != nil {
		return nil, rpcErr
	}
	asset.SetSystemAddress()
	return asset, nil
}

func GetAllowAssetObjectFromSettings() ([]assets.AssetType, error) {
	allowAssets, err := utils.GetAllowAssetNames()
	if err != nil {
		return []assets.AssetType{assets.USDWalletAsset}, nil
	}
	result := make([]assets.AssetType, 0)
	for _, assetName := range allowAssets {
		assetObj := assets.StringToAssetType(strings.TrimSpace(assetName))
		if assetObj != assets.NilAsset {
			result = append(result, assetObj)
		}
	}
	if len(result) == 0 {
		result = append(result, assets.USDWalletAsset)
	}
	return result, nil
}

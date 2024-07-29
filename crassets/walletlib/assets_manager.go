package walletlib

import (
	"crassets/models"
	"crassets/utils"
	"crassets/walletlib/assets"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
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

func GetAllowAssetFromSettings() ([]string, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Settings)).Limit(1).One(&settings)
	if queryErr != nil {
		if queryErr != orm.ErrNoRows {
			return nil, queryErr
		}
		return []string{assets.USDWalletAsset.String()}, nil
	}
	if utils.IsEmpty(settings.ActiveAssets) {
		return []string{assets.USDWalletAsset.String()}, nil
	}
	result := make([]string, 0)
	assetArr := strings.Split(settings.ActiveAssets, ",")
	for _, asset := range assetArr {
		assetObj := assets.StringToAssetType(strings.TrimSpace(asset))
		if assetObj != assets.NilAsset {
			result = append(result, assetObj.String())
		}
	}
	if len(result) == 0 {
		result = append(result, assets.USDWalletAsset.String())
	}
	return result, nil
}

func GetAllowAssetObjectFromSettings() ([]assets.AssetType, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Settings)).Limit(1).One(&settings)
	if queryErr != nil {
		if queryErr != orm.ErrNoRows {
			return nil, queryErr
		}
		return []assets.AssetType{assets.USDWalletAsset}, nil
	}
	if utils.IsEmpty(settings.ActiveAssets) {
		return []assets.AssetType{assets.USDWalletAsset}, nil
	}
	result := make([]assets.AssetType, 0)
	assetArr := strings.Split(settings.ActiveAssets, ",")
	for _, asset := range assetArr {
		assetObj := assets.StringToAssetType(strings.TrimSpace(asset))
		if assetObj != assets.NilAsset {
			result = append(result, assetObj)
		}
	}
	if len(result) == 0 {
		result = append(result, assets.USDWalletAsset)
	}
	return result, nil
}

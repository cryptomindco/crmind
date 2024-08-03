package utils

import (
	"crmind/models"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

func GetAllowAssetFromSettings() ([]string, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Settings)).Limit(1).One(&settings)
	if queryErr != nil {
		if queryErr != orm.ErrNoRows {
			return nil, queryErr
		}
		return []string{"usd"}, nil
	}
	if IsEmpty(settings.ActiveAssets) {
		return []string{"usd"}, nil
	}
	result := make([]string, 0)
	assetArr := strings.Split(settings.ActiveAssets, ",")
	for _, asset := range assetArr {
		if IsEmpty(asset) {
			continue
		}
		result = append(result, asset)
	}
	if len(result) == 0 {
		result = append(result, "usd")
	}
	return result, nil
}

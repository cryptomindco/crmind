package utils

import (
	"crmind/models"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

func GetAllowAssetFromSettings() ([]string, error) {
	activeAssets, err := GetAssetStrFromSettings()
	if err != nil {
		return []string{"usd"}, nil
	}
	result := make([]string, 0)
	assetArr := strings.Split(activeAssets, ",")
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

func GetActiveServicesFromSettings() ([]string, error) {
	activeServices, err := GetServicesStrFromSettings()
	if err != nil {
		return []string{"auth"}, nil
	}
	return HandlerActiveServiceStr(activeServices)
}

func HandlerActiveServiceStr(activeServices string) ([]string, error) {
	result := make([]string, 0)
	serviceArr := strings.Split(activeServices, ",")
	for _, service := range serviceArr {
		if IsEmpty(service) {
			continue
		}
		result = append(result, service)
	}
	if len(result) == 0 {
		result = append(result, "auth")
	}
	return result, nil
}

func GetServicesStrFromSettings() (string, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Settings)).Limit(1).One(&settings)
	if queryErr != nil {
		if queryErr != orm.ErrNoRows {
			return "", queryErr
		}
		return "auth", nil
	}
	if IsEmpty(settings.ActiveServices) {
		return "auth", nil
	}
	return settings.ActiveServices, nil
}

func GetAssetStrFromSettings() (string, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Settings)).Limit(1).One(&settings)
	if queryErr != nil {
		if queryErr != orm.ErrNoRows {
			return "", queryErr
		}
		return "usd", nil
	}
	if IsEmpty(settings.ActiveAssets) {
		return "usd", nil
	}
	return settings.ActiveAssets, nil
}

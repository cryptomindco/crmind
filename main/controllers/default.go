package controllers

import (
	"crmind/models"
	"crmind/utils"
)

var (
	settingsModel = new(models.Settings)
)

type MainController struct {
	BaseController
}

func (this *MainController) Get() {
	loginUser, err := this.AuthCheck()
	if err != nil {
		this.TplName = "login.html"
		return
	}
	assetList := make([]*models.AssetDisplay, 0)
	if this.IsSuperAdmin(*loginUser) {
		assetList, err = this.GetAdminAssetsBalance(loginUser.Username, int64(loginUser.Role))
		if err != nil {
			this.TplName = "err_403.html"
			return
		}
	}
	this.Data["AssetList"] = assetList
	assets, _ := this.GetUserAssetList(loginUser.Username, loginUser.Username)
	this.Data["Assets"] = assets
	//get currency name list of asset list
	currencies := this.GetAssetNamesFromAssetList(assets)
	typeListJson := utils.ObjectToJsonString(currencies)
	this.Data["TypeJson"] = typeListJson

	this.TplName = "index.html"
}

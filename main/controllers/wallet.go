package controllers

import (
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/url"
)

type WalletController struct {
	BaseController
}

func (this *WalletController) CreateNewAddress() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed. Please try again!", utils.GetFuncName(), err)
		return
	}

	//get selected type
	selectedType := this.GetString("assetType")
	if utils.IsEmpty(selectedType) {
		this.ResponseLoginError(loginUser.Id, "Get Asset Type param failed. Please try again!", utils.GetFuncName(), nil)
		return
	}

	var response utils.ResponseData
	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"assetType":     {selectedType},
	}
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AssetsSite(), "/wallet/create-new-address"), formData, &response); err != nil {
		this.ResponseError("Create new address failed", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	address, isOk := response.Data.(string)
	if !isOk {
		this.ResponseError("Get new address failed", utils.GetFuncName(), fmt.Errorf("Get new address failed"))
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Create new address successfully", utils.GetFuncName(), address)
}

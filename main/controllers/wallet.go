package controllers

import (
	"crmind/pb/assetspb"
	"crmind/services"
	"crmind/utils"
	"fmt"
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

	res, err := services.CreateNewAddressHandler(this.Ctx.Request.Context(), &assetspb.OneStringRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
			Role:      int64(loginUser.Role),
		},
		Data: selectedType,
	})

	if err != nil {
		this.ResponseLoginError(loginUser.Id, "Create new address failed", utils.GetFuncName(), err)
		return
	}
	var resMap map[string]string
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		this.ResponseLoginError(loginUser.Id, parseErr.Error(), utils.GetFuncName(), parseErr)
		return
	}
	address, isOk := resMap["address"]
	if !isOk {
		this.ResponseError("Get new address failed", utils.GetFuncName(), fmt.Errorf("Get new address failed"))
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Create new address successfully", utils.GetFuncName(), address)
}

package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/pb/assetspb"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"time"
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

func (this *WalletController) WalletNotify() {
	//Check if txid exist in txhistory
	txid := this.GetString("txid")
	assetType := this.GetString("type")
	if utils.IsEmpty(txid) || utils.IsEmpty(assetType) {
		this.ResponseError("Not enough param to call the processing function", utils.GetFuncName(), nil)
		return
	}
	_, err := services.WalletSocketHandler(this.Ctx.Request.Context(), &assetspb.WalletNotifyRequest{
		Txid: txid,
		Type: assetType,
	})
	if err != nil {
		this.ResponseError(fmt.Sprintf("Update transaction failed for: Type: %s, Txid: %s", assetType, txid), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(0, fmt.Sprintf("Created txhistory successfully: Type: %s, Txid: %s", assetType, txid), utils.GetFuncName())
}

func (this *WalletController) GetWithdrawlAPI() {
	code := this.GetString("code")
	//check valid code, get code
	res, err := services.GetTxCodeHandler(this.Ctx.Request.Context(), &assetspb.OneStringRequest{
		Common: &assetspb.CommonRequest{},
		Data:   code,
	})
	if err != nil {
		logpack.Error("Get code failed. Check assets service server!", utils.GetFuncName(), err)
		this.TplName = "err_403.html"
		return
	}

	var txCode models.TxCode
	parseErr := utils.JsonStringToObject(res.Data, &txCode)
	if parseErr != nil {
		logpack.Error("Get TxCode failed", utils.GetFuncName(), parseErr)
		this.TplName = "err_403.html"
		return
	}
	this.TplName = "transactions/withdraw_confirm.html"
	txCodeDisp := &models.TxCodeDisplay{
		TxCode:          txCode,
		CreatedtDisplay: time.Unix(txCode.Createdt, 0).Format("2006/01/02, 15:04:05"),
	}
	this.Data["TxCode"] = txCodeDisp
}

func (this *WalletController) ConfirmWithdrawal() {
	target := this.GetString("target")
	code := this.GetString("code")
	_, err := services.ConfirmWithdrawalHandler(this.Ctx.Request.Context(), &assetspb.ConfirmWithdrawalRequest{
		Target: target,
		Code:   code,
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(0, "Confirm withdrawl successfully", utils.GetFuncName())
}

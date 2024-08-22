package controllers

import (
	"crmind/pb/assetspb"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"strings"
)

type TransferController struct {
	BaseController
}

func (this *TransferController) TransferAmount() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed. Please try again!", utils.GetFuncName(), err)
		return
	}
	amount, amountErr := this.GetFloat("amount")
	rate, _ := this.GetFloat("rate")
	if amountErr != nil {
		this.ResponseError("Get amount param failed. Please try again!", utils.GetFuncName(), amountErr)
		return
	}
	receiverName := this.GetString("receiver")
	receiverUser, receiverErr := this.GetUserByUsername(receiverName)
	if receiverErr != nil {
		this.ResponseError(receiverErr.Error(), utils.GetFuncName(), receiverErr)
		return
	}
	//Transfer amount
	res, err := services.TransferAmountHandler(this.Ctx.Request.Context(), &assetspb.TransferAmountRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		Currency:     this.GetString("currency"),
		Receiver:     this.GetString("receiver"),
		Note:         this.GetString("note"),
		Address:      this.GetString("address"),
		SendBy:       this.GetString("sendBy"),
		Amount:       amount,
		Rate:         rate,
		ReceiverRole: int64(receiverUser.Role),
	})

	if err != nil {
		this.ResponseError("Transfer amount failed", utils.GetFuncName(), err)
	}
	this.Data["json"] = res
	this.ServeJSON()
}

func (this *TransferController) UpdateNewLabel() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed. Please try again!", utils.GetFuncName(), err)
		return
	}
	assetId, assetIdErr := this.GetInt64("assetId")
	addressId, addressIdErr := this.GetInt64("addressId")
	if assetIdErr != nil || addressIdErr != nil {
		this.ResponseError("Get param failed", utils.GetFuncName(), nil)
		return
	}
	res, err := services.UpdateNewLabelHandler(this.Ctx.Request.Context(), &assetspb.UpdateLabelRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		AssetId:      assetId,
		AddressId:    addressId,
		NewMainLabel: this.GetString("newMainLabel"),
		AssetType:    this.GetString("assetType"),
	})

	if err != nil {
		this.ResponseError("Update new label failed", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = res
	this.ServeJSON()
}

func (this *TransferController) FilterTxHistory() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	allowAssetStr := utils.GetAllowAssets()
	perPage, _ := this.GetInt64("perpage")
	pageNum, _ := this.GetInt64("pageNum")
	res, err := services.FilterTxHistoryHandler(this.Ctx.Request.Context(), &assetspb.FilterTxHistoryRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		AllowAssets: allowAssetStr,
		Type:        strings.TrimSpace(this.GetString("type")),
		Direction:   strings.TrimSpace(this.GetString("direction")),
		PerPage:     perPage,
		PageNum:     pageNum,
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	this.Data["json"] = res
	this.ServeJSON()
}

func (this *TransferController) CheckContactUser() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	username := strings.TrimSpace(this.GetString("username"))
	if utils.IsEmpty(username) {
		this.ResponseLoginError(loginUser.Id, "Username param failed", utils.GetFuncName(), fmt.Errorf("Username param failed"))
		return
	}

	if loginUser.Username == username {
		this.ResponseLoginError(loginUser.Id, "The recipient cannot be you", utils.GetFuncName(), nil)
		return
	}
	res, err := services.CheckContactUserHandler(this.Ctx.Request.Context(), &assetspb.OneStringRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		Data: username,
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	var resMap map[string]bool
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		this.ResponseError(parseErr.Error(), utils.GetFuncName(), parseErr)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check contact user successfully", utils.GetFuncName(), resMap["exist"])
}

func (this *TransferController) ConfirmAmount() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), nil)
		return
	}
	asset := strings.TrimSpace(this.GetString("asset"))
	address := strings.TrimSpace(this.GetString("toaddress"))
	sendBy := strings.TrimSpace(this.GetString("sendBy"))
	amount, amountErr := this.GetFloat("amount")
	if amountErr != nil {
		this.ResponseError("Amount param error", utils.GetFuncName(), nil)
		return
	}
	res, err := services.ConfirmAmountHandler(this.Ctx.Request.Context(), &assetspb.ConfirmAmountRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		Asset:     asset,
		ToAddress: address,
		SendBy:    sendBy,
		Amount:    amount,
	})

	if err != nil {
		this.ResponseLoginError(loginUser.Id, err.Error(), utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = res
	this.ServeJSON()
}

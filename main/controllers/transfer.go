package controllers

import (
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TransferController struct {
	BaseController
}

func (this *TransferController) FilterTxHistory() {
	_, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	allowAssetStr := utils.GetAllowAssets()
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/transfer/GetHistoryList"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"allowassets":   allowAssetStr,
			"type":          strings.TrimSpace(this.GetString("type")),
			"direction":     strings.TrimSpace(this.GetString("direction")),
			"perpage":       strings.TrimSpace(this.GetString("perpage")),
			"pageNum":       strings.TrimSpace(this.GetString("pageNum")),
		},
	}
	err = services.HttpRequest(req, &response)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	this.Data["json"] = response
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

	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/check-contact-user"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"username":      username,
		},
	}
	err = services.HttpRequest(req, &response)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check contact user successfully", utils.GetFuncName(), response.Data)
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
	amountStr := strings.TrimSpace(this.GetString("amount"))

	//send sync request
	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"asset":         {asset},
		"toaddress":     {address},
		"sendBy":        {sendBy},
		"amount":        {amountStr},
	}
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AssetsSite(), "/confirmAmount"), formData, &response); err != nil {
		this.ResponseLoginError(loginUser.Id, "can't send new chat message", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

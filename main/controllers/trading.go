package controllers

import (
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/url"
)

type TradingController struct {
	BaseController
}

func (this *TradingController) SendTradingRequest() {
	_, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed. Please try again!", utils.GetFuncName(), err)
		return
	}
	//get all param
	var response utils.ResponseData
	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"asset":         {this.GetString("asset")},
		"tradingType":   {this.GetString("tradingType")},
		"amount":        {this.GetString("amount")},
		"paymentType":   {this.GetString("paymentType")},
		"rate":          {this.GetString("rate")},
	}
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AssetsSite(), "/send-trading-request"), formData, &response); err != nil {
		this.ResponseError("Send trading request failed", utils.GetFuncName(), err)
		return
	}

	this.Data["json"] = response
	this.ServeJSON()
}

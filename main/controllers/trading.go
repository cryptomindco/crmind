package controllers

import (
	"crmind/pb/assetspb"
	"crmind/services"
	"crmind/utils"
)

type TradingController struct {
	BaseController
}

func (this *TradingController) SendTradingRequest() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed. Please try again!", utils.GetFuncName(), err)
		return
	}
	amount, amountErr := this.GetFloat("amount")
	rate, rateErr := this.GetFloat("rate")
	if amountErr != nil || rateErr != nil {
		this.ResponseError("Get param failed", utils.GetFuncName(), nil)
		return
	}
	res, err := services.SendTradingRequestHandler(this.Ctx.Request.Context(), &assetspb.SendTradingDataRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		Asset:       this.GetString("asset"),
		TradingType: this.GetString("tradingType"),
		PaymentType: this.GetString("paymentType"),
		Amount:      amount,
		Rate:        rate,
	})

	if err != nil {
		this.ResponseError("Send trading request failed", utils.GetFuncName(), err)
		return
	}

	this.Data["json"] = res
	this.ServeJSON()
}

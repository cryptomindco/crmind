package controllers

import (
	"crassets/models"
	"crassets/utils"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type TradingController struct {
	BaseController
}

func (this *TradingController) SendTradingRequest() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	asset := strings.TrimSpace(this.GetString("asset"))
	tradingType := strings.TrimSpace(this.GetString("tradingType"))
	amount, amountErr := this.GetFloat("amount")
	paymentType := strings.TrimSpace(this.GetString("paymentType"))

	rate, rateErr := this.GetFloat("rate")
	if amountErr != nil || rateErr != nil {
		this.ResponseLoginError(loginUser.Id, "Parse param failed. Please check again!", utils.GetFuncName(), nil)
		return
	}

	paymentAmount := amount * rate

	if utils.IsEmpty(asset) {
		this.ResponseLoginError(loginUser.Id, "Asset param failed. Please check again!", utils.GetFuncName(), nil)
		return
	}

	if utils.IsEmpty(tradingType) || (tradingType != utils.TradingTypeBuy && tradingType != utils.TradingTypeSell) {
		this.ResponseLoginError(loginUser.Id, "Trading Type param failed. Please check again!", utils.GetFuncName(), nil)
		return
	}

	if utils.IsEmpty(paymentType) {
		this.ResponseLoginError(loginUser.Id, "Payment Type param failed. Please check again!", utils.GetFuncName(), nil)
		return
	}

	//get asset of system user
	systemAsset, systemAssetErr := utils.GetSystemUserAsset(asset)
	if systemAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get system user asset failed. Please check again!", utils.GetFuncName(), systemAssetErr)
		return
	}

	//Get asset of payment type for system user
	systemPaymentAsset, paymentErr := utils.GetSystemUserAsset(paymentType)
	if paymentErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get system user payment asset failed. Please check again!", utils.GetFuncName(), paymentErr)
		return
	}

	//check valid balance of trading with system user asset
	o := orm.NewOrm()

	//get asset of loginUser
	loginAsset, loginAssetErr := utils.GetAssetByOwner(loginUser.Id, o, asset)
	if loginAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get login user asset failed. Please check again!", utils.GetFuncName(), loginAssetErr)
		return
	}

	//Get asset of payment type for loginuser
	loginPaymentAsset, loginPaymentAssetErr := utils.GetAssetByOwner(loginUser.Id, o, paymentType)
	if loginPaymentAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get login user payment asset failed. Please check again!", utils.GetFuncName(), loginPaymentAssetErr)
		return
	}
	now := time.Now().Unix()

	//check valid balance of asset and payment asset
	//if payment type is buy
	if tradingType == utils.TradingTypeBuy {
		//check balance of payment asset of login user
		if paymentAmount > loginPaymentAsset.Balance {
			//return error
			this.ResponseLoginError(loginUser.Id, fmt.Sprintf("%s balance is not enough to buy this amount", strings.ToUpper(paymentType)), utils.GetFuncName(), nil)
			return
		}
		//check balance of system user of asset
		if amount > systemAsset.Balance {
			this.ResponseLoginError(loginUser.Id, fmt.Sprintf("%s balance of system user is not enough to sell this amount", strings.ToUpper(asset)), utils.GetFuncName(), nil)
			return
		}
		//update amount of all asset for login user and system user
		//login asset
		loginAsset.Balance += amount
		loginAsset.LocalReceived += amount

		//login payment asset
		loginPaymentAsset.Balance -= paymentAmount
		loginPaymentAsset.LocalSent += paymentAmount

		//system user asset
		systemAsset.Balance -= amount
		systemAsset.LocalSent += amount

		//system user payment asset
		systemPaymentAsset.Balance += paymentAmount
		systemPaymentAsset.LocalReceived += paymentAmount
	} else if tradingType == utils.TradingTypeSell {
		//check balance of payment asset of login user
		if amount > loginAsset.Balance {
			//return error
			this.ResponseLoginError(loginUser.Id, fmt.Sprintf("%s balance is not enough to sell this amount", strings.ToUpper(asset)), utils.GetFuncName(), nil)
			return
		}
		//check balance of system user of asset
		if paymentAmount > systemPaymentAsset.Balance {
			this.ResponseLoginError(loginUser.Id, fmt.Sprintf("%s balance of system user is not enough to buy this amount", strings.ToUpper(paymentType)), utils.GetFuncName(), nil)
			return
		}
		//update amount of all asset for login user and system user
		//login asset
		loginAsset.Balance -= amount
		loginAsset.LocalSent += amount
		//login payment asset
		loginPaymentAsset.Balance += paymentAmount
		loginPaymentAsset.LocalReceived += paymentAmount

		//system user asset
		systemAsset.Balance += amount
		systemAsset.LocalReceived += amount

		//system user payment asset
		systemPaymentAsset.Balance -= paymentAmount
		systemPaymentAsset.LocalSent += paymentAmount
	}

	//update all asset
	loginAsset.Updatedt = now
	loginPaymentAsset.Updatedt = now
	systemAsset.Updatedt = now
	systemPaymentAsset.Updatedt = now

	//init tx
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred with DB. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	_, loginAssetUpdateErr := tx.Update(loginAsset)
	_, loginPaymentAssetUpdateErr := tx.Update(loginPaymentAsset)
	_, systemAssetUpdateErr := tx.Update(systemAsset)
	_, systemPaymentAssetUpdateErr := tx.Update(systemPaymentAsset)
	if loginAssetUpdateErr != nil || loginPaymentAssetUpdateErr != nil || systemAssetUpdateErr != nil || systemPaymentAssetUpdateErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Update data for assets failed", utils.GetFuncName(), nil)
		return
	}

	//create description for trading
	note := ""
	if tradingType == utils.TradingTypeBuy {
		note = fmt.Sprintf("Buy from system %f %s. Paid %f %s", amount, strings.ToUpper(asset), paymentAmount, strings.ToUpper(paymentType))
	} else {
		note = fmt.Sprintf("Sell to system %f %s. Received %f %s", amount, strings.ToUpper(asset), paymentAmount, strings.ToUpper(paymentType))
	}

	//insert new txhistory
	txHistory := &models.TxHistory{
		TransType:   int(utils.TransTypeLocal),
		SenderId:    loginUser.Id,
		Sender:      loginUser.Username,
		Currency:    asset,
		Amount:      amount,
		Rate:        rate,
		Status:      int(utils.StatusActive),
		IsTrading:   true,
		TradingType: tradingType,
		PaymentType: paymentType,
		Description: note,
		Createdt:    time.Now().Unix(),
	}

	_, txHistoryInsertErr := tx.Insert(txHistory)
	if txHistoryInsertErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Insert transaction history failed", utils.GetFuncName(), txHistoryInsertErr)
		return
	}
	tx.Commit()
	//return successfully response
	this.ResponseSuccessfully(loginUser.Id, "Transaction completed", utils.GetFuncName())
}

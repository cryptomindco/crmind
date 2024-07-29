package controllers

import (
	"crassets/handler"
	"crassets/logpack"
	"crassets/models"
	"crassets/services"
	"crassets/utils"
	"crassets/walletlib/assets"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type WalletController struct {
	BaseController
}

// Handler receiver transaction when have notify from bitcoin core return
func (this *WalletController) WalletSocket() {
	o := orm.NewOrm()
	//Check if txid exist in txhistory
	txid := this.GetString("txid")
	assetType := this.GetString("type")

	if utils.IsEmpty(txid) || utils.IsEmpty(assetType) {
		this.ResponseError("Not enough param to call the processing function", utils.GetFuncName(), nil)
		return
	}

	txCount, txErr := o.QueryTable(txHistoryModel).Filter("txid", txid).Count()
	var txExist = txErr == nil && txCount > 0
	//if txid is existed in txhistory, return
	if txExist {
		this.ResponseError("Txid has been processed in txhistory", utils.GetFuncName(), nil)
		return
	}
	//Check Asset Manager
	handler.UpdateAssetManagerByType(assetType)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !assetMgrExist {
		this.ResponseError("Create RPC client failed", utils.GetFuncName(), nil)
		return
	}

	transResult, transErr := assetObj.GetTransactionByTxhash(txid)
	if transErr != nil {
		this.ResponseError("Get transaction from txid failed", utils.GetFuncName(), transErr)
		return
	}

	//Get received address of transaction
	receivedAddress := ""
	amount := float64(0)
	for _, detail := range transResult.Details {
		if detail.Category == assets.CategoryReceive {
			receivedAddress = detail.Address
			amount = detail.Amount
		}
	}

	if utils.IsEmpty(receivedAddress) {
		this.ResponseError("Get address of transaction failed", utils.GetFuncName(), nil)
		return
	}

	//Get asset from address
	asset, assetErr := utils.GetAssetFromAddress(receivedAddress, assetType)
	if assetErr != nil || asset == nil {
		this.ResponseError("No asset with the address exist", utils.GetFuncName(), assetErr)
		return
	}

	//update address
	addressObj, addrErr := utils.GetAddress(receivedAddress)
	if addrErr != nil {
		this.ResponseError("Get address object failed", utils.GetFuncName(), addrErr)
		return
	}

	//check achived address
	if addressObj.Archived {
		this.ResponseError("The address has been archived. This transaction cannot be updated", utils.GetFuncName(), nil)
		return
	}

	//update asset
	isConfirmed := utils.IsConfirmed(transResult.Confirmations, assetType)
	if isConfirmed {
		asset.Balance += amount
		asset.OnChainBalance += amount
		asset.ChainReceived += amount
	}
	asset.Updatedt = time.Now().Unix()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseError("Initialize error Transaction DB to insert txhistory", utils.GetFuncName(), beginErr)
		return
	}

	addressObj.ChainReceived += amount
	//update address object
	_, addressUpdateErr := tx.Update(addressObj)
	if addressUpdateErr != nil {
		this.ResponseRollbackError(tx, "Update address object error", utils.GetFuncName(), addressUpdateErr)
		return
	}

	//Insert to txhistory
	var transactionService services.TransactionService
	price, err := transactionService.GetExchangePrice(assetType)
	if err != nil {
		price = 0
	}
	txHistory := models.TxHistory{}
	txHistory.ReceiverId = asset.UserId
	txHistory.Receiver = asset.UserName
	txHistory.ToAddress = receivedAddress
	txHistory.Currency = assetType
	txHistory.Amount = amount
	txHistory.Rate = price
	txHistory.Txid = txid
	txHistory.TransType = int(utils.TransTypeChainReceive)
	txHistory.Status = 1
	txHistory.Description = fmt.Sprintf("Received %s from blockchain", strings.ToUpper(assetType))
	txHistory.Createdt = time.Now().Unix()
	txHistory.Confirmed = isConfirmed
	txHistory.Confirmations = int(transResult.Confirmations)
	_, HistoryErr := tx.Insert(&txHistory)
	if HistoryErr != nil {
		this.ResponseRollbackError(tx, "Insert DB txhistory error", utils.GetFuncName(), HistoryErr)
		return
	}
	tx.Commit()
	logpack.Info(fmt.Sprintf("Created txhistory successfully. Asset: %s, Txid =%s", assetType, txid), utils.GetFuncName())
	this.ServeJSON()
}

func (this *WalletController) CreateNewAddress() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	//get selected type
	selectedType := this.GetString("assetType")
	if utils.IsEmpty(selectedType) {
		this.ResponseLoginError(loginUser.Id, "Get Asset Type param failed. Please try again!", utils.GetFuncName(), nil)
		return
	}
	assetObject := assets.StringToAssetType(selectedType)
	asset, _, createErr := this.CreateNewAddressForAsset(*loginUser, assetObject)
	if createErr != nil {
		this.ResponseLoginError(loginUser.Id, createErr.Error(), utils.GetFuncName(), nil)
		return
	}

	this.Data["json"] = map[string]string{"error": "", "address": asset.Address}
	this.ServeJSON()
}

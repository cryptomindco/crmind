package services

import (
	"context"
	"crassets/pkg/models"
	"crassets/pkg/pb"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"fmt"
	"strings"
	"time"
)

// Handler receiver transaction when have notify from bitcoin core return
func (s *Server) WalletSocket(ctx context.Context, reqData *pb.WalletNotifyRequest) (*pb.ResponseData, error) {
	//Check if txid exist in txhistory
	txid := reqData.Txid
	assetType := reqData.Type

	if utils.IsEmpty(txid) || utils.IsEmpty(assetType) {
		return ResponseError("Not enough param to call the processing function", utils.GetFuncName(), nil)
	}
	var txCount int64
	txErr := s.H.DB.Model(&models.TxHistory{}).Where("txid = ?", txid).Count(&txCount).Error
	if txErr != nil {
		return ResponseError("Check txid exist on DB failed", utils.GetFuncName(), txErr)
	}
	var txExist = txCount > 0
	//if txid is existed in txhistory, return
	if txExist {
		return ResponseError("Txid has been processed in txhistory", utils.GetFuncName(), nil)
	}
	//Check Asset Manager
	s.UpdateAssetManagerByType(assetType)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !assetMgrExist {
		return ResponseError("Create RPC client failed", utils.GetFuncName(), nil)
	}

	transResult, transErr := assetObj.GetTransactionByTxhash(txid)
	if transErr != nil {
		return ResponseError("Get transaction from txid failed", utils.GetFuncName(), transErr)
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
		return ResponseError("Get address of transaction failed", utils.GetFuncName(), nil)
	}

	//Get asset from address
	asset, assetErr := s.H.GetAssetFromAddress(receivedAddress, assetType)
	if assetErr != nil || asset == nil {
		return ResponseError("No asset with the address exist", utils.GetFuncName(), assetErr)
	}

	//update address
	addressObj, addrErr := s.H.GetAddress(receivedAddress)
	if addrErr != nil {
		return ResponseError("Get address object failed", utils.GetFuncName(), addrErr)
	}

	//check achived address
	if addressObj.Archived {
		return ResponseError("The address has been archived. This transaction cannot be updated", utils.GetFuncName(), nil)
	}

	//update asset
	isConfirmed := utils.IsConfirmed(transResult.Confirmations, assetType)
	if isConfirmed {
		asset.Balance += amount
		asset.OnChainBalance += amount
		asset.ChainReceived += amount
	}
	asset.Updatedt = time.Now().Unix()
	tx := s.H.DB.Begin()
	addressObj.ChainReceived += amount
	addressObj.Transactions++
	//update address object
	addressUpdateErr := tx.Save(addressObj).Error
	if addressUpdateErr != nil {
		tx.Rollback()
		return ResponseError("Update address object error", utils.GetFuncName(), addressUpdateErr)
	}

	//Insert to txhistory
	price, err := s.GetExchangePrice(assetType)
	if err != nil {
		price = 0
	}
	txHistory := models.TxHistory{}
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
	HistoryErr := tx.Create(&txHistory).Error
	if HistoryErr != nil {
		tx.Rollback()
		return ResponseError("Insert DB txhistory error", utils.GetFuncName(), HistoryErr)
	}
	tx.Commit()
	return ResponseSuccessfully("", fmt.Sprintf("Created txhistory successfully. Asset: %s, Txid =%s", assetType, txid), utils.GetFuncName())
}

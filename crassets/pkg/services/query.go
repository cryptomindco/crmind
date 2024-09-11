package services

import (
	"context"
	"crassets/pkg/config"
	"crassets/pkg/db"
	"crassets/pkg/logpack"
	"crassets/pkg/models"
	"crassets/pkg/pb"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Server struct {
	H    db.Handler // handler
	Conf config.Config
	pb.UnimplementedAssetsServiceServer
}

func (s *Server) CreateNewAddress(ctx context.Context, reqData *pb.OneStringRequest) (*pb.ResponseData, error) {
	//get selected type
	selectedType := reqData.Data
	if utils.IsEmpty(selectedType) {
		return ResponseError("Get Asset Type param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetObject := assets.StringToAssetType(selectedType)
	address, _, createErr := s.CreateNewAddressForAsset(reqData.Common.LoginName, utils.IsSuperAdmin(int(reqData.Common.Role)), assetObject)
	if createErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, createErr.Error(), utils.GetFuncName(), nil)
	}

	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Create new address successfully", utils.GetFuncName(), map[string]string{
		"address": address.Address,
	})
}

func (s *Server) GetTxCode(ctx context.Context, reqData *pb.OneStringRequest) (*pb.ResponseData, error) {
	code := reqData.Data
	if utils.IsEmpty(code) {
		return ResponseError("Get code param failed", utils.GetFuncName(), nil)
	}
	txCode, exist := s.H.GetTxcode(code)
	if !exist {
		return ResponseError("TxCode does not exist", utils.GetFuncName(), nil)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Tx Code successfully", utils.GetFuncName(), txCode)
}

func (s *Server) AdminUpdateBalance(ctx context.Context, reqData *pb.AdminBalanceUpdateRequest) (*pb.ResponseData, error) {
	inputValue := reqData.Input
	username := reqData.Username
	typeStr := reqData.Type
	action := reqData.Action
	if inputValue == 0 || utils.IsEmpty(username) || utils.IsEmpty(typeStr) {
		return ResponseLoginError(reqData.Common.LoginName, "Get param value failed! Please try again", utils.GetFuncName(), nil)
	}
	if !utils.IsSuperAdmin(int(reqData.Common.Role)) {
		return ResponseLoginError(reqData.Common.LoginName, "No access to this feature", utils.GetFuncName(), nil)
	}
	//get balance of asset
	asset, assetErr := s.H.GetUserAsset(username, typeStr)
	if assetErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get asset of user failed. Please check asset table on DB", utils.GetFuncName(), assetErr)
	}
	//if asset is nil, create new asset for user
	assetObj := assets.StringToAssetType(typeStr)
	tx := s.H.DB.Begin()
	if asset == nil {
		if assetObj == assets.USDWalletAsset {
			asset = &models.Asset{
				DisplayName: assetObj.ToFullName(),
				UserName:    username,
				Type:        assetObj.String(),
				Sort:        assetObj.AssetSortInt(),
				Status:      int(utils.AssetStatusActive),
				Createdt:    time.Now().Unix(),
				Updatedt:    time.Now().Unix(),
			}
			//insert asset
			assetInsertErr := tx.Create(asset).Error
			if assetInsertErr != nil {
				tx.Rollback()
				return ResponseLoginError(reqData.Common.LoginName, "Insert new asset error", utils.GetFuncName(), assetInsertErr)
			}
		} else {
			//create new address and asset
			_, asset, assetErr = s.H.CreateNewAddressForAsset(username, utils.IsSuperAdmin(int(reqData.UserRole)), assetObj)
			if assetErr != nil {
				tx.Rollback()
				return ResponseLoginError(reqData.Common.LoginName, "Create new asset and address failed. Please try again!", utils.GetFuncName(), assetErr)
			}
		}
	}
	//if withdraw and inputvalue more than balance
	if action == utils.AdminActionWithdraw && inputValue > asset.Balance {
		tx.Rollback()
		return ResponseLoginError(reqData.Common.LoginName, "The balance is not enough to withdraw", utils.GetFuncName(), nil)
	}
	//if action type is update
	if action == utils.AdminActionUpdate {
		if inputValue == asset.Balance {
			tx.Rollback()
			return ResponseLoginError(reqData.Common.LoginName, "The balance does not change. Please try again!", utils.GetFuncName(), nil)
		}
		//if input value is greater than the balance, is deposit type
		if inputValue > asset.Balance {
			action = utils.AdminActionDeposit
			inputValue = inputValue - asset.Balance
		} else {
			//else, is withdraw type
			action = utils.AdminActionWithdraw
			inputValue = asset.Balance - inputValue
		}
	}

	//if deposit, add to balance
	if action == utils.AdminActionDeposit {
		asset.Balance += inputValue
		asset.LocalReceived += inputValue
	} else if action == utils.AdminActionWithdraw {
		asset.Balance -= inputValue
		asset.LocalSent += inputValue
	}
	asset.Updatedt = time.Now().Unix()
	//update
	assetUpdateErr := tx.Save(asset).Error
	if assetUpdateErr != nil {
		tx.Rollback()
		return ResponseLoginError(reqData.Common.LoginName, "Update asset of user failed!", utils.GetFuncName(), assetUpdateErr)
	}
	//insert to txhistory
	txHistory := models.TxHistory{}
	note := ""
	//create rate for cryptocurrency
	rate := float64(0)
	isCryptoCurrency := utils.IsCryptoCurrency(typeStr)
	if isCryptoCurrency {
		price, err := s.GetExchangePrice(typeStr)
		if err != nil {
			price = 0
		}
		rate = price
	}
	if action == utils.AdminActionDeposit {
		txHistory.Sender = reqData.Common.LoginName
		txHistory.Receiver = username
		if typeStr == assets.USDWalletAsset.String() {
			note = fmt.Sprintf("Deposited $%.2f by superadmin", inputValue)
		} else {
			note = fmt.Sprintf("Deposited $%.8f by superadmin", inputValue)
		}
	} else if action == utils.AdminActionWithdraw {
		txHistory.Sender = username
		txHistory.Receiver = reqData.Common.LoginName
		if typeStr == assets.USDWalletAsset.String() {
			note = fmt.Sprintf("$%.2f withdrawn by superadmin", inputValue)
		} else {
			note = fmt.Sprintf("$%.8f withdrawn by superadmin", inputValue)
		}
	}
	txHistory.Rate = rate
	txHistory.Currency = typeStr
	txHistory.Amount = inputValue
	txHistory.Status = 1
	txHistory.Description = note
	txHistory.Createdt = time.Now().UnixNano() / 1e9
	txHistory.TransType = int(utils.TransTypeLocal)
	HistoryErr := tx.Create(&txHistory).Error
	if HistoryErr != nil {
		tx.Rollback()
		return ResponseLoginError(reqData.Common.LoginName, "Insert transaction history failed!", utils.GetFuncName(), HistoryErr)
	}
	tx.Commit()
	return ResponseSuccessfully(reqData.Common.LoginName, "Updated the user's balance successfully", utils.GetFuncName())
}

func (s *Server) TransactionDetail(ctx context.Context, reqData *pb.OneIntegerRequest) (*pb.ResponseData, error) {
	historyId := reqData.Data
	if historyId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "History ID param error", utils.GetFuncName(), nil)
	}
	history, err := s.H.GetTxHistoryById(historyId)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	var transaction *assets.TransactionResult
	var neededConfirmation int
	if history.TransType != int(utils.TransTypeLocal) {
		s.UpdateAssetManagerByType(history.Currency)
		assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[history.Currency]
		//get confirm count
		if !utils.IsEmpty(history.Txid) && assetMgrExist {
			transaction, _ = assetObj.GetTransactionByTxhash(history.Txid)
			if history.Currency == assets.DCRWalletAsset.String() {
				neededConfirmation = 2
			} else {
				neededConfirmation = 6
			}
		}
	}
	createDt := time.Unix(history.Createdt, 0)
	historyDisp := models.TxHistoryDisplay{
		TxHistory:            *history,
		IsSender:             reqData.Common.LoginName == history.Sender,
		RateValue:            history.Rate * history.Amount,
		CreatedtDisp:         createDt.Format("2006/01/02, 15:04:05"),
		Transaction:          transaction,
		ConfirmationNeed:     neededConfirmation,
		IsOffChain:           history.TransType == int(utils.TransTypeLocal),
		TypeDisplay:          utils.GetTransTypeFromValue(history.TransType).ToString(),
		TradingPaymentAmount: history.Rate * history.Amount,
	}

	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Transaction detail successfully", utils.GetFuncName(), historyDisp)
}

func (s *Server) SyncTransactions(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	if reqData.Role != int64(utils.RoleSuperAdmin) {
		return ResponseError("Check admin login failed", utils.GetFuncName(), nil)
	}
	//create task to sync transactions
	go func() {
		s.SystemSyncHandler()
	}()
	return ResponseSuccessfully(reqData.LoginName, "Synchronized transaction successfully", utils.GetFuncName())
}

func (s *Server) SendTradingRequest(ctx context.Context, reqData *pb.SendTradingDataRequest) (*pb.ResponseData, error) {
	asset := reqData.Asset
	tradingType := reqData.TradingType
	amount := reqData.Amount
	paymentType := reqData.PaymentType
	rate := reqData.Rate
	paymentAmount := amount * rate

	if utils.IsEmpty(asset) || utils.IsEmpty(tradingType) || amount == 0 || utils.IsEmpty(paymentType) || rate == 0 {
		return ResponseError("Get param failed. Please try again!", utils.GetFuncName(), nil)
	}

	if tradingType != utils.TradingTypeBuy && tradingType != utils.TradingTypeSell {
		return ResponseLoginError(reqData.Common.LoginName, "Trading Type param failed. Please check again!", utils.GetFuncName(), nil)
	}

	//get asset of system user
	systemAsset, systemAssetErr := s.H.GetSystemUserAsset(asset)
	if systemAssetErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get system user asset failed. Please check again!", utils.GetFuncName(), systemAssetErr)
	}

	//Get asset of payment type for system user
	systemPaymentAsset, paymentErr := s.H.GetSystemUserAsset(paymentType)
	if paymentErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get system user payment asset failed. Please check again!", utils.GetFuncName(), paymentErr)
	}

	//get asset of loginUser
	loginAsset, loginAssetErr := s.H.GetUserAsset(reqData.Common.LoginName, asset)
	if loginAssetErr != nil || loginAsset == nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get login user asset failed. Please check again!", utils.GetFuncName(), loginAssetErr)
	}

	//Get asset of payment type for loginuser
	loginPaymentAsset, loginPaymentAssetErr := s.H.GetUserAsset(reqData.Common.LoginName, paymentType)
	if loginPaymentAssetErr != nil || loginPaymentAsset == nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get login user payment asset failed. Please check again!", utils.GetFuncName(), loginPaymentAssetErr)
	}
	now := time.Now().Unix()

	//check valid balance of asset and payment asset
	//if payment type is buy
	if tradingType == utils.TradingTypeBuy {
		//check balance of payment asset of login user
		if paymentAmount > loginPaymentAsset.Balance {
			//return error
			return ResponseLoginError(reqData.Common.LoginName, fmt.Sprintf("%s balance is not enough to buy this amount", strings.ToUpper(paymentType)), utils.GetFuncName(), nil)
		}
		//check balance of system user of asset
		if amount > systemAsset.Balance {
			return ResponseLoginError(reqData.Common.LoginName, fmt.Sprintf("%s balance of system user is not enough to sell this amount", strings.ToUpper(asset)), utils.GetFuncName(), nil)
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
			return ResponseLoginError(reqData.Common.LoginName, fmt.Sprintf("%s balance is not enough to sell this amount", strings.ToUpper(asset)), utils.GetFuncName(), nil)
		}
		//check balance of system user of asset
		if paymentAmount > systemPaymentAsset.Balance {
			return ResponseLoginError(reqData.Common.LoginName, fmt.Sprintf("%s balance of system user is not enough to buy this amount", strings.ToUpper(paymentType)), utils.GetFuncName(), nil)
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
	tx := s.H.DB.Begin()
	loginAssetUpdateErr := tx.Save(loginAsset).Error
	loginPaymentAssetUpdateErr := tx.Save(loginPaymentAsset).Error
	systemAssetUpdateErr := tx.Save(systemAsset).Error
	systemPaymentAssetUpdateErr := tx.Save(systemPaymentAsset).Error
	if loginAssetUpdateErr != nil || loginPaymentAssetUpdateErr != nil || systemAssetUpdateErr != nil || systemPaymentAssetUpdateErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update data for assets failed", utils.GetFuncName(), nil)
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
		Sender:      reqData.Common.LoginName,
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

	txHistoryInsertErr := tx.Create(txHistory).Error
	if txHistoryInsertErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Insert transaction history failed", utils.GetFuncName(), txHistoryInsertErr)
	}
	tx.Commit()
	//return successfully response
	return ResponseSuccessfully(reqData.Common.LoginName, "Transaction completed", utils.GetFuncName())
}

func (s *Server) ConfirmWithdrawal(ctx context.Context, reqData *pb.ConfirmWithdrawalRequest) (*pb.ResponseData, error) {
	target := reqData.Target
	code := reqData.Code
	if utils.IsEmpty(target) || utils.IsEmpty(code) {
		return ResponseError("Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	//Check valid token and get user of token
	txCode, exist := s.H.GetTxcode(code)
	if !exist {
		return ResponseError("Retrieve TxCode data error", utils.GetFuncName(), nil)
	}

	//current rate
	rateSend := s.H.GetRateFromDBByAsset(txCode.Asset)
	txCode.Note = fmt.Sprintf("%s: Withdraw with URL Code", txCode.Note)
	//Get address
	if utils.IsEmpty(target) {
		return ResponseError("Address cannot be left blank. Please check again!", utils.GetFuncName(), nil)
	}
	isSystemAddress := false
	//check to address is of system address
	addressObj, addrErr := s.H.GetAddress(target)
	if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
		return ResponseError("DB error. Please try again!", utils.GetFuncName(), addrErr)
	}
	var existAsset *models.Asset
	if addressObj != nil {
		var assetErr error
		existAsset, assetErr = s.H.GetAssetById(addressObj.AssetId)
		if assetErr != nil && assetErr != gorm.ErrRecordNotFound {
			return ResponseError("DB error. Please try again!", utils.GetFuncName(), assetErr)
		}
		//if exist user for address
		if existAsset != nil {
			isSystemAddress = true
		}
	}

	if !isSystemAddress {
		//change note of txCode
		//create sender
		sender := &models.UserInfo{
			Username: txCode.OwnerName,
		}
		//Check RPC client
		s.UpdateAssetManagerByType(txCode.Asset)
		assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[txCode.Asset]
		if !assetMgrExist {
			return ResponseError("Create RPC Client failed", utils.GetFuncName(), nil)
		}
		validAddress := assetObj.IsValidAddress(target)
		if !validAddress {
			return ResponseError("Invalid address", utils.GetFuncName(), nil)
		}
		txHistory, btcHandlerErr := s.H.HandlerTransferOnchainCryptocurrency(sender.Username, txCode.Asset, target, txCode.Amount, rateSend, txCode.Note)
		if btcHandlerErr != nil {
			return ResponseError(btcHandlerErr.Error(), utils.GetFuncName(), nil)
		}
		//update txCode status
		txCode.Status = int(utils.UrlCodeStatusConfirmed)
		txCode.Txid = txHistory.Txid
		txCode.HistoryId = txHistory.Id
		txCode.Confirmdt = time.Now().Unix()
		tx := s.H.DB.Begin()
		txUpdateErr := tx.Save(txCode).Error
		if txUpdateErr != nil {
			tx.Rollback()
			return ResponseError("Update Tx Code status failed!", utils.GetFuncName(), txUpdateErr)
		}
		tx.Commit()
		return ResponseSuccessfully("", "Confirm withdrawl successfully", utils.GetFuncName())
	} else {
		user := &models.UserInfo{
			Username: existAsset.UserName,
		}
		completeInternal := s.H.HandlerInternalWithdrawl(txCode, *user, rateSend)
		if !completeInternal {
			return ResponseError("Hanlder internal withdrawl failed", utils.GetFuncName(), nil)
		}
		return ResponseSuccessfully("", "Withdrawl with internal account successfully", utils.GetFuncName())
	}
}

func (s *Server) UpdateNewLabel(ctx context.Context, reqData *pb.UpdateLabelRequest) (*pb.ResponseData, error) {
	assetId := reqData.AssetId
	addressId := reqData.AddressId
	newMainLabel := reqData.NewMainLabel
	assetType := reqData.AssetType

	if assetId < 1 || addressId < 1 || utils.IsEmpty(newMainLabel) || utils.IsEmpty(assetType) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}

	if !s.H.CheckMatchAddressWithUser(assetId, addressId, reqData.Common.LoginName, false) {
		return ResponseLoginError(reqData.Common.LoginName, "The user login information and assets do not match", utils.GetFuncName(), nil)
	}

	s.UpdateAssetManagerByType(assetType)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !assetMgrExist {
		return ResponseLoginError(reqData.Common.LoginName, "Create RPC Client failed!", utils.GetFuncName(), nil)
	}

	//get address
	address, addressErr := s.H.GetAddressById(addressId)
	if addressErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get address from DB failed", utils.GetFuncName(), addressErr)
	}

	//get token of loginuser
	token := s.H.GetTokenFromUsername(reqData.Common.LoginName)
	if utils.IsEmpty(token) {
		return ResponseLoginError(reqData.Common.LoginName, "Get user token failed", utils.GetFuncName(), nil)
	}
	//set new label for address on DB
	newLabel := fmt.Sprintf("%s_%s", token, newMainLabel)
	address.Label = newLabel
	tx := s.H.DB.Begin()
	addressUpdateErr := tx.Save(address).Error
	if addressUpdateErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update address label from DB failed", utils.GetFuncName(), addressUpdateErr)
	}
	//update label on daemon
	daemonUpdateErr := assetObj.UpdateLabel(address.Address, newLabel)
	if daemonUpdateErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update address label on daemon failed", utils.GetFuncName(), daemonUpdateErr)
	}
	tx.Commit()
	return ResponseSuccessfully(reqData.Common.LoginName, "Update address label successfully", utils.GetFuncName())
}

func (s *Server) GetAddressListDataWithStatus(ctx context.Context, reqData *pb.GetAddressListRequest) (*pb.ResponseData, error) {
	assetId := reqData.AssetId
	status := reqData.Status
	if assetId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	//check asset id match with loginUser
	assetMatch := s.H.CheckAssetMatchWithUser(assetId, reqData.Common.LoginName)
	if !assetMatch {
		return ResponseError("asset not match with login user", utils.GetFuncName(), fmt.Errorf("asset not match with login user"))
	}
	//Get url code list
	addressList, addressErr := s.H.FilterAddressList(assetId, status)
	if addressErr != nil {
		return ResponseError(addressErr.Error(), utils.GetFuncName(), addressErr)
	}
	//user token
	token, _, err := s.H.CheckAndCreateAccountToken(reqData.Common.LoginName, int(reqData.Common.Role))
	if err != nil {
		return ResponseError("Check or create user token failed", utils.GetFuncName(), fmt.Errorf("Check or create user token failed"))
	}
	addressDisplayList := make([]*models.AddressesDisplay, 0)
	for _, address := range addressList {
		//check label of address
		var mainLabel = ""
		if !utils.IsEmpty(address.Label) && strings.Contains(address.Label, fmt.Sprintf("%s_", token)) {
			mainLabel = strings.ReplaceAll(address.Label, fmt.Sprintf("%s_", token), "")
		}
		addressDisp := &models.AddressesDisplay{
			CreatedtDisplay: utils.GetDateTimeDisplay(address.Createdt),
			Addresses:       address,
			TotalReceived:   address.ChainReceived + address.LocalReceived,
			LabelMainPart:   mainLabel,
		}
		addressDisplayList = append(addressDisplayList, addressDisp)
	}

	resultMap := make(map[string]any)
	resultMap["list"] = addressDisplayList
	resultMap["userToken"] = token
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Address List data successfully", utils.GetFuncName(), resultMap)
}
func (s *Server) GetCodeListData(ctx context.Context, reqData *pb.GetCodeListRequest) (*pb.ResponseData, error) {
	assetType := reqData.Asset
	status := reqData.CodeStatus
	if utils.IsEmpty(assetType) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	//Get url code list
	urlCodeList, urlCodeErr := s.H.FilterUrlCodeList(assetType, status, reqData.Common.LoginName)
	if urlCodeErr != nil {
		return ResponseError(urlCodeErr.Error(), utils.GetFuncName(), urlCodeErr)
	}
	urlCodeDisplayList := make([]*models.TxCodeDisplay, 0)
	for _, urlCode := range urlCodeList {
		urlCodeStatus, statusErr := utils.GetUrlCodeStatusFromValue(urlCode.Status)
		if statusErr != nil {
			logpack.FError(statusErr.Error(), reqData.Common.LoginName, utils.GetFuncName(), nil)
			continue
		}
		txCodeDisp := &models.TxCodeDisplay{
			TxCode:           urlCode,
			CreatedtDisplay:  utils.GetDateTimeDisplay(urlCode.Createdt),
			ConfirmdtDisplay: utils.GetDateTimeDisplay(urlCode.Confirmdt),
			StatusDisplay:    urlCodeStatus.ToString(),
			IsCancelled:      urlCode.Status == int(utils.UrlCodeStatusCancelled),
			IsConfirmed:      urlCode.Status == int(utils.UrlCodeStatusConfirmed),
			IsCreatedt:       urlCode.Status == int(utils.UrlCodeStatusCreated),
		}
		if urlCode.Status == int(utils.UrlCodeStatusConfirmed) && urlCode.HistoryId > 0 {
			history, err := s.H.GetTxHistoryById(urlCode.HistoryId)
			if err == nil {
				txCodeDisp.TxHistory = history
			}

		}
		urlCodeDisplayList = append(urlCodeDisplayList, txCodeDisp)
	}

	resultMap := make(map[string]any)
	resultMap["list"] = urlCodeDisplayList
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Code List data successfully", utils.GetFuncName(), resultMap)
}

func (s *Server) FilterTxHistory(ctx context.Context, reqData *pb.FilterTxHistoryRequest) (*pb.ResponseData, error) {
	perpage := reqData.PerPage
	pageNum := reqData.PageNum
	if perpage < 1 {
		perpage = 15
	}
	if pageNum < 1 {
		pageNum = 1
	}
	allowAssetStr := reqData.AllowAssets
	assetType := reqData.Type
	direction := reqData.Direction
	if utils.IsEmpty(allowAssetStr) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	allowAssets := utils.GetAssetsNameFromStr(allowAssetStr)
	if assetType == "all" {
		assetType = ""
	}

	if direction == "all" {
		direction = ""
	}

	txHistoryList, pageCount, totalRowCount := s.H.InitTransactionHistoryList(&models.UserInfo{
		Username: reqData.Common.LoginName,
	}, assetType, direction, perpage, pageNum, allowAssets)
	resultMap := make(map[string]any)
	resultMap["pageCount"] = pageCount
	resultMap["list"] = txHistoryList
	resultMap["rowsCount"] = totalRowCount
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Tx History List successfully", utils.GetFuncName(), resultMap)
}

func (s *Server) ConfirmAmount(ctx context.Context, reqData *pb.ConfirmAmountRequest) (*pb.ResponseData, error) {
	asset := reqData.Asset
	address := reqData.ToAddress
	sendBy := reqData.SendBy
	amountToSend := reqData.Amount
	if utils.IsEmpty(asset) || amountToSend == 0 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	s.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return ResponseLoginError(reqData.Common.LoginName, "Create RPC Client failed!", utils.GetFuncName(), nil)
	}
	if utils.IsEmpty(address) {
		//if has no address from param. Get address of system admin to check
		address = assetObj.GetSystemAddress()
		if utils.IsEmpty(address) {
			return ResponseLoginError(reqData.Common.LoginName, "Address param failed. Please try again!", utils.GetFuncName(), nil)
		}
	}

	//check to address is of system address
	addressObj, addrErr := s.H.GetAddress(address)
	if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
		return ResponseLoginError(reqData.Common.LoginName, "DB error. Please try again!", utils.GetFuncName(), addrErr)
	}

	if addressObj != nil && sendBy != "urlcode" {
		existAsset, assetErr := s.H.GetAssetById(addressObj.AssetId)
		if assetErr != nil && assetErr != gorm.ErrRecordNotFound {
			return ResponseLoginError(reqData.Common.LoginName, "DB error. Please try again!", utils.GetFuncName(), addrErr)
		}
		//if exist user for address
		if existAsset != nil {
			return ResponseLoginErrorWithCode(reqData.Common.LoginName, "exist", fmt.Sprintf("<span>Is the user's address: <span class=\"fw-600 fs-16\">%s</span>. You'll not be charged transaction fees</span>", existAsset.UserName), utils.GetFuncName(), nil)
		}
	}
	//check valid address
	validAddress := utils.CheckValidAddress(asset, address)
	if !validAddress {
		return ResponseLoginError(reqData.Common.LoginName, "Invalid address. Please check again!", utils.GetFuncName(), nil)
	}
	//Get login user asset
	loginAsset, loginAssetErr := s.H.GetUserAsset(reqData.Common.LoginName, asset)
	if loginAssetErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get Asset of loginUser failed. Please check again!", utils.GetFuncName(), loginAssetErr)
	}

	fromAddressList, addrListErr := s.H.GetAddressListByAssetId(loginAsset.Id)
	if addrListErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get address list of loginUser failed or address list empty. Please check again!", utils.GetFuncName(), addrListErr)
	}
	//start estimate fee and size
	unitAmount, unitAmountErr := utils.GetUnitAmount(amountToSend, asset)
	if unitAmountErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get UnitAmount from amount failed!", utils.GetFuncName(), unitAmountErr)
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()
	feeAndSize, err := assetObj.EstimateFeeAndSize(&assets.TxTarget{
		FromAddresses: fromAddressList,
		ToAddress:     address,
		Amount:        amountToSend,
		UnitAmount:    unitAmount,
		Account:       reqData.Common.LoginName,
	})
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get Estimate fee and size failed!", utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Confirm amount successfully", utils.GetFuncName(), map[string]float64{"fee": feeAndSize.Fee.CoinValue})
}

func (s *Server) AddToContact(ctx context.Context, reqData *pb.OneStringRequest) (*pb.ResponseData, error) {
	currentContacts, contactErr := s.H.GetContactListFromUser(reqData.Common.LoginName)
	if contactErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Parse contact of loginUser failed", utils.GetFuncName(), contactErr)
	}
	targetName := reqData.Data
	if utils.IsEmpty(targetName) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	//check receiveruser exist on contact
	isExist := utils.CheckUserExistOnContactList(targetName, currentContacts)
	if !isExist {
		//add to contact
		currentContacts = append(currentContacts, models.ContactItem{
			UserName: targetName,
			Addeddt:  time.Now().Unix(),
		})
		jsonByte, err := json.Marshal(currentContacts)
		if err != nil {
			return ResponseLoginError(reqData.Common.LoginName, "Parse contacts failed", utils.GetFuncName(), err)
		}
		if err == nil {
			contactStr := string(jsonByte)
			updateErr := s.H.UpdateUserContacts(reqData.Common.LoginName, contactStr)
			if updateErr != nil {
				return ResponseLoginError(reqData.Common.LoginName, "Update user contacts failed", utils.GetFuncName(), updateErr)
			}
		}
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Update contacts successfully", utils.GetFuncName(), isExist)
}

func (s *Server) TransferAmount(ctx context.Context, reqData *pb.TransferAmountRequest) (*pb.ResponseData, error) {
	currency := reqData.Currency
	amountToSend := reqData.Amount
	receiverUsername := reqData.Receiver
	rateSend := reqData.Rate
	note := reqData.Note
	sendBy := reqData.SendBy
	address := reqData.Address
	receiverRole := reqData.ReceiverRole
	addToContact := reqData.AddToContact
	addedContacts := false
	if utils.IsEmpty(currency) || amountToSend == 0 || utils.IsEmpty(sendBy) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	tx := s.H.DB.Begin()
	if currency == assets.USDWalletAsset.String() || (currency != assets.USDWalletAsset.String() && sendBy == "username") {
		if utils.IsEmpty(receiverUsername) {
			return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Recipient information cannot be left blank. Please try again!", utils.GetFuncName(), nil)
		}
		// check and get contact
		if addToContact {
			//get account from username
			account, err := s.H.GetAccountFromUsername(reqData.Common.LoginName)
			if err != nil && err != gorm.ErrRecordNotFound {
				return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Check account from DB failed", utils.GetFuncName(), err)
			}
			currentContacts := make([]models.ContactItem, 0)
			isCreate := err == gorm.ErrRecordNotFound
			if account != nil && err != gorm.ErrRecordNotFound && !utils.IsEmpty(account.Contacts) {
				parseErr := utils.JsonStringToObject(account.Contacts, &currentContacts)
				if parseErr != nil {
					return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Get account contacts failed", utils.GetFuncName(), parseErr)
				}
			}
			//check receiveruser exist on contact
			isExist := utils.CheckUserExistOnContactList(receiverUsername, currentContacts)
			if isExist {
				logpack.Warn("The recipient already exists in contact", utils.GetFuncName())
			} else {
				//add to contact
				currentContacts = append(currentContacts, models.ContactItem{
					UserName: receiverUsername,
					Addeddt:  time.Now().Unix(),
				})
				jsonByte, err := json.Marshal(currentContacts)
				if err == nil {
					var saveContactErr error
					if isCreate {
						account = &models.Accounts{
							Username: reqData.Common.LoginName,
							Role:     int(reqData.Common.Role),
							Contacts: string(jsonByte),
						}
						saveContactErr = tx.Create(account).Error
					} else {
						account.Contacts = string(jsonByte)
						//update loginUser
						saveContactErr = tx.Save(account).Error
					}

					if saveContactErr != nil {
						logpack.Warn("Save to contact failed. Transfer is continue", utils.GetFuncName())
					}
				}
				addedContacts = true
			}
		}
	}
	//if transfer is cryptocurrency
	if utils.IsCryptoCurrency(currency) {
		if sendBy == "urlcode" {
			//if is urlcode, create new code and create data in DB
			urlCodeErr := s.H.HanlderWithdrawWithUrlCode(reqData.Common.LoginName, currency, amountToSend, note)
			if urlCodeErr != nil {
				tx.Rollback()
				return ResponseError(urlCodeErr.Error(), utils.GetFuncName(), urlCodeErr)
			} else {
				return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Create url code successfully", utils.GetFuncName(), map[string]bool{"addedContact": addedContacts})
			}
		}
		//if sendBy address, check address
		if sendBy == "address" {
			if utils.IsEmpty(address) {
				return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Address param failed. Please try again!", utils.GetFuncName(), nil)
			}
			isSystemAddress := false
			//check to address is of system address
			addressObj, addrErr := s.H.GetAddress(address)
			if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
				return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "DB error. Please try again!", utils.GetFuncName(), addrErr)
			}
			var existAsset *models.Asset
			if addressObj != nil {
				var assetErr error
				existAsset, assetErr = s.H.GetAssetById(addressObj.AssetId)
				if assetErr != nil && assetErr != gorm.ErrRecordNotFound {
					return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "DB error. Please try again!", utils.GetFuncName(), assetErr)
				}
				//if exist user for address
				if existAsset != nil {
					isSystemAddress = true
				}
			}
			if !isSystemAddress {
				_, btcHandlerErr := s.H.HandlerTransferOnchainCryptocurrency(reqData.Common.LoginName, currency, address, amountToSend, rateSend, note)
				if btcHandlerErr != nil {
					return ResponseLoginRollbackError(reqData.Common.LoginName, tx, btcHandlerErr.Error(), utils.GetFuncName(), btcHandlerErr)
				} else {
					tx.Commit()
					return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Successfully performed transfer", utils.GetFuncName(), map[string]bool{"addedContact": addedContacts})
				}
			} else {
				sendBy = "username"
				receiverUsername = existAsset.UserName
			}
		}
	}

	//get assets of sender
	assetObj := assets.StringToAssetType(currency)
	senderAsset, senderAssetErr := s.H.GetUserAsset(reqData.Common.LoginName, currency)
	if senderAssetErr != nil || senderAsset == nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Error getting Asset data from DB or sender asset does not exist. Please try again!", utils.GetFuncName(), nil)
	}
	//if balance less than amount to send, return error
	if senderAsset.Balance < amountToSend {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Insufficient balance. Please check again or deposit more balance", utils.GetFuncName(), nil)
	}
	//Deduct money from balance and update local transfer total
	senderAsset.Balance -= amountToSend
	senderAsset.LocalSent += amountToSend

	//get assets of receiver
	receiverAsset, receiverAssetErr := s.H.GetUserAsset(receiverUsername, currency)
	if receiverAssetErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Retrieve recipient asset data failed. Please try again!", utils.GetFuncName(), receiverAssetErr)
	}
	//update sender asset
	senderAssetUpdateErr := tx.Save(senderAsset).Error
	if senderAssetUpdateErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update Sender failed. Please try again!", utils.GetFuncName(), senderAssetUpdateErr)
	}
	receiverAssetCreate := receiverAsset == nil

	//if receiver create asset
	if receiverAssetCreate {
		var newReceiverAsset *models.Asset
		var newErr error
		if currency == assets.USDWalletAsset.String() {
			newReceiverAsset, newErr = s.H.CreateNewUSDAsset(receiverUsername, utils.IsSuperAdmin(int(receiverRole)), assetObj)
			if newErr != nil {
				return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Create new USD asset and address failed. Please check again!", utils.GetFuncName(), newErr)
			}
		} else {
			_, newReceiverAsset, newErr = s.H.CreateNewAddressForAsset(receiverUsername, utils.IsSuperAdmin(int(receiverRole)), assetObj)
			if newErr != nil {
				return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Create new asset and address failed. Please check again!", utils.GetFuncName(), newErr)
			}
		}

		newReceiverAsset.Balance = amountToSend
		newReceiverAsset.LocalReceived = amountToSend
		newReceiverAsset.Updatedt = time.Now().Unix()

		receiverAssetUpdateErr := tx.Save(newReceiverAsset).Error
		if receiverAssetUpdateErr != nil {
			return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update balance for asset failed. Please check again!", utils.GetFuncName(), receiverAssetUpdateErr)
		}
	} else {
		//update receiver asset
		receiverAsset.Balance += amountToSend
		receiverAsset.LocalReceived += amountToSend
		receiverAssetUpdateErr := tx.Save(receiverAsset).Error
		if receiverAssetUpdateErr != nil {
			return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update recipient assets failed. Please try again!", utils.GetFuncName(), receiverAssetUpdateErr)
		}
	}

	//insert to transaction history
	txHistory := models.TxHistory{}
	txHistory.Sender = reqData.Common.LoginName
	txHistory.Receiver = receiverUsername
	txHistory.Currency = currency
	txHistory.Amount = amountToSend
	txHistory.Status = 1
	txHistory.Description = note
	txHistory.Createdt = time.Now().UnixNano() / 1e9
	txHistory.TransType = int(utils.TransTypeLocal)
	txHistory.Rate = rateSend

	HistoryErr := tx.Create(&txHistory).Error
	if HistoryErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Recorded history is corrupted. Please check your balance again!", utils.GetFuncName(), HistoryErr)
	}
	tx.Commit()
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Money transfer successful", utils.GetFuncName(), map[string]bool{"addedContact": addedContacts})
}

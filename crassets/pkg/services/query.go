package services

import (
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
	pb.UnimplementedChatServiceServer
}

func (s *Server) CreateNewAddress(reqData *pb.RequestData) *pb.ResponseData {
	//get selected type
	assetTypeAny, assetTypeExist := reqData.DataMap["assetType"]
	if !assetTypeExist {
		return pb.ResponseError("Get Asset Type param failed. Please try again!", utils.GetFuncName(), nil)
	}
	selectedType := assetTypeAny.(string)
	assetObject := assets.StringToAssetType(selectedType)
	address, _, createErr := s.CreateNewAddressForAsset(reqData.LoginId, reqData.LoginName, utils.IsSuperAdmin(reqData.Role), assetObject)
	if createErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, createErr.Error(), utils.GetFuncName(), nil)
	}

	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Create new address successfully", utils.GetFuncName(), address.Address)
}

func (s *Server) SyncTransactions(reqData *pb.RequestData) *pb.ResponseData {
	if reqData.Role != int(utils.RoleSuperAdmin) {
		return pb.ResponseError("Check admin login failed", utils.GetFuncName(), nil)
	}
	//create task to sync transactions
	go func() {
		s.SystemSyncHandler()
	}()
	return pb.ResponseSuccessfully(reqData.LoginId, "Synchronized transaction successfully", utils.GetFuncName())
}

func (s *Server) SendTradingRequest(reqData *pb.RequestData) *pb.ResponseData {
	assetAny, assetExist := reqData.DataMap["asset"]
	tradingTypeAny, tradingTypeExist := reqData.DataMap["tradingType"]
	paymentTypeAny, paymentTypeExist := reqData.DataMap["paymentType"]
	amountAny, amountExist := reqData.DataMap["amount"]
	rateAny, rateExist := reqData.DataMap["rate"]

	if !assetExist || !tradingTypeExist || !paymentTypeExist || !amountExist || !rateExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	asset := assetAny.(string)
	tradingType := tradingTypeAny.(string)
	amount := amountAny.(float64)
	paymentType := paymentTypeAny.(string)
	rate := rateAny.(float64)
	paymentAmount := amount * rate

	if tradingType != utils.TradingTypeBuy && tradingType != utils.TradingTypeSell {
		return pb.ResponseLoginError(reqData.LoginId, "Trading Type param failed. Please check again!", utils.GetFuncName(), nil)
	}

	//get asset of system user
	systemAsset, systemAssetErr := s.H.GetSystemUserAsset(asset)
	if systemAssetErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get system user asset failed. Please check again!", utils.GetFuncName(), systemAssetErr)
	}

	//Get asset of payment type for system user
	systemPaymentAsset, paymentErr := s.H.GetSystemUserAsset(paymentType)
	if paymentErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get system user payment asset failed. Please check again!", utils.GetFuncName(), paymentErr)
	}

	//get asset of loginUser
	loginAsset, loginAssetErr := s.H.GetUserAsset(reqData.LoginName, asset)
	if loginAssetErr != nil || loginAsset == nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get login user asset failed. Please check again!", utils.GetFuncName(), loginAssetErr)
	}

	//Get asset of payment type for loginuser
	loginPaymentAsset, loginPaymentAssetErr := s.H.GetUserAsset(reqData.LoginName, paymentType)
	if loginPaymentAssetErr != nil || loginPaymentAsset == nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get login user payment asset failed. Please check again!", utils.GetFuncName(), loginPaymentAssetErr)
	}
	now := time.Now().Unix()

	//check valid balance of asset and payment asset
	//if payment type is buy
	if tradingType == utils.TradingTypeBuy {
		//check balance of payment asset of login user
		if paymentAmount > loginPaymentAsset.Balance {
			//return error
			return pb.ResponseLoginError(reqData.LoginId, fmt.Sprintf("%s balance is not enough to buy this amount", strings.ToUpper(paymentType)), utils.GetFuncName(), nil)
		}
		//check balance of system user of asset
		if amount > systemAsset.Balance {
			return pb.ResponseLoginError(reqData.LoginId, fmt.Sprintf("%s balance of system user is not enough to sell this amount", strings.ToUpper(asset)), utils.GetFuncName(), nil)
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
			return pb.ResponseLoginError(reqData.LoginId, fmt.Sprintf("%s balance is not enough to sell this amount", strings.ToUpper(asset)), utils.GetFuncName(), nil)
		}
		//check balance of system user of asset
		if paymentAmount > systemPaymentAsset.Balance {
			return pb.ResponseLoginError(reqData.LoginId, fmt.Sprintf("%s balance of system user is not enough to buy this amount", strings.ToUpper(paymentType)), utils.GetFuncName(), nil)
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
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update data for assets failed", utils.GetFuncName(), nil)
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
		Sender:      reqData.LoginName,
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
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Insert transaction history failed", utils.GetFuncName(), txHistoryInsertErr)
	}
	tx.Commit()
	//return successfully response
	return pb.ResponseSuccessfully(reqData.LoginId, "Transaction completed", utils.GetFuncName())
}

func (s *Server) ConfirmWithdrawal(reqData *pb.RequestData) *pb.ResponseData {
	targetAny, targetExist := reqData.DataMap["target"]
	codeAny, codeExist := reqData.DataMap["code"]
	if !targetExist || !codeExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	target := targetAny.(string)
	code := codeAny.(string)
	//Check valid token and get user of token
	txCode, exist := s.H.GetTxcode(code)
	if !exist {
		return pb.ResponseError("Retrieve TxCode data error", utils.GetFuncName(), nil)
	}

	//current rate
	rateSend := s.H.GetRateFromDBByAsset(txCode.Asset)
	txCode.Note = fmt.Sprintf("%s: Withdraw with URL Code", txCode.Note)
	//Get address
	if utils.IsEmpty(target) {
		return pb.ResponseError("Address cannot be left blank. Please check again!", utils.GetFuncName(), nil)
	}
	isSystemAddress := false
	//check to address is of system address
	addressObj, addrErr := s.H.GetAddress(target)
	if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
		return pb.ResponseError("DB error. Please try again!", utils.GetFuncName(), addrErr)
	}
	var existAsset *models.Asset
	if addressObj != nil {
		var assetErr error
		existAsset, assetErr = s.H.GetAssetById(addressObj.AssetId)
		if assetErr != nil && assetErr != gorm.ErrRecordNotFound {
			return pb.ResponseError("DB error. Please try again!", utils.GetFuncName(), assetErr)
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
			return pb.ResponseError("Create RPC Client failed", utils.GetFuncName(), nil)
		}
		validAddress := assetObj.IsValidAddress(target)
		if !validAddress {
			return pb.ResponseError("Invalid address", utils.GetFuncName(), nil)
		}
		txHistory, btcHandlerErr := s.H.HandlerTransferOnchainCryptocurrency(sender.Username, txCode.Asset, target, txCode.Amount, rateSend, txCode.Note)
		if btcHandlerErr != nil {
			return pb.ResponseError(btcHandlerErr.Error(), utils.GetFuncName(), nil)
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
			return pb.ResponseError("Update Tx Code status failed!", utils.GetFuncName(), txUpdateErr)
		}
		tx.Commit()
		return pb.ResponseSuccessfully(0, "Confirm withdrawl successfully", utils.GetFuncName())
	} else {
		user := &models.UserInfo{
			Username: existAsset.UserName,
		}
		completeInternal := s.H.HandlerInternalWithdrawl(txCode, *user, rateSend)
		if !completeInternal {
			return pb.ResponseError("Hanlder internal withdrawl failed", utils.GetFuncName(), nil)
		}
		return pb.ResponseSuccessfully(0, "Withdrawl with internal account successfully", utils.GetFuncName())
	}
}

func (s *Server) UpdateNewLabel(reqData *pb.RequestData) *pb.ResponseData {
	assetIdAny, assetIdExist := reqData.DataMap["assetId"]
	addressIdAny, addressIdExist := reqData.DataMap["addressId"]
	newLabelAny, newLabelExist := reqData.DataMap["newMainLabel"]
	assetTypeAny, assetTypeExist := reqData.DataMap["assetType"]

	if !assetIdExist || !addressIdExist || !newLabelExist || !assetTypeExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetIdAny.(int64)
	addressId := addressIdAny.(int64)
	newMainLabel := newLabelAny.(string)
	assetType := assetTypeAny.(string)

	if !s.H.CheckMatchAddressWithUser(assetId, addressId, reqData.LoginName, false) {
		return pb.ResponseLoginError(reqData.LoginId, "The user login information and assets do not match", utils.GetFuncName(), nil)
	}

	s.UpdateAssetManagerByType(assetType)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !assetMgrExist {
		return pb.ResponseLoginError(reqData.LoginId, "Create RPC Client failed!", utils.GetFuncName(), nil)
	}

	//get address
	address, addressErr := s.H.GetAddressById(addressId)
	if addressErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get address from DB failed", utils.GetFuncName(), addressErr)
	}

	tx := s.H.DB
	//get token of loginuser
	token := s.H.GetTokenFromUsername(reqData.LoginName)
	if utils.IsEmpty(token) {
		return pb.ResponseLoginError(reqData.LoginId, "Get user token failed", utils.GetFuncName(), nil)
	}
	//set new label for address on DB
	newLabel := fmt.Sprintf("%s_%s", token, newMainLabel)
	address.Label = newLabel
	addressUpdateErr := tx.Save(address).Error
	if addressUpdateErr != nil {
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update address label from DB failed", utils.GetFuncName(), addressUpdateErr)
	}
	//update label on daemon
	daemonUpdateErr := assetObj.UpdateLabel(address.Address, newLabel)
	if daemonUpdateErr != nil {
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update address label on daemon failed", utils.GetFuncName(), daemonUpdateErr)
	}
	tx.Commit()
	return pb.ResponseSuccessfully(reqData.LoginId, "Update address label successfully", utils.GetFuncName())
}

func (s *Server) GetAddressListData(reqData *pb.RequestData) *pb.ResponseData {
	assetIdAny, assetIdExist := reqData.DataMap["assetId"]
	statusAny, statusExist := reqData.DataMap["status"]
	if !assetIdExist || !statusExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetIdAny.(int64)
	status := statusAny.(string)
	//check asset id match with loginUser
	assetMatch := s.H.CheckAssetMatchWithUser(assetId, reqData.LoginName)
	if !assetMatch {
		return pb.ResponseError("asset not match with login user", utils.GetFuncName(), fmt.Errorf("asset not match with login user"))
	}
	//Get url code list
	addressList, addressErr := s.H.FilterAddressList(assetId, status)
	if addressErr != nil {
		return pb.ResponseError(addressErr.Error(), utils.GetFuncName(), addressErr)
	}
	//user token
	token, _, err := s.H.CheckAndCreateAccountToken(reqData.LoginName, reqData.Role)
	if err != nil {
		return pb.ResponseError("Check or create user token failed", utils.GetFuncName(), fmt.Errorf("Check or create user token failed"))
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
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get Address List data successfully", utils.GetFuncName(), resultMap)
}
func (s *Server) GetCodeListData(reqData *pb.RequestData) *pb.ResponseData {
	assetAny, assetExist := reqData.DataMap["asset"]
	codeStatusAny, codeStatusExist := reqData.DataMap["codeStatus"]
	if !assetExist || !codeStatusExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetType := assetAny.(string)
	status := codeStatusAny.(string)
	//Get url code list
	urlCodeList, urlCodeErr := s.H.FilterUrlCodeList(assetType, status, reqData.LoginName)
	if urlCodeErr != nil {
		return pb.ResponseError(urlCodeErr.Error(), utils.GetFuncName(), urlCodeErr)
	}
	urlCodeDisplayList := make([]*models.TxCodeDisplay, 0)
	for _, urlCode := range urlCodeList {
		urlCodeStatus, statusErr := utils.GetUrlCodeStatusFromValue(urlCode.Status)
		if statusErr != nil {
			logpack.FError(statusErr.Error(), reqData.LoginId, utils.GetFuncName(), nil)
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
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get Code List data successfully", utils.GetFuncName(), resultMap)
}

func (s *Server) FilterTxHistory(reqData *pb.RequestData) *pb.ResponseData {
	allowassetsAny, allowassetsExist := reqData.DataMap["allowassets"]
	typeAny, typeExist := reqData.DataMap["type"]
	directionAny, directionExist := reqData.DataMap["direction"]
	perpageAny, perpageExist := reqData.DataMap["perpage"]
	pageNumAny, pageNumExist := reqData.DataMap["pageNum"]
	var perpage int64
	var pageNum int64
	if !perpageExist {
		perpage = 15
	} else {
		perpage = perpageAny.(int64)
	}
	if !pageNumExist {
		pageNum = 1
	} else {
		pageNum = pageNumAny.(int64)
	}
	if !allowassetsExist || !typeExist || !directionExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	allowAssetStr := allowassetsAny.(string)
	assetType := typeAny.(string)
	direction := directionAny.(string)
	allowAssets := utils.GetAssetsNameFromStr(allowAssetStr)
	if assetType == "all" {
		assetType = ""
	}

	if direction == "all" {
		direction = ""
	}

	txHistoryList, pageCount := s.H.InitTransactionHistoryList(&models.UserInfo{
		Username: reqData.LoginName,
	}, assetType, direction, perpage, pageNum, allowAssets)
	resultMap := make(map[string]any)
	resultMap["pageCount"] = pageCount
	resultMap["list"] = txHistoryList
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get Tx History List successfully", utils.GetFuncName(), resultMap)
}

func (s *Server) ConfirmAmount(reqData *pb.RequestData) *pb.ResponseData {
	assetAny, assetExist := reqData.DataMap["asset"]
	toaddressAny, toaddressExist := reqData.DataMap["toaddress"]
	sendByAny, sendByExist := reqData.DataMap["sendBy"]
	amountAny, amountExist := reqData.DataMap["amount"]
	if !assetExist || !toaddressExist || !sendByExist || !amountExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}

	asset := assetAny.(string)
	address := toaddressAny.(string)
	sendBy := sendByAny.(string)
	amountToSend := amountAny.(float64)
	s.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return pb.ResponseLoginError(reqData.LoginId, "Create RPC Client failed!", utils.GetFuncName(), nil)
	}
	if utils.IsEmpty(address) {
		//if has no address from param. Get address of system admin to check
		address = assetObj.GetSystemAddress()
		if utils.IsEmpty(address) {
			return pb.ResponseLoginError(reqData.LoginId, "Address param failed. Please try again!", utils.GetFuncName(), nil)
		}
	}

	//check to address is of system address
	addressObj, addrErr := s.H.GetAddress(address)
	if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
		return pb.ResponseLoginError(reqData.LoginId, "DB error. Please try again!", utils.GetFuncName(), addrErr)
	}

	if addressObj != nil && sendBy != "urlcode" {
		existAsset, assetErr := s.H.GetAssetById(addressObj.AssetId)
		if assetErr != nil && assetErr != gorm.ErrRecordNotFound {
			return pb.ResponseLoginError(reqData.LoginId, "DB error. Please try again!", utils.GetFuncName(), addrErr)
		}
		//if exist user for address
		if existAsset != nil {
			return pb.ResponseLoginErrorWithCode(reqData.LoginId, "exist", fmt.Sprintf("<span>Is the user's address: <span class=\"fw-600 fs-16\">%s</span>. You'll not be charged transaction fees</span>", existAsset.UserName), utils.GetFuncName(), nil)
		}
	}
	//check valid address
	validAddress := utils.CheckValidAddress(asset, address)
	if !validAddress {
		return pb.ResponseLoginError(reqData.LoginId, "Invalid address. Please check again!", utils.GetFuncName(), nil)
	}
	//Get login user asset
	loginAsset, loginAssetErr := s.H.GetUserAsset(reqData.LoginName, asset)
	if loginAssetErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get Asset of loginUser failed. Please check again!", utils.GetFuncName(), loginAssetErr)
	}

	fromAddressList, addrListErr := s.H.GetAddressListByAssetId(loginAsset.Id)
	if addrListErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get address list of loginUser failed or address list empty. Please check again!", utils.GetFuncName(), addrListErr)
	}
	//start estimate fee and size
	unitAmount, unitAmountErr := utils.GetUnitAmount(amountToSend, asset)
	if unitAmountErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get UnitAmount from amount failed!", utils.GetFuncName(), unitAmountErr)
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()
	feeAndSize, err := assetObj.EstimateFeeAndSize(&assets.TxTarget{
		FromAddresses: fromAddressList,
		ToAddress:     address,
		Amount:        amountToSend,
		UnitAmount:    unitAmount,
		Account:       reqData.LoginName,
	})
	if err != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get Estimate fee and size failed!", utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Confirm amount successfully", utils.GetFuncName(), fmt.Sprintf("%f", feeAndSize.Fee.CoinValue))
}

func (s *Server) AddToContact(reqData *pb.RequestData) *pb.ResponseData {
	currentContacts, contactErr := s.H.GetContactListFromUser(reqData.LoginName)
	if contactErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Parse contact of loginUser failed", utils.GetFuncName(), contactErr)
	}
	targetNameAny, targetNameExist := reqData.DataMap["targetName"]
	if !targetNameExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	targetName := targetNameAny.(string)
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
			return pb.ResponseLoginError(reqData.LoginId, "Parse contacts failed", utils.GetFuncName(), err)
		}
		if err == nil {
			contactStr := string(jsonByte)
			updateErr := s.H.UpdateUserContacts(reqData.LoginName, contactStr)
			if updateErr != nil {
				return pb.ResponseLoginError(reqData.LoginId, "Update user contacts failed", utils.GetFuncName(), updateErr)
			}
		}
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Update contacts successfully", utils.GetFuncName(), isExist)
}

func (s *Server) TransferAmount(reqData *pb.RequestData) *pb.ResponseData {
	currencyAny, currencyExist := reqData.DataMap["currency"]
	amountAny, amountExist := reqData.DataMap["amount"]
	receiverAny, receiverExist := reqData.DataMap["receiver"]
	receiverRoleAny, receiverRoleExist := reqData.DataMap["receiverRole"]
	rateAny, rateExist := reqData.DataMap["rate"]
	noteAny, noteExist := reqData.DataMap["note"]
	sendByAny, sendByExist := reqData.DataMap["sendBy"]
	addressAny, addressExist := reqData.DataMap["address"]

	if !currencyExist || !amountExist || !receiverExist || !rateExist || !noteExist || !sendByExist || !addressExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}

	receiverRole := int(utils.RoleRegular)
	if receiverRoleExist {
		receiverRole = receiverRoleAny.(int)
	}

	currency := currencyAny.(string)
	amountToSend := amountAny.(float64)
	receiverUsername := receiverAny.(string)
	rateSend := rateAny.(float64)
	note := noteAny.(string)
	sendBy := sendByAny.(string)
	address := addressAny.(string)
	tx := s.H.DB.Begin()
	if currency == assets.USDWalletAsset.String() || (currency != assets.USDWalletAsset.String() && sendBy == "username") {
		if utils.IsEmpty(receiverUsername) {
			return pb.ResponseLoginError(reqData.LoginId, "Recipient information cannot be left blank. Please try again!", utils.GetFuncName(), nil)
		}
	}
	//if transfer is cryptocurrency
	if utils.IsCryptoCurrency(currency) {
		if sendBy == "urlcode" {
			//if is urlcode, create new code and create data in DB
			urlCodeErr := s.H.HanlderWithdrawWithUrlCode(reqData.LoginName, currency, amountToSend, note)
			if urlCodeErr != nil {
				return pb.ResponseError(urlCodeErr.Error(), utils.GetFuncName(), urlCodeErr)
			} else {
				return pb.ResponseSuccessfully(reqData.LoginId, "Create url code successfully", utils.GetFuncName())
			}
		}
		//if sendBy address, check address
		if sendBy == "address" {
			if utils.IsEmpty(address) {
				return pb.ResponseLoginError(reqData.LoginId, "Address param failed. Please try again!", utils.GetFuncName(), nil)
			}
			isSystemAddress := false
			//check to address is of system address
			addressObj, addrErr := s.H.GetAddress(address)
			if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
				return pb.ResponseLoginError(reqData.LoginId, "DB error. Please try again!", utils.GetFuncName(), addrErr)
			}
			var existAsset *models.Asset
			if addressObj != nil {
				var assetErr error
				existAsset, assetErr = s.H.GetAssetById(addressObj.AssetId)
				if assetErr != nil && assetErr != gorm.ErrRecordNotFound {
					return pb.ResponseLoginError(reqData.LoginId, "DB error. Please try again!", utils.GetFuncName(), assetErr)
				}
				//if exist user for address
				if existAsset != nil {
					isSystemAddress = true
				}
			}
			if !isSystemAddress {
				_, btcHandlerErr := s.H.HandlerTransferOnchainCryptocurrency(reqData.LoginName, currency, address, amountToSend, rateSend, note)
				if btcHandlerErr != nil {
					return pb.ResponseLoginError(reqData.LoginId, btcHandlerErr.Error(), utils.GetFuncName(), btcHandlerErr)
				} else {
					return pb.ResponseSuccessfully(reqData.LoginId, "Successfully performed transfer", utils.GetFuncName())
				}
			} else {
				sendBy = "username"
				receiverUsername = existAsset.UserName
			}
		}
	}

	//get assets of sender
	assetObj := assets.StringToAssetType(currency)
	senderAsset, senderAssetErr := s.H.GetUserAsset(reqData.LoginName, currency)
	if senderAssetErr != nil || senderAsset == nil {
		return pb.ResponseLoginError(reqData.LoginId, "Error getting Asset data from DB or sender asset does not exist. Please try again!", utils.GetFuncName(), nil)
	}
	//if balance less than amount to send, return error
	if senderAsset.Balance < amountToSend {
		return pb.ResponseLoginError(reqData.LoginId, "Insufficient balance. Please check again or deposit more balance", utils.GetFuncName(), nil)
	}
	//Deduct money from balance and update local transfer total
	senderAsset.Balance -= amountToSend
	senderAsset.LocalSent += amountToSend

	//get assets of receiver
	receiverAsset, receiverAssetErr := s.H.GetUserAsset(receiverUsername, currency)
	if receiverAssetErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Retrieve recipient asset data failed. Please try again!", utils.GetFuncName(), receiverAssetErr)
	}
	//update sender asset
	senderAssetUpdateErr := tx.Save(senderAsset).Error
	if senderAssetUpdateErr != nil {
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update Sender failed. Please try again!", utils.GetFuncName(), senderAssetUpdateErr)
	}
	receiverAssetCreate := receiverAsset == nil

	//if receiver create asset
	if receiverAssetCreate {
		_, newReceiverAsset, newErr := s.H.CreateNewAddressForAsset(receiverUsername, utils.IsSuperAdmin(receiverRole), assetObj)
		if newErr != nil {
			return pb.ResponseLoginError(reqData.LoginId, "Create new asset and address failed. Please check again!", utils.GetFuncName(), newErr)
		}
		newReceiverAsset.Balance = amountToSend
		newReceiverAsset.LocalReceived = amountToSend
		newReceiverAsset.Updatedt = time.Now().Unix()

		receiverAssetUpdateErr := tx.Save(newReceiverAsset).Error
		if receiverAssetUpdateErr != nil {
			return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update balance for asset failed. Please check again!", utils.GetFuncName(), receiverAssetUpdateErr)
		}
	} else {
		//update receiver asset
		receiverAsset.Balance += amountToSend
		receiverAsset.LocalReceived += amountToSend
		receiverAssetUpdateErr := tx.Save(receiverAsset).Error
		if receiverAssetUpdateErr != nil {
			return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update recipient assets failed. Please try again!", utils.GetFuncName(), receiverAssetUpdateErr)
		}
	}

	//insert to transaction history
	txHistory := models.TxHistory{}
	txHistory.Sender = reqData.LoginName
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
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Recorded history is corrupted. Please check your balance again!", utils.GetFuncName(), HistoryErr)
	}
	tx.Commit()
	return pb.ResponseSuccessfully(reqData.LoginId, "Money transfer successful", utils.GetFuncName())
}

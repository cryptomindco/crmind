package controllers

import (
	"crassets/handler"
	"crassets/logpack"
	"crassets/models"
	"crassets/utils"
	"crassets/walletlib/assets"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type TransferController struct {
	BaseController
}

func (this *TransferController) ConfirmWithdrawal() {
	target := strings.TrimSpace(this.GetString("target"))
	code := strings.TrimSpace(this.GetString("code"))

	if utils.IsEmpty(target) || utils.IsEmpty(code) {
		this.ResponseError("Param failed", utils.GetFuncName(), nil)
		return
	}

	o := orm.NewOrm()
	//Check valid token and get user of token
	txCode, exist := utils.GetTxcode(code)
	if !exist {
		this.ResponseError("Retrieve TxCode data error", utils.GetFuncName(), nil)
		return
	}

	//current rate
	rateSend := utils.GetRateFromDBByAsset(txCode.Asset)
	txCode.Note = fmt.Sprintf("%s: Withdraw with URL Code", txCode.Note)
	//Get address
	if utils.IsEmpty(target) {
		this.ResponseError("Address cannot be left blank. Please check again!", utils.GetFuncName(), nil)
		return
	}
	isSystemAddress := false
	//check to address is of system address
	addressObj, addrErr := utils.GetAddress(target)
	if addrErr != nil && addrErr != orm.ErrNoRows {
		this.ResponseError("DB error. Please try again!", utils.GetFuncName(), addrErr)
		return
	}
	var existAsset *models.Asset
	if addressObj != nil {
		var assetErr error
		existAsset, assetErr = utils.GetAssetById(addressObj.AssetId)
		if assetErr != nil && assetErr != orm.ErrNoRows {
			this.ResponseError("DB error. Please try again!", utils.GetFuncName(), assetErr)
			return
		}
		//if exist user for address
		if existAsset != nil {
			isSystemAddress = true
		}
	}

	if !isSystemAddress {
		//change note of txCode
		//create sender
		sender := &models.AuthClaims{
			Id:       txCode.OwnerId,
			Username: txCode.OwnerName,
		}
		//Check RPC client
		handler.UpdateAssetManagerByType(txCode.Asset)
		assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[txCode.Asset]
		if !assetMgrExist {
			this.ResponseError("Create RPC Client failed", utils.GetFuncName(), nil)
			return
		}
		validAddress := assetObj.IsValidAddress(target)
		if !validAddress {
			this.ResponseError("Invalid address", utils.GetFuncName(), nil)
			return
		}
		txHistory, btcHandlerErr := this.HandlerTransferOnchainCryptocurrency(o, sender, txCode.Asset, target, txCode.Amount, rateSend, txCode.Note)
		if btcHandlerErr != nil {
			this.ResponseError(btcHandlerErr.Error(), utils.GetFuncName(), nil)
		} else {
			logpack.Info("Successfully performed transfer", utils.GetFuncName())
			this.Data["json"] = map[string]string{"error": ""}
		}
		//update txCode status
		txCode.Status = int(utils.UrlCodeStatusConfirmed)
		txCode.Txid = txHistory.Txid
		txCode.HistoryId = txHistory.Id
		txCode.Confirmdt = time.Now().Unix()
		tx, beginErr := o.Begin()
		if beginErr != nil {
			this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
			return
		}
		_, txUpdateErr := tx.Update(txCode)
		if txUpdateErr != nil {
			this.ResponseRollbackError(tx, "Update Tx Code status failed!", utils.GetFuncName(), txUpdateErr)
			return
		}
		tx.Commit()
		this.ServeJSON()
		return
	} else {
		user := &models.AuthClaims{
			Id:       existAsset.UserId,
			Username: existAsset.UserName,
		}
		this.HandlerInternalWithdrawl(txCode, *user, rateSend, o)
	}
}

func (this *TransferController) UpdateNewLabel() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetId, parseAssetErr := this.GetInt64("assetId")
	addressId, parseAddrErr := this.GetInt64("addressId")
	if parseAddrErr != nil || parseAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Check param error. Please try again!", utils.GetFuncName(), nil)
		return
	}

	if !utils.CheckMatchAddressWithUser(assetId, addressId, loginUser.Id, false) {
		this.ResponseLoginError(loginUser.Id, "The user login information and assets do not match", utils.GetFuncName(), nil)
		return
	}

	newMainLabel := this.GetString("newMainLabel")
	assetType := this.GetString("assetType")
	if utils.IsEmpty(newMainLabel) {
		this.ResponseLoginError(loginUser.Id, "New label cannot be empty", utils.GetFuncName(), nil)
		return
	}

	handler.UpdateAssetManagerByType(assetType)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !assetMgrExist {
		this.ResponseLoginError(loginUser.Id, "Create RPC Client failed!", utils.GetFuncName(), nil)
		return
	}

	//get address
	address, addressErr := utils.GetAddressById(addressId)
	if addressErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get address from DB failed", utils.GetFuncName(), addressErr)
		return
	}

	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	//get token of loginuser
	token := utils.GetTokenFromUserId(loginUser.Id)
	if utils.IsEmpty(token) {
		this.ResponseLoginError(loginUser.Id, "Get user token failed", utils.GetFuncName(), nil)
		return
	}
	//set new label for address on DB
	newLabel := fmt.Sprintf("%s_%s", token, newMainLabel)
	address.Label = newLabel
	_, addressUpdateErr := tx.Update(address)
	if addressUpdateErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Update address label from DB failed", utils.GetFuncName(), addressUpdateErr)
		return
	}
	//update label on daemon
	daemonUpdateErr := assetObj.UpdateLabel(address.Address, newLabel)
	if daemonUpdateErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Update address label on daemon failed", utils.GetFuncName(), daemonUpdateErr)
		return
	}
	tx.Commit()
	this.ResponseSuccessfully(loginUser.Id, "Update address label successfully", utils.GetFuncName())
}

func (this *TransferController) GetAddressListData() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetId, assetErr := this.GetInt64("assetId")
	if assetErr != nil {
		this.ResponseError(assetErr.Error(), utils.GetFuncName(), assetErr)
		return
	}

	//check asset id match with loginUser
	assetMatch := utils.CheckAssetMatchWithUser(assetId, loginUser.Id)
	if !assetMatch {
		this.ResponseError("asset not match with login user", utils.GetFuncName(), fmt.Errorf("asset not match with login user"))
		return
	}
	status := strings.TrimSpace(this.GetString("status"))
	//Get url code list
	addressList, addressErr := utils.FilterAddressList(assetId, status)
	if addressErr != nil {
		this.ResponseError(addressErr.Error(), utils.GetFuncName(), addressErr)
		return
	}
	//user token
	token, _, tokenErr := utils.CheckAndCreateUserToken(*loginUser)
	if tokenErr != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
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
	this.Data["json"] = resultMap
	this.ServeJSON()
}

func (this *TransferController) GetCodeListData() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetType := strings.TrimSpace(this.GetString("asset"))
	status := strings.TrimSpace(this.GetString("codeStatus"))

	//Get url code list
	urlCodeList, urlCodeErr := utils.FilterUrlCodeList(assetType, status, loginUser.Id)
	if urlCodeErr != nil {
		this.ResponseError(urlCodeErr.Error(), utils.GetFuncName(), urlCodeErr)
		return
	}
	urlCodeDisplayList := make([]*models.TxCodeDisplay, 0)
	for _, urlCode := range urlCodeList {
		urlCodeStatus, statusErr := utils.GetUrlCodeStatusFromValue(urlCode.Status)
		if statusErr != nil {
			logpack.FError(statusErr.Error(), loginUser.Id, utils.GetFuncName(), nil)
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
			history, err := utils.GetTxHistoryById(urlCode.HistoryId)
			if err == nil {
				txCodeDisp.TxHistory = history
			}

		}
		urlCodeDisplayList = append(urlCodeDisplayList, txCodeDisp)
	}

	resultMap := make(map[string]any)
	resultMap["list"] = urlCodeDisplayList
	this.Data["json"] = resultMap
	this.ServeJSON()
}

func (this *TransferController) GetLastTxs() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	limit := strings.TrimSpace(this.GetString("limit"))
	limitNum, parseErr := strconv.ParseInt(limit, 0, 32)
	if parseErr != nil {
		limitNum = 5
	}
	//get lastest 5 tx
	queryLastTxs := fmt.Sprintf("SELECT * from %stx_history WHERE sender_id = %d OR receiver_id = %d ORDER BY createdt DESC LIMIT %d", utils.GetAssetRelatedTablePrefix(), loginUser.Id, loginUser.Id, limitNum)
	var lastTxList []models.TxHistory
	o := orm.NewOrm()
	o.Raw(queryLastTxs).QueryRows(&lastTxList)
	historyDispList := make([]models.TxHistoryDisplay, 0)
	for _, txHistory := range lastTxList {
		createDt := time.Unix(txHistory.Createdt, 0)
		typeDisplay := utils.GetTransTypeFromValue(txHistory.TransType).ToString()
		historyDisp := models.TxHistoryDisplay{
			TxHistory:    txHistory,
			IsSender:     loginUser.Id == txHistory.SenderId,
			RateValue:    txHistory.Rate * txHistory.Amount,
			TypeDisplay:  typeDisplay,
			IsOffChain:   txHistory.TransType == int(utils.TransTypeLocal),
			CreatedtDisp: createDt.Format("2006/01/02, 15:04:05"),
		}
		//handler number of needed confirmations
		if txHistory.Currency == assets.DCRWalletAsset.String() {
			historyDisp.ConfirmationNeed = 2
		} else {
			historyDisp.ConfirmationNeed = 6
		}
		historyDispList = append(historyDispList, historyDisp)
	}

	this.Data["json"] = historyDispList
	this.ServeJSON()
}

func (this *TransferController) FilterTxHistory() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetType := strings.TrimSpace(this.GetString("type"))
	direction := strings.TrimSpace(this.GetString("direction"))
	perpageStr := strings.TrimSpace(this.GetString("perpage"))
	pageNumStr := strings.TrimSpace(this.GetString("pageNum"))
	perpage, parseErr := strconv.ParseInt(perpageStr, 0, 32)
	pageNum, parsePageNumErr := strconv.ParseInt(pageNumStr, 0, 32)

	if parseErr != nil {
		perpage = 15
	}

	if parsePageNumErr != nil {
		pageNum = 1
	}

	if assetType == "all" {
		assetType = ""
	}

	if direction == "all" {
		direction = ""
	}

	txHistoryList, pageCount := this.InitTransactionHistoryList(loginUser, assetType, direction, perpage, pageNum)
	resultMap := make(map[string]any)
	resultMap["pageCount"] = pageCount
	resultMap["list"] = txHistoryList
	this.Data["json"] = resultMap
	this.ServeJSON()
}

func (this *TransferController) ConfirmAmount() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	asset := strings.TrimSpace(this.GetString("asset"))
	if utils.IsEmpty(asset) {
		this.ResponseLoginError(loginUser.Id, "Asset param failed. Please try again!", utils.GetFuncName(), nil)
		return
	}
	handler.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		this.ResponseLoginError(loginUser.Id, "Create RPC Client failed!", utils.GetFuncName(), nil)
		return
	}
	address := strings.TrimSpace(this.GetString("toaddress"))
	if utils.IsEmpty(address) {
		//if has no address from param. Get address of system admin to check
		address = assetObj.GetSystemAddress()
		if utils.IsEmpty(address) {
			this.ResponseLoginError(loginUser.Id, "Address param failed. Please try again!", utils.GetFuncName(), nil)
			return
		}
	}

	//check to address is of system address
	addressObj, addrErr := utils.GetAddress(address)
	if addrErr != nil && addrErr != orm.ErrNoRows {
		this.ResponseLoginError(loginUser.Id, "DB error. Please try again!", utils.GetFuncName(), addrErr)
		return
	}

	sendBy := strings.TrimSpace(this.GetString("sendBy"))

	if addressObj != nil && sendBy != "urlcode" {
		existAsset, assetErr := utils.GetAssetById(addressObj.AssetId)
		if assetErr != nil && assetErr != orm.ErrNoRows {
			this.ResponseLoginError(loginUser.Id, "DB error. Please try again!", utils.GetFuncName(), addrErr)
			return
		}
		//if exist user for address
		if existAsset != nil {
			this.ResponseLoginError(loginUser.Id, fmt.Sprintf("<span>Is the user's address: <span class=\"fw-600 fs-16\">%s</span>. You'll not be charged transaction fees</span>", existAsset.UserName), utils.GetFuncName(), nil)
			return
		}
	}

	amountStr := strings.TrimSpace(this.GetString("amount"))
	//check valid address
	validAddress := utils.CheckValidAddress(asset, address)
	if !validAddress {
		this.ResponseLoginError(loginUser.Id, "Invalid address. Please check again!", utils.GetFuncName(), nil)
		return
	}

	amountToSend, amountErr := strconv.ParseFloat(amountStr, 64)
	if amountErr != nil {
		this.ResponseLoginError(loginUser.Id, "Parse amount to float failed. Please check again!", utils.GetFuncName(), amountErr)
		return
	}
	//Get login user asset
	loginAsset, loginAssetErr := utils.GetUserAsset(loginUser.Id, asset)
	if loginAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get Asset of loginUser failed. Please check again!", utils.GetFuncName(), loginAssetErr)
		return
	}

	fromAddressList, addrListErr := this.GetAddressListByAssetId(loginAsset.Id)
	if addrListErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get address list of loginUser failed or address list empty. Please check again!", utils.GetFuncName(), addrListErr)
		return
	}
	//start estimate fee and size
	unitAmount, unitAmountErr := utils.GetUnitAmount(amountToSend, asset)
	if unitAmountErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get UnitAmount from amount failed!", utils.GetFuncName(), unitAmountErr)
		return
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()
	feeAndSize, err := assetObj.EstimateFeeAndSize(&assets.TxTarget{
		FromAddresses: fromAddressList,
		ToAddress:     address,
		Amount:        amountToSend,
		UnitAmount:    unitAmount,
		Account:       loginUser.Username,
	})
	if err != nil {
		this.ResponseLoginError(loginUser.Id, "Get Estimate fee and size failed!", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Confirm amount successfully", utils.GetFuncName(), fmt.Sprintf("%f", feeAndSize.Fee.CoinValue))
}

func (this *TransferController) ConfirmAddressAction() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetId, parseAssetErr := this.GetInt64("assetId")
	addressId, parseAddrErr := this.GetInt64("addressId")
	if parseAddrErr != nil || parseAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Check param error. Please try again!", utils.GetFuncName(), nil)
		return
	}
	action := this.GetString("action")
	//check valid assetId, addressId
	if !utils.CheckMatchAddressWithUser(assetId, addressId, loginUser.Id, action == "reuse") {
		this.ResponseLoginError(loginUser.Id, "The user login information and assets do not match", utils.GetFuncName(), nil)
		return
	}

	//Get address object
	address, addressErr := utils.GetAddressById(addressId)
	if addressErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get address from DB failed", utils.GetFuncName(), addressErr)
		return
	}
	address.Archived = action != "reuse"

	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	_, updateErr := tx.Update(address)
	if updateErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Update Address failed", utils.GetFuncName(), updateErr)
		return
	}
	tx.Commit()
	//return successfully
	this.ResponseSuccessfully(loginUser.Id, "Update address from DB successfully", utils.GetFuncName())
}

func (this *TransferController) CancelUrlCode() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	//get codeid
	codeIdStr := strings.TrimSpace(this.GetString("codeId"))
	codeId, err := strconv.ParseInt(codeIdStr, 0, 32)
	if err != nil {
		this.ResponseLoginError(loginUser.Id, "Parse codeId to cancel failed. Please try again!", utils.GetFuncName(), nil)
		return
	}
	cancelErr := utils.CancelTxCodeById(loginUser.Id, codeId)
	if cancelErr != nil {
		this.ResponseLoginError(loginUser.Id, cancelErr.Error(), utils.GetFuncName(), nil)
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Cancel Withdraw Code successfully!", utils.GetFuncName())
}

func (this *TransferController) TransferAmount() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	currency := strings.TrimSpace(this.GetString("currency"))
	amountStr := strings.TrimSpace(this.GetString("amount"))
	receiverUsername := strings.TrimSpace(this.GetString("receiver"))
	rate := strings.TrimSpace(this.GetString("rate"))
	note := strings.TrimSpace(this.GetString("note"))
	rateSend := float64(0)
	//if selected currency is empty, return error
	if utils.IsEmpty(currency) {
		this.ResponseLoginError(loginUser.Id, "Error getting property type parameter failed. Please check again!", utils.GetFuncName(), nil)
		return
	}

	amountToSend, amountErr := strconv.ParseFloat(amountStr, 64)
	//amount parse error
	if amountErr != nil {
		this.ResponseLoginError(loginUser.Id, "Parse amount to float failed. Please check again!", utils.GetFuncName(), amountErr)
		return
	}

	sendBy := strings.TrimSpace(this.GetString("sendBy"))
	var addToContact bool
	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	receiver := models.User{}
	if currency == assets.USDWalletAsset.String() || (currency != assets.USDWalletAsset.String() && sendBy == "username") {
		if utils.IsEmpty(receiverUsername) {
			this.ResponseLoginError(loginUser.Id, "Recipient information cannot be left blank. Please try again!", utils.GetFuncName(), nil)
			return
		}
		receiverErr := o.QueryTable(userModel).Filter("username", receiverUsername).Filter("status", int(utils.StatusActive)).Limit(1).One(&receiver)
		if receiverErr != nil {
			this.ResponseLoginError(loginUser.Id, "The recipient does not exist in the database. Please try again!", utils.GetFuncName(), receiverErr)
			return
		}
		// check and get contact
		var contactErr error
		addToContact, contactErr = this.GetBool("addToContact", false)
		if contactErr != nil {
			this.ResponseLoginError(loginUser.Id, "Get Add To Contact param failed. Please try again!", utils.GetFuncName(), contactErr)
			return
		}
		if addToContact {
			currentContacts, contactErr := utils.GetContactListFromUser(loginUser.Id)
			if contactErr != nil {
				this.ResponseLoginError(loginUser.Id, "Parse contact of loginUser failed", utils.GetFuncName(), contactErr)
				return
			}
			//check receiveruser exist on contact
			isExist := utils.CheckUserExistOnContactList(receiver.Id, currentContacts)
			if isExist {
				logpack.FError("The recipient already exists in contact", loginUser.Id, utils.GetFuncName(), nil)
			} else {
				//add to contact
				currentContacts = append(currentContacts, models.ContactItem{
					UserId:   receiver.Id,
					UserName: receiver.Username,
					Addeddt:  time.Now().Unix(),
				})
				jsonByte, err := json.Marshal(currentContacts)
				if err == nil {
					loginUser.Contacts = string(jsonByte)
					loginUser.Updatedt = time.Now().Unix()
					//update loginUser
					tx.Update(loginUser)
					// if loginUpdateErr == nil {
					// 	addedToContact = true
					// }

					//When add to contact, check and add chat
					//check if exist chat conversion
					chatExist, chatErr := utils.CheckExistChat(loginUser.Id, receiver.Id)
					if chatErr == nil && !chatExist {
						//Create new chat
						newChatMsg := &models.ChatMsg{
							FromId:   loginUser.Id,
							FromName: loginUser.Username,
							ToId:     receiver.Id,
							ToName:   receiver.Username,
							Createdt: time.Now().Unix(),
							Updatedt: time.Now().Unix(),
						}
						//insert new ChatMsg
						id, chatInsertErr := tx.Insert(newChatMsg)
						if chatInsertErr != nil {
							logpack.FError("Create new chat failed", loginUser.Username, utils.GetFuncName(), chatInsertErr)
						} else {
							helloChat := &models.ChatContent{
								ChatId:   id,
								UserId:   loginUser.Id,
								UserName: loginUser.Username,
								Content:  fmt.Sprintf("%s has added %s to contacts. Start chatting now", loginUser.Username, receiver.Username),
								IsHello:  true,
								Createdt: time.Now().Unix(),
							}
							//insert to chat content
							tx.Insert(helloChat)
						}
					}
				}
			}
		}
	}
	//if transfer is cryptocurrency
	if this.IsCryptoCurrency(currency) {
		address := strings.TrimSpace(this.GetString("address"))
		rateSend, _ = strconv.ParseFloat(rate, 64)
		if sendBy == "urlcode" {
			//if is urlcode, create new code and create data in DB
			urlCodeErr := this.HanlderWithdrawWithUrlCode(o, loginUser, currency, amountToSend, note)
			if urlCodeErr != nil {
				logpack.Error(urlCodeErr.Error(), utils.GetFuncName(), nil)
				this.Data["json"] = map[string]string{"error": "true", "error_msg": urlCodeErr.Error()}
			} else {
				logpack.Info("Create url code successfully", utils.GetFuncName())
				this.Data["json"] = map[string]string{"error": ""}
			}
			this.ServeJSON()
			return
		}
		//if sendBy address, check address
		if sendBy == "address" {
			if utils.IsEmpty(address) {
				this.ResponseLoginError(loginUser.Id, "Address param failed. Please try again!", utils.GetFuncName(), nil)
				return
			}
			isSystemAddress := false
			//check to address is of system address
			addressObj, addrErr := utils.GetAddress(address)
			if addrErr != nil && addrErr != orm.ErrNoRows {
				this.ResponseLoginError(loginUser.Id, "DB error. Please try again!", utils.GetFuncName(), addrErr)
				return
			}
			var existAsset *models.Asset
			if addressObj != nil {
				var assetErr error
				existAsset, assetErr = utils.GetAssetById(addressObj.AssetId)
				if assetErr != nil && assetErr != orm.ErrNoRows {
					this.ResponseLoginError(loginUser.Id, "DB error. Please try again!", utils.GetFuncName(), assetErr)
					return
				}
				//if exist user for address
				if existAsset != nil {
					isSystemAddress = true
				}
			}
			if !isSystemAddress {
				_, btcHandlerErr := this.HandlerTransferOnchainCryptocurrency(o, loginUser, currency, address, amountToSend, rateSend, note)
				if btcHandlerErr != nil {
					logpack.Error(btcHandlerErr.Error(), utils.GetFuncName(), btcHandlerErr)
					this.Data["json"] = map[string]string{"error": "true", "error_msg": btcHandlerErr.Error()}
				} else {
					logpack.Info("Successfully performed transfer", utils.GetFuncName())
					this.Data["json"] = map[string]string{"error": ""}
				}
				this.ServeJSON()
				return
			} else {
				sendBy = "username"
				receiver.Id = existAsset.UserId
				receiver.Username = existAsset.UserName
			}
		}
	}

	//get assets of sender
	assetObj := assets.StringToAssetType(currency)
	senderAsset, senderAssetErr := utils.GetUserAsset(loginUser.Id, currency)
	if senderAssetErr != nil || senderAsset == nil {
		this.ResponseLoginError(loginUser.Id, "Error getting Asset data from DB or sender asset does not exist. Please try again!", utils.GetFuncName(), nil)
		return
	}
	//if balance less than amount to send, return error
	if senderAsset.Balance < amountToSend {
		this.ResponseLoginError(loginUser.Id, "Insufficient balance. Please check again or deposit more balance", utils.GetFuncName(), nil)
		return
	}
	//Deduct money from balance and update local transfer total
	senderAsset.Balance -= amountToSend
	senderAsset.LocalSent += amountToSend

	//get assets of receiver
	receiverAsset, receiverAssetErr := utils.GetUserAsset(receiver.Id, currency)
	if receiverAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Retrieve recipient asset data failed. Please try again!", utils.GetFuncName(), receiverAssetErr)
		return
	}
	//update sender asset
	_, senderAssetUpdateErr := tx.Update(senderAsset)
	if senderAssetUpdateErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Update Sender failed. Please try again!", utils.GetFuncName(), senderAssetUpdateErr)
		return
	}
	receiverAssetCreate := receiverAsset == nil

	//if receiver create asset
	if receiverAssetCreate {
		_, newReceiverAsset, newErr := this.CreateNewAddressForAsset(receiver, assetObj)
		if newErr != nil {
			this.ResponseLoginError(loginUser.Id, "Create new asset and address failed. Please check again!", utils.GetFuncName(), newErr)
			return
		}
		newReceiverAsset.Balance = amountToSend
		newReceiverAsset.LocalReceived = amountToSend
		newReceiverAsset.Updatedt = time.Now().Unix()

		_, receiverAssetUpdateErr := tx.Update(newReceiverAsset)
		if receiverAssetUpdateErr != nil {
			this.ResponseLoginRollbackError(loginUser.Id, tx, "Update balance for asset failed. Please check again!", utils.GetFuncName(), receiverAssetUpdateErr)
			return
		}
	} else {
		//update receiver asset
		receiverAsset.Balance += amountToSend
		receiverAsset.LocalReceived += amountToSend
		_, receiverAssetUpdateErr := tx.Update(receiverAsset)
		if receiverAssetUpdateErr != nil {
			this.ResponseLoginRollbackError(loginUser.Id, tx, "Update recipient assets failed. Please try again!", utils.GetFuncName(), receiverAssetUpdateErr)
			return
		}
	}

	//insert to transaction history
	txHistory := models.TxHistory{}
	txHistory.SenderId = loginUser.Id
	txHistory.Sender = loginUser.Username
	txHistory.ReceiverId = receiver.Id
	txHistory.Receiver = receiver.Username
	txHistory.Currency = currency
	txHistory.Amount = amountToSend
	txHistory.Status = 1
	txHistory.Description = note
	txHistory.Createdt = time.Now().UnixNano() / 1e9
	txHistory.TransType = int(utils.TransTypeLocal)
	txHistory.Rate = rateSend

	_, HistoryErr := tx.Insert(&txHistory)
	if HistoryErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Recorded history is corrupted. Please check your balance again!", utils.GetFuncName(), HistoryErr)
		return
	} else {
		prefix := ""
		postfix := ""
		amount := ""
		if currency == assets.USDWalletAsset.String() {
			prefix = "$"
			amount = fmt.Sprintf("%.2f", amountToSend)
		} else {
			postfix = assetObj.ToStringUpper()
			amount = fmt.Sprintf("%.8f", amountToSend)
		}
		this.SetSession("successMessage", fmt.Sprintf("Sent %s%s%s to %s successfully. Please check your balance again!", prefix, amount, postfix, receiver.Username))
	}
	tx.Commit()
	// //update loginUser session if adding to contact
	// if addedToContact {
	// 	this.SetSession("LoginUser", loginUser)
	// }
	this.ResponseSuccessfully(loginUser.Id, "Money transfer successful", utils.GetFuncName())
}

func (this *TransferController) HanlderWithdrawWithUrlCode(o orm.Ormer, sender *models.AuthClaims, asset string, amountToSend float64, note string) error {
	//check valid balance
	senderAsset, senderErr := utils.GetUserAsset(sender.Id, asset)
	if senderErr != nil {
		return fmt.Errorf("Get asset of sender failed. Please try again!")
	}
	//send to address
	fromAddressList, addrListErr := this.GetAddressListByAssetId(senderAsset.Id)
	if addrListErr != nil {
		this.ResponseLoginError(sender.Id, "Get address list of loginUser failed or address list empty. Please check again!", utils.GetFuncName(), addrListErr)
		return fmt.Errorf("Get address list of loginUser failed or address list empty. Please check again!")
	}
	//start estimate fee and size
	unitAmount, unitAmountErr := utils.GetUnitAmount(amountToSend, asset)
	if unitAmountErr != nil {
		return fmt.Errorf("Get UnitAmount from amount failed!")
	}
	//Check RPC client
	handler.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return fmt.Errorf("Create RPC Client failed")
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()

	senderBalance := senderAsset.Balance
	//check fee before transfer handler
	feeAndSize, err := assetObj.EstimateFeeAndSize(&assets.TxTarget{
		FromAddresses: fromAddressList,
		ToAddress:     assetObj.GetSystemAddress(),
		Amount:        amountToSend,
		UnitAmount:    unitAmount,
		Account:       sender.Username,
	})
	if err != nil {
		return fmt.Errorf("Check transaction cost error. Please check again!")
	}
	//if sender balance less than amount to send
	if senderBalance < amountToSend+feeAndSize.Fee.CoinValue {
		return fmt.Errorf("The balance is not enough to make this transaction. Please try again!")
	}

	//Generate code
	newCode, codeCreated := utils.CreateNewUrlCode()
	if !codeCreated {
		return fmt.Errorf("Create new code failed. Please try again!")
	}
	//create new TxCode
	newTxCode := &models.TxCode{
		Asset:     asset,
		Code:      newCode,
		OwnerId:   sender.Id,
		OwnerName: sender.Username,
		Amount:    amountToSend,
		Status:    int(utils.UrlCodeStatusCreated),
		Note:      note,
		Createdt:  time.Now().Unix(),
	}
	tx, beginErr := o.Begin()
	if beginErr != nil {
		return fmt.Errorf("Start transaction db failed")
	}

	_, insertErr := tx.Insert(newTxCode)
	if insertErr != nil {
		tx.Rollback()
		return fmt.Errorf("Creating new URL Code failed")
	}
	tx.Commit()
	return nil
}

func (this *TransferController) HandlerTransferOnchainCryptocurrency(o orm.Ormer, sender *models.AuthClaims, currency string, address string, amountToSend float64, rate float64, note string) (*models.TxHistory, error) {
	//check address valid
	valid := utils.CheckValidAddress(currency, address)
	if !valid {
		return nil, fmt.Errorf("Invalid address. Please check again!")
	}
	assetType := assets.StringToAssetType(currency)
	tx, beginErr := o.Begin()
	if beginErr != nil {
		return nil, fmt.Errorf("Start transaction db failed")
	}

	//Check RPC client
	handler.UpdateAssetManagerByType(assetType.String())
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType.String()]
	if !assetMgrExist {
		return nil, fmt.Errorf("Create RPC Client failed")
	}
	senderAsset, summaryErr := utils.GetUserAsset(sender.Id, assetType.String())
	if summaryErr != nil {
		return nil, fmt.Errorf("Get summary of loginuser asset failed. Please try again!")
	}
	//send to address
	var hash string
	var sendErr error
	fromAddressList, addrListErr := this.GetAddressListByAssetId(senderAsset.Id)
	if addrListErr != nil {
		this.ResponseLoginError(sender.Id, "Get address list of loginUser failed or address list empty. Please check again!", utils.GetFuncName(), addrListErr)
		return nil, fmt.Errorf("Get address list of loginUser failed or address list empty. Please check again!")
	}
	//start estimate fee and size
	unitAmount, unitAmountErr := utils.GetUnitAmount(amountToSend, currency)
	if unitAmountErr != nil {
		return nil, fmt.Errorf("Get UnitAmount from amount failed!")
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()

	senderBalance := senderAsset.Balance
	//check fee before transfer handler
	feeAndSize, err := assetObj.EstimateFeeAndSize(&assets.TxTarget{
		FromAddresses: fromAddressList,
		ToAddress:     address,
		Amount:        amountToSend,
		UnitAmount:    unitAmount,
		Account:       sender.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("Check transaction cost error. Please check again!")
	}
	//if sender balance less than amount to send
	if senderBalance < amountToSend+feeAndSize.Fee.CoinValue {
		return nil, fmt.Errorf("The balance is not enough to make this transaction. Please try again!")
	}

	hash, sendErr = assetObj.SendToAddressObfuscateUtxos(&assets.TxTarget{
		FromAddresses: fromAddressList,
		ToAddress:     address,
		Amount:        amountToSend,
		UnitAmount:    unitAmount,
		Account:       sender.Username,
	})
	if sendErr != nil {
		return nil, fmt.Errorf(sendErr.Error())
	}

	//Get Transaction fee
	transFee, feeErr := assetObj.GetTransactionFee(hash)
	if feeErr != nil {
		return nil, fmt.Errorf(feeErr.Error())
	}

	if transFee < 0 {
		transFee = -transFee
	}

	//update for sender: = amount + fee
	senderAsset.ChainSent += amountToSend + transFee
	senderAsset.TotalFee += transFee
	senderAsset.Balance -= amountToSend + transFee
	senderAsset.OnChainBalance -= amountToSend + transFee
	//update asset
	_, updateErr := tx.Update(senderAsset)
	if updateErr != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Update Sender Asset failed. Please try again!")
	}

	//update txhistory for sender
	txHistory := models.TxHistory{}
	txHistory.SenderId = sender.Id
	txHistory.Sender = sender.Username
	txHistory.TransType = int(utils.TransTypeChainSend)
	txHistory.ToAddress = address
	txHistory.Currency = currency
	txHistory.Amount = amountToSend
	txHistory.Rate = rate
	txHistory.Txid = hash
	txHistory.Fee = transFee
	txHistory.Status = int(utils.AssetStatusActive)
	txHistory.Description = note
	txHistory.Createdt = time.Now().Unix()

	_, HistoryErr := tx.Insert(&txHistory)
	if HistoryErr != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Generate txHistory error. Please contact admin")
	} else {
		amount := fmt.Sprintf("%.8f", amountToSend)
		var toName = address
		this.SetSession("successMessage", fmt.Sprintf("Sent %s %s to %s successfully. Please check your balance again!", amount, assetType.ToStringUpper(), toName))
	}
	tx.Commit()
	return &txHistory, nil
}

func (this *TransferController) InitTransactionHistoryList(loginUser *models.AuthClaims, assetType string, direction string, perpage int64, pageNum int64) ([]models.TxHistoryDisplay, int64) {
	historyDispList := make([]models.TxHistoryDisplay, 0)
	o := orm.NewOrm()
	var txHistoryList []*models.TxHistory
	var filterStr = ""
	if !utils.IsEmpty(assetType) {
		filterStr = fmt.Sprintf(" AND currency = '%s'", assetType)
	}

	//direction condition initialization
	var directionFilter = ""
	if !utils.IsEmpty(direction) {
		switch direction {
		case "buy":
			directionFilter = fmt.Sprintf("sender_id = %d AND is_trading = %v AND trading_type = '%s'", loginUser.Id, true, utils.TradingTypeBuy)
		case "sell":
			directionFilter = fmt.Sprintf("sender_id = %d AND is_trading = %v AND trading_type = '%s'", loginUser.Id, true, utils.TradingTypeSell)
		case "sent":
			directionFilter = fmt.Sprintf("sender_id = %d AND is_trading = %v", loginUser.Id, false)
		case "received":
			directionFilter = fmt.Sprintf("receiver_id = %d", loginUser.Id)
		case "offchainsent":
			directionFilter = fmt.Sprintf("sender_id = %d AND trans_type= %d AND is_trading = %v", loginUser.Id, int(utils.TransTypeLocal), false)
		case "offchainreceived":
			directionFilter = fmt.Sprintf("receiver_id = %d AND trans_type= %d", loginUser.Id, int(utils.TransTypeLocal))
		case "onchainsent":
			directionFilter = fmt.Sprintf("sender_id = %d AND trans_type= %d", loginUser.Id, int(utils.TransTypeChainSend))
		case "onchainreceived":
			directionFilter = fmt.Sprintf("receiver_id = %d AND trans_type= %d", loginUser.Id, int(utils.TransTypeChainReceive))
		}
	} else {
		directionFilter = fmt.Sprintf("(sender_id = %d OR receiver_id = %d)", loginUser.Id, loginUser.Id)
	}

	queryCount := fmt.Sprintf("SELECT COUNT(*) from %stx_history WHERE %s%s", utils.GetAssetRelatedTablePrefix(), directionFilter, filterStr)
	var totalRowCount int64
	countErr := o.Raw(queryCount).QueryRow(&totalRowCount)
	if countErr != nil {
		return historyDispList, 0
	}

	//page number
	pageCount := totalRowCount / perpage
	if totalRowCount%perpage != 0 {
		pageCount += 1
	}

	offset := perpage * (pageNum - 1)
	queryBuilder := fmt.Sprintf("SELECT * from %stx_history WHERE %s%s ORDER BY createdt DESC OFFSET %d LIMIT %d", utils.GetAssetRelatedTablePrefix(), directionFilter, filterStr, offset, perpage)
	_, listErr := o.Raw(queryBuilder).QueryRows(&txHistoryList)
	if listErr != nil {
		return historyDispList, 0
	}

	//get assetlist of user
	assetList, assetErr := this.GetAssetList(loginUser.Id)
	if assetErr != nil {
		return historyDispList, pageCount
	}
	rpcClientMap := make(map[string]assets.Asset)
	for _, asset := range assetList {
		if asset.Type == assets.USDWalletAsset.String() {
			continue
		}
		//Check RPC client
		handler.UpdateAssetManagerByType(asset.Type)
		assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset.Type]
		if !assetMgrExist {
			continue
		}
		rpcClientMap[asset.Type] = assetObj
	}
	for _, txHistory := range txHistoryList {
		createDt := time.Unix(txHistory.Createdt, 0)
		typeDisplay := utils.GetTransTypeFromValue(txHistory.TransType).ToString()
		historyDisp := models.TxHistoryDisplay{
			TxHistory:    *txHistory,
			IsSender:     loginUser.Id == txHistory.SenderId,
			RateValue:    txHistory.Rate * txHistory.Amount,
			TypeDisplay:  typeDisplay,
			IsOffChain:   txHistory.TransType == int(utils.TransTypeLocal),
			CreatedtDisp: createDt.Format("2006/01/02, 15:04:05"),
		}
		//handler number of needed confirmations
		if txHistory.Currency == assets.DCRWalletAsset.String() {
			historyDisp.ConfirmationNeed = 2
		} else {
			historyDisp.ConfirmationNeed = 6
		}
		historyDispList = append(historyDispList, historyDisp)
	}
	return historyDispList, pageCount
}

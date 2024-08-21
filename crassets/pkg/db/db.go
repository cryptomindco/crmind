package db

import (
	"crassets/pkg/config"
	"crassets/pkg/logpack"
	"crassets/pkg/models"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var TestnetFlg bool

type Handler struct {
	DB *gorm.DB
}

func Init(config config.Config) Handler {
	isTestnet := flag.Bool("testnet", false, "a bool")
	flag.Parse()
	TestnetFlg = *isTestnet
	networkName := "mainnet"
	if TestnetFlg {
		networkName = "testnet"
	}
	logpack.Info("Start init DB with ", networkName)
	tablePrefix := ""
	if TestnetFlg {
		tablePrefix = "t_"
	}
	db, err := gorm.Open(postgres.Open(config.DBUrl), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: tablePrefix,
		},
	})

	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(
		&models.TxHistory{},
		&models.Asset{},
		&models.Addresses{},
		&models.TxCode{},
		&models.Accounts{},
		&models.Rates{},
	)
	return Handler{DB: db}
}

func (h *Handler) GetSuperadminSystemAddress(assetType string) (string, error) {
	//Get superadmin asset
	adminAsset, err := h.GetSuperAdminAsset(assetType)
	if err != nil || adminAsset == nil {
		return "", fmt.Errorf("get superAdmin asset failed")
	}
	address, err := h.GetLastestAddressOfAsset(adminAsset.Id)
	if err != nil {
		return "", err
	}
	return address.Address, nil
}

func (h *Handler) GetLastestAddressOfAsset(assetId int64) (*models.Addresses, error) {
	address := models.Addresses{}
	queryErr := h.DB.Where(&models.Addresses{AssetId: assetId}).First(&address).Error
	if queryErr != nil {
		if queryErr != gorm.ErrRecordNotFound {
			return nil, queryErr
		}
		return nil, nil
	}
	return &address, queryErr
}

func (h *Handler) GetSuperAdminAsset(assetType string) (*models.Asset, error) {
	asset := models.Asset{}
	queryErr := h.DB.Where(&models.Asset{IsAdmin: true, Type: assetType}).First(&asset).Error
	if queryErr != nil {
		if queryErr == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, queryErr
	}
	return &asset, nil
}

func (h *Handler) ReadRateFromDB() (*models.RateObject, error) {
	//get rate String
	usdResult := make(map[string]float64)
	allResult := make(map[string]float64)
	rateJsonStr, allRate, readErr := h.ReadRateJsonStrFromDB()
	if readErr != nil || utils.IsEmpty(rateJsonStr) {
		return nil, fmt.Errorf("Get asset rates failed")
	}
	//Unamrshal json
	json.Unmarshal([]byte(rateJsonStr), &usdResult)
	json.Unmarshal([]byte(allRate), &allResult)
	return &models.RateObject{
		UsdRates: usdResult,
		AllRates: allResult,
	}, nil
}

// return: usdRate, allRate, error
func (h *Handler) ReadRateJsonStrFromDB() (string, string, error) {
	settings := models.Rates{}
	queryBuilder := fmt.Sprintf("SELECT * from rates")
	settingsErr := h.DB.Raw(queryBuilder).Scan(&settings).Error
	if settingsErr != nil {
		return "", "", settingsErr
	}
	return settings.UsdRate, settings.AllRate, nil
}

func (h *Handler) GetUnconfirmedTxHistoryList() ([]models.TxHistory, error) {
	result := make([]models.TxHistory, 0)
	listErr := h.DB.Where("confirmed AND currency <> ?", assets.USDWalletAsset.String()).Find(&result).Error
	return result, listErr
}

func (h *Handler) GetAddress(address string) (*models.Addresses, error) {
	addressObj := models.Addresses{}
	queryErr := h.DB.Where(&models.Addresses{Address: address}).First(&addressObj).Error
	if queryErr != nil {
		return nil, queryErr
	}
	return &addressObj, nil
}

func (h *Handler) GetAssetById(assetId int64) (*models.Asset, error) {
	asset := models.Asset{}
	queryErr := h.DB.Where(&models.Asset{Id: assetId, Status: int(utils.AssetStatusActive)}).First(&asset).Error
	if queryErr != nil {
		return nil, queryErr
	}
	return &asset, nil
}

func (h *Handler) GetUserAsset(username string, assetType string) (*models.Asset, error) {
	asset := models.Asset{}
	queryErr := h.DB.Where(&models.Asset{UserName: username, Type: assetType}).First(&asset).Error
	if queryErr != nil {
		if queryErr == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, queryErr
	}
	return &asset, nil
}

func (h *Handler) GetTxHistoryByTxid(txid string) (*models.TxHistory, error) {
	//Get txHistory by id
	history := models.TxHistory{}
	err := h.DB.Where(&models.TxHistory{Txid: txid}).First(&history).Error
	return &history, err
}

func (h *Handler) WriteRateToDB(usdRateMap map[string]float64, allRateMap map[string]float64) {
	usdResultString, jsonErr := json.Marshal(usdRateMap)
	allRateString, allJsonErr := json.Marshal(allRateMap)
	if jsonErr != nil || allJsonErr != nil {
		return
	}
	rates, err := h.GetRates()
	isCreate := false
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//insert new rates
			isCreate = true
			rates = &models.Rates{}
		} else {
			logpack.Warn("Get rates from DB failed", utils.GetFuncName())
			return
		}
	}
	rates.UsdRate = string(usdResultString)
	rates.AllRate = string(allRateString)
	rates.Updatedt = time.Now().Unix()
	tx := h.DB.Begin()
	if isCreate {
		err = tx.Create(rates).Error
	} else {
		err = tx.Save(rates).Error
	}
	if err != nil {
		tx.Rollback()
		logpack.Error("Update rates table failed", utils.GetFuncName(), err)
	}
}

func (h *Handler) GetAccountFromUsername(username string) (*models.Accounts, error) {
	account := models.Accounts{}
	queryErr := h.DB.Where(&models.Accounts{Username: username}).First(&account).Error
	if queryErr != nil {
		return nil, queryErr
	}
	return &account, nil
}

func (h *Handler) GetTotalUserBalance(asset string) float64 {
	var totalBalance float64
	totalBalance = 0
	//TODO: handler on services
	// queryBuilder := fmt.Sprintf("SELECT SUM(balance) FROM (SELECT * FROM public.%sasset WHERE type = '%s') ats "+
	// 	"INNER JOIN "+
	// 	"(SELECT * FROM public.user WHERE status = %d) us ON ats.user_id = us.id;", utils.GetAssetRelatedTablePrefix(), asset, int(utils.StatusActive))
	// queryErr := h.DB.Raw(queryBuilder).Scan(&totalBalance).Error
	// if queryErr != nil {
	// 	return 0
	// }
	return totalBalance
}

// Create new user token, 6 characters
func (h *Handler) CreateNewUserToken() (string, bool) {
	breakLoop := 0
	//Try up to 10 times if token creation fails
	for breakLoop < 10 {
		newToken := utils.RandSeq(6)
		breakLoop++
		//check token exist on user table
		var userCount int64
		queryErr := h.DB.Where(&models.Accounts{Token: newToken}).Count(&userCount).Error
		if queryErr != nil {
			continue
		}
		if userCount == 0 {
			return newToken, true
		}
	}
	return "", false
}

// Check and create new token for user, if exist, ignore
func (h *Handler) CheckAndCreateAccountToken(username string, role int) (token string, updated bool, err error) {
	isCreate := false
	//get user
	currentAccount, userErr := h.GetAccountFromUsername(username)
	if userErr != nil {
		if userErr == gorm.ErrRecordNotFound {
			isCreate = true
		} else {
			err = userErr
			return
		}
	}

	if !isCreate && !utils.IsEmpty(currentAccount.Token) {
		token = currentAccount.Token
		updated = false
		return
	}
	//Create new token
	newToken, ok := h.CreateNewUserToken()
	if !ok {
		err = fmt.Errorf("%s", "Create new token failed")
		return
	}
	tx := h.DB.Begin()
	var updateErr error
	if isCreate {
		currentAccount = &models.Accounts{
			Username: username,
			Token:    newToken,
			Role:     role,
		}
		updateErr = tx.Create(currentAccount).Error
	} else {
		currentAccount.Token = newToken
		//update new Token
		updateErr = tx.Save(currentAccount).Error
	}
	if updateErr != nil {
		tx.Rollback()
		err = updateErr
		return
	}
	token = newToken
	updated = true
	tx.Commit()
	return
}

func (h *Handler) GetAssetFromAddress(address string, assetType string) (*models.Asset, error) {
	//Check asset exist on assets table
	assets := models.Asset{}
	queryBuilder := fmt.Sprintf("SELECT * FROM %sasset WHERE type='%s' AND status=%d AND id IN (SELECT asset_id FROM %saddresses WHERE address='%s')", utils.GetAssetRelatedTablePrefix(), assetType, int(utils.AssetStatusActive), utils.GetAssetRelatedTablePrefix(), address)
	err := h.DB.Raw(queryBuilder).Scan(&assets).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}

	return &assets, nil
}

// func GetSystemUserAsset(assetType string) (*models.Asset, error) {
// 	systemUser, userErr := GetSystemUser()
// 	if userErr != nil {
// 		return nil, userErr
// 	}
// 	o := orm.NewOrm()
// 	return GetAssetByOwner(systemUser, o, assetType)
// }

func (h *Handler) GetTxcode(code string) (*models.TxCode, bool) {
	if utils.IsEmpty(code) {
		return nil, false
	}
	breakLoop := 0
	var txCode models.TxCode
	var exist bool
	//Try up to 10 times if code fails
	for breakLoop < 10 {
		breakLoop++
		//check code exist on txcode table
		queryErr := h.DB.Where(&models.TxCode{Code: code, Status: int(utils.UrlCodeStatusCreated)}).First(&txCode).Error
		if queryErr != nil {
			continue
		}
		exist = true
		break
	}
	return &txCode, exist
}

func (h *Handler) GetRateFromDBByAsset(assetType string) float64 {
	rates := models.Rates{}
	queryBuilder := fmt.Sprintf("SELECT * from rates")
	settingsErr := h.DB.Raw(queryBuilder).Scan(&rates).Error
	if settingsErr != nil {
		return 0
	}
	//get rate String
	result := make(map[string]float64)
	rateJsonStr := rates.UsdRate
	if utils.IsEmpty(rateJsonStr) {
		return 0
	}
	//Unamrshal json
	json.Unmarshal([]byte(rateJsonStr), &result)
	return result[assetType]
}

func (h *Handler) CheckMatchAddressWithUser(assetId, addressId int64, username string, archived bool) bool {
	queryBuilder := fmt.Sprintf("SELECT count(*) from %sasset as aet where id = %d AND user_name = %s AND EXISTS(SELECT 1 FROM %saddresses WHERE id = %d AND asset_id = aet.id AND archived=%v)", utils.GetAssetRelatedTablePrefix(),
		assetId, username, utils.GetAssetRelatedTablePrefix(), addressId, archived)
	var count int64
	countErr := h.DB.Raw(queryBuilder).Scan(&count).Error
	return countErr == nil && count > 0
}

func (h *Handler) GetAddressById(addressId int64) (*models.Addresses, error) {
	address := models.Addresses{}
	queryErr := h.DB.Where(&models.Addresses{Id: addressId}).First(&address).Error
	if queryErr != nil {
		return nil, queryErr
	}
	return &address, nil
}

func (h *Handler) CheckAssetMatchWithUser(assetId int64, username string) bool {
	var count int64
	queryErr := h.DB.Where(&models.Asset{UserName: username, Id: assetId}).Count(&count).Error
	return queryErr == nil && count > 0
}

func (h *Handler) FilterAddressList(assetId int64, status string) ([]models.Addresses, error) {
	var checkArchived, archived bool
	switch status {
	case "all":
		checkArchived = false
	case "active":
		checkArchived = true
		archived = false
	case "archived":
		checkArchived = true
		archived = true
	default:
		checkArchived = true
		archived = false
	}
	result := make([]models.Addresses, 0)
	var listErr error
	if checkArchived {
		listErr = h.DB.Where(&models.Addresses{AssetId: assetId, Archived: archived}).Order("createdt desc").Find(&result).Error
	} else {
		listErr = h.DB.Where(&models.Addresses{AssetId: assetId}).Order("createdt desc").Find(&result).Error
	}
	if listErr != nil && listErr != gorm.ErrRecordNotFound {
		return nil, listErr
	}
	return result, nil
}

func (h *Handler) FilterUrlCodeList(assetType string, status string, username string) ([]models.TxCode, error) {
	var statusInt int
	switch status {
	case "unconfirmed":
		statusInt = int(utils.UrlCodeStatusCreated)
	case "confirmed":
		statusInt = int(utils.UrlCodeStatusConfirmed)
	case "cancelled":
		statusInt = int(utils.UrlCodeStatusCancelled)
	default:
		statusInt = -1
	}
	result := make([]models.TxCode, 0)
	var listErr error
	if statusInt >= 0 {
		listErr = h.DB.Where(&models.TxCode{Asset: assetType, OwnerName: username, Status: statusInt}).Find(&result).Error
	} else {
		listErr = h.DB.Where(&models.TxCode{Asset: assetType, OwnerName: username}).Find(&result).Error
	}
	if listErr != nil && listErr != gorm.ErrRecordNotFound {
		return nil, listErr
	}
	return result, nil
}

func (h *Handler) GetTxHistoryById(txHistoryId int64) (*models.TxHistory, error) {
	//Get txHistory by id
	history := models.TxHistory{}
	err := h.DB.Where(&models.TxHistory{Id: txHistoryId}).First(&history).Error
	return &history, err
}

func (h *Handler) CancelTxCodeById(ownername string, codeId int64) error {
	txCode := models.TxCode{}
	queryErr := h.DB.Where(&models.TxCode{Id: codeId}).First(&txCode).Error
	if queryErr != nil {
		return queryErr
	}
	//if ownerId not match
	if ownername != txCode.OwnerName {
		return fmt.Errorf("%s", "Owner not match. There is no right to cancel this Code")
	}
	txCode.Status = int(utils.UrlCodeStatusCancelled)
	//update txCode
	tx := h.DB.Begin()
	updateErr := tx.Save(&txCode).Error
	if updateErr != nil {
		tx.Rollback()
	}
	return updateErr
}

// Create new code for withdraw url, 32 characters
func (h *Handler) CreateNewUrlCode() (string, bool) {
	breakLoop := 0
	//Try up to 10 times if token creation fails
	for breakLoop <= 10 {
		newCode := utils.RandSeq(32)
		breakLoop++
		//check token exist on user table
		var codeCount int64
		queryErr := h.DB.Where(&models.TxCode{Code: newCode}).Count(&codeCount).Error
		if queryErr != nil {
			continue
		}
		if codeCount == 0 {
			return newCode, true
		}
	}
	return "", false
}

func (h *Handler) GetRates() (*models.Rates, error) {
	rates := models.Rates{}
	queryBuilder := fmt.Sprintf("SELECT * from rates")
	ratesErr := h.DB.Raw(queryBuilder).Scan(&rates).Error
	if ratesErr != nil {
		return nil, ratesErr
	}
	return &rates, nil
}

func (h *Handler) GetContactListFromUser(username string) ([]models.ContactItem, error) {
	account, userErr := h.GetAccountFromUsername(username)
	if userErr != nil {
		return nil, userErr
	}
	result := make([]models.ContactItem, 0)
	if utils.IsEmpty(account.Contacts) {
		return result, nil
	}
	err := json.Unmarshal([]byte(account.Contacts), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *Handler) UpdateUserContacts(username, contacts string) error {
	account := models.Accounts{}
	isCreate := false
	err := h.DB.Where(&models.Accounts{Username: username}).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			isCreate = true
		} else {
			return err
		}
	}
	tx := h.DB.Begin()
	account.Contacts = contacts
	if isCreate {
		account.Username = username
		err = tx.Create(&account).Error
	} else {
		err = tx.Save(&account).Error
	}
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) GetTokenFromUsername(username string) string {
	account, err := h.GetAccountFromUsername(username)
	if err != nil {
		return ""
	}
	return account.Token
}

func (h *Handler) GetUserFromToken(token string) (*models.Accounts, error) {
	account := models.Accounts{}
	queryErr := h.DB.Where(&models.Accounts{Token: token}).First(&account).Error
	if queryErr != nil {
		return nil, queryErr
	}
	return &account, nil
}

// Check and get user from address label
func (h *Handler) GetUserFromLabel(label string) (*models.Accounts, bool) {
	if utils.IsEmpty(label) {
		return nil, false
	}

	//split label
	labelArr := strings.Split(label, "_")
	if len(labelArr) <= 1 {
		return nil, false
	}

	//token
	token := labelArr[0]
	//check token from user
	account, userErr := h.GetUserFromToken(token)
	if userErr != nil {
		return nil, false
	}
	return account, true
}

func (h *Handler) GetSystemUserAsset(assetType string) (*models.Asset, error) {
	asset := models.Asset{}
	queryErr := h.DB.Where(&models.Asset{IsAdmin: true, Type: assetType, Status: int(utils.AssetStatusActive)}).First(&asset).Error
	return &asset, queryErr
}

// Check user exist with username and status active
func (h *Handler) CountAddressesWithStatus(assetId int64, activeFlg bool) int64 {
	var count int64
	queryErr := h.DB.Where(&models.Addresses{AssetId: assetId, Archived: !activeFlg}).Count(&count).Error
	if queryErr != nil {
		return 0
	}
	return count
}

func (h *Handler) CheckHasCodeList(assetType string, username string) bool {
	var count int64
	queryErr := h.DB.Where(&models.TxCode{Asset: assetType, OwnerName: username}).Count(&count).Error
	var exist = queryErr == nil && count > 0
	return exist
}

func (h *Handler) GetContactListOfUser(username string) []string {
	result := make([]string, 0)
	account, userErr := h.GetAccountFromUsername(username)
	if userErr != nil {
		return result
	}

	if utils.IsEmpty(account.Contacts) {
		return result
	}

	var contacts []models.ContactItem
	err := json.Unmarshal([]byte(account.Contacts), &contacts)
	if err != nil {
		return result
	}
	for _, contact := range contacts {
		result = append(result, contact.UserName)
	}
	return result
}

func (h *Handler) HandlerTransferOnchainCryptocurrency(senderName string, currency string, address string, amountToSend float64, rate float64, note string) (*models.TxHistory, error) {
	//check address valid
	valid := utils.CheckValidAddress(currency, address)
	if !valid {
		return nil, fmt.Errorf("Invalid address. Please check again!")
	}
	assetType := assets.StringToAssetType(currency)
	tx := h.DB.Begin()
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetType.String()]
	if !assetMgrExist {
		return nil, fmt.Errorf("Create RPC Client failed")
	}
	senderAsset, summaryErr := h.GetUserAsset(senderName, assetType.String())
	if summaryErr != nil {
		return nil, fmt.Errorf("Get summary of loginuser asset failed. Please try again!")
	}
	//send to address
	var hash string
	var sendErr error
	fromAddressList, addrListErr := h.GetAddressListByAssetId(senderAsset.Id)
	if addrListErr != nil {
		logpack.Error("Get address list of loginUser failed or address list empty. Please check again!", utils.GetFuncName(), addrListErr)
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
		Account:       senderName,
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
		Account:       senderName,
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
	updateErr := tx.Save(senderAsset).Error
	if updateErr != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Update Sender Asset failed. Please try again!")
	}

	//update txhistory for sender
	txHistory := models.TxHistory{}
	txHistory.Sender = senderName
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

	HistoryErr := tx.Create(&txHistory).Error
	if HistoryErr != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Generate txHistory error. Please contact admin")
	}
	tx.Commit()
	return &txHistory, nil
}

func (h *Handler) GetAddressListByAssetId(assetId int64) ([]string, error) {
	addressList := make([]*models.Addresses, 0)
	queryErr := h.DB.Where(&models.Addresses{AssetId: assetId, Archived: false}).Order("createdt desc").Find(&addressList).Error
	if queryErr != nil {
		return make([]string, 0), queryErr
	}
	result := make([]string, 0)
	for _, address := range addressList {
		result = append(result, address.Address)
	}
	return result, nil
}

func (h *Handler) HandlerInternalWithdrawl(txCode *models.TxCode, user models.UserInfo, rateSend float64) bool {
	tx := h.DB.Begin()
	//get assets of sender
	assetObj := assets.StringToAssetType(txCode.Asset)
	senderAsset, senderAssetErr := h.GetUserAsset(txCode.OwnerName, txCode.Asset)
	if senderAssetErr != nil || senderAsset == nil {
		logpack.Error("Error getting Asset data from DB or sender asset does not exist. Please try again!", utils.GetFuncName(), nil)
		return false
	}
	//if balance is not enough to withdraw
	if senderAsset.Balance < txCode.Amount {
		logpack.Error("Balance is not enough to withdraw", utils.GetFuncName(), nil)
		return false
	}

	//Deduct money from balance and update local transfer total
	senderAsset.Balance -= txCode.Amount
	senderAsset.LocalSent += txCode.Amount

	//get assets of receiver
	receiverAsset, receiverAssetErr := h.GetUserAsset(user.Username, txCode.Asset)
	if receiverAssetErr != nil {
		logpack.Error("Retrieve recipient asset data failed. Please try again!", utils.GetFuncName(), receiverAssetErr)
		return false
	}

	//update sender asset
	senderAssetUpdateErr := tx.Save(senderAsset).Error
	if senderAssetUpdateErr != nil {
		logpack.Error("Update Sender failed. Please try again!", utils.GetFuncName(), senderAssetUpdateErr)
		return false
	}
	receiverAssetCreate := receiverAsset == nil

	//if receiver create asset
	if receiverAssetCreate {
		_, newReceiverAsset, newErr := h.CreateNewAddressForAsset(user.Username, utils.IsSuperAdmin(user.Role), assetObj)
		if newErr != nil {
			logpack.Error("Create new asset and address failed. Please check again!", utils.GetFuncName(), newErr)
			return false
		}
		newReceiverAsset.Balance = txCode.Amount
		newReceiverAsset.LocalReceived = txCode.Amount
		newReceiverAsset.Updatedt = time.Now().Unix()

		receiverAssetUpdateErr := tx.Save(newReceiverAsset).Error
		if receiverAssetUpdateErr != nil {
			logpack.Error("Update balance for asset failed. Please check again!", utils.GetFuncName(), receiverAssetUpdateErr)
			return false
		}
	} else {
		//update receiver asset
		receiverAsset.Balance += txCode.Amount
		receiverAsset.LocalReceived += txCode.Amount
		receiverAssetUpdateErr := tx.Save(receiverAsset).Error
		if receiverAssetUpdateErr != nil {
			tx.Rollback()
			logpack.Error("Update recipient assets failed. Please try again!", utils.GetFuncName(), receiverAssetUpdateErr)
			return false
		}
	}

	//insert to transaction history
	txHistory := models.TxHistory{}
	txHistory.Sender = txCode.OwnerName
	txHistory.Receiver = user.Username
	txHistory.Currency = txCode.Asset
	txHistory.Amount = txCode.Amount
	txHistory.Status = 1
	txHistory.Description = txCode.Note
	txHistory.Createdt = time.Now().UnixNano() / 1e9
	txHistory.TransType = int(utils.TransTypeLocal)
	txHistory.Rate = rateSend

	HistoryErr := tx.Create(&txHistory).Error
	if HistoryErr != nil {
		tx.Rollback()
		logpack.Error("Recorded history is corrupted. Please check your balance again!", utils.GetFuncName(), HistoryErr)
		return false
	}

	//update txCode status
	txCode.Status = int(utils.UrlCodeStatusConfirmed)
	txCode.HistoryId = txHistory.Id
	txCode.Confirmdt = time.Now().Unix()
	txUpdateErr := tx.Save(txCode).Error
	if txUpdateErr != nil {
		tx.Rollback()
		logpack.Error("Update Tx Code status failed!", utils.GetFuncName(), txUpdateErr)
		return false
	}
	tx.Commit()
	return true
}

func (h *Handler) CreateNewAddressForAsset(username string, isAdmin bool, assetObject assets.AssetType) (*models.Addresses, *models.Asset, error) {
	tx := h.DB.Begin()
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetObject.String()]
	if !assetMgrExist {
		return nil, nil, fmt.Errorf("RPC Client failed at the server. Please contact admin!")
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()
	var assetLabel string
	if assetObject == assets.DCRWalletAsset {
		assetLabel = username
	} else {
		//Check and get user token
		var role int
		if isAdmin {
			role = int(utils.RoleSuperAdmin)
		} else {
			role = int(utils.RoleRegular)
		}
		token, _, err := h.CheckAndCreateAccountToken(username, role)
		if err != nil {
			return nil, nil, fmt.Errorf("Check or create user token failed")
		}
		//default label format: token_%label%
		assetLabel = fmt.Sprintf("%s%s", token, utils.CreateDefaultAddressLabelPostfix(assetObject.String()))
	}
	//Create new address with label. Label form is: btc_address_$username
	newAddress, addrErr := assetObj.CreateNewAddressWithLabel(username, assetLabel)
	if addrErr != nil {
		return nil, nil, fmt.Errorf("Creating an address with Label failed. Please try again!")
	}

	//Get asset from DB
	userAsset, err := h.GetUserAsset(username, assetObject.String())
	if err != nil {
		return nil, nil, fmt.Errorf("Get Asset from DB failed. Please try again!")
	}
	//if user asset is nil, insert new asset to DB
	var assetId int64
	if userAsset == nil {
		asset := models.Asset{
			DisplayName: assetObject.ToFullName(),
			UserName:    username,
			Type:        assetObject.String(),
			Sort:        assetObject.AssetSortInt(),
			Status:      int(utils.AssetStatusActive),
			IsAdmin:     isAdmin,
			Createdt:    time.Now().Unix(),
			Updatedt:    time.Now().Unix(),
		}
		//update user
		insertErr := tx.Create(&asset).Error
		if insertErr != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("Insert new asset failed. Please try again!")
		}
		userAsset = &asset
		assetId = asset.Id
	} else {
		assetId = userAsset.Id
	}

	//insert new Address
	insertAddress := models.Addresses{
		AssetId:  assetId,
		Address:  newAddress,
		Label:    assetLabel,
		Createdt: time.Now().Unix(),
	}
	insertAddressErr := tx.Create(&insertAddress).Error
	if insertAddressErr != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("Insert new address label failed. Please try again!")
	}
	//Insert to asset table
	tx.Commit()
	return &insertAddress, userAsset, nil
}

func (h *Handler) InitTransactionHistoryList(loginUser *models.UserInfo, assetType string, direction string, perpage int64, pageNum int64, allowAssets []string) ([]models.TxHistoryDisplay, int64) {
	historyDispList := make([]models.TxHistoryDisplay, 0)
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
			directionFilter = fmt.Sprintf("sender = %s AND is_trading = %v AND trading_type = '%s'", loginUser.Username, true, utils.TradingTypeBuy)
		case "sell":
			directionFilter = fmt.Sprintf("sender = %s AND is_trading = %v AND trading_type = '%s'", loginUser.Username, true, utils.TradingTypeSell)
		case "sent":
			directionFilter = fmt.Sprintf("sender = %s AND is_trading = %v", loginUser.Username, false)
		case "received":
			directionFilter = fmt.Sprintf("receiver = %s", loginUser.Username)
		case "offchainsent":
			directionFilter = fmt.Sprintf("sender = %s AND trans_type= %d AND is_trading = %v", loginUser.Username, int(utils.TransTypeLocal), false)
		case "offchainreceived":
			directionFilter = fmt.Sprintf("receiver = %s AND trans_type= %d", loginUser.Username, int(utils.TransTypeLocal))
		case "onchainsent":
			directionFilter = fmt.Sprintf("sender = %s AND trans_type= %d", loginUser.Username, int(utils.TransTypeChainSend))
		case "onchainreceived":
			directionFilter = fmt.Sprintf("receiver = %s AND trans_type= %d", loginUser.Username, int(utils.TransTypeChainReceive))
		}
	} else {
		directionFilter = fmt.Sprintf("(sender_id = %d OR receiver_id = %d)", loginUser.Username, loginUser.Username)
	}

	queryCount := fmt.Sprintf("SELECT COUNT(*) from %stx_history WHERE %s%s", utils.GetAssetRelatedTablePrefix(), directionFilter, filterStr)
	var totalRowCount int64
	countErr := h.DB.Raw(queryCount).Scan(&totalRowCount).Error
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
	listErr := h.DB.Raw(queryBuilder).Scan(&txHistoryList).Error
	if listErr != nil {
		return historyDispList, 0
	}

	//get assetlist of user
	assetList, assetErr := h.GetAssetList(loginUser.Username, allowAssets)
	if assetErr != nil {
		return historyDispList, pageCount
	}
	rpcClientMap := make(map[string]assets.Asset)
	for _, asset := range assetList {
		if asset.Type == assets.USDWalletAsset.String() {
			continue
		}
		//Check RPC client
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
			IsSender:     loginUser.Username == txHistory.Sender,
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

func (h *Handler) GetAssetList(username string, allowAsset []string) ([]*models.Asset, error) {
	var assetList []*models.Asset
	tempAllowAssets := make([]string, 0)
	for _, asset := range allowAsset {
		tempAllowAssets = append(tempAllowAssets, fmt.Sprintf("'%s'", asset))
	}
	builderSQL := fmt.Sprintf("SELECT * from %sasset WHERE user_name=%s AND status = %d AND type IN (%s) ORDER BY sort", utils.GetAssetRelatedTablePrefix(), username, int(utils.AssetStatusActive), strings.Join(tempAllowAssets, ","))
	err := h.DB.Raw(builderSQL).Scan(&assetList).Error
	if err != nil {
		return nil, err
	}
	return assetList, nil
}

func (h *Handler) HanlderWithdrawWithUrlCode(senderName, asset string, amountToSend float64, note string) error {
	//check valid balance
	senderAsset, senderErr := h.GetUserAsset(senderName, asset)
	if senderErr != nil {
		return fmt.Errorf("get asset of sender failed. Please try again!")
	}
	//send to address
	fromAddressList, addrListErr := h.GetAddressListByAssetId(senderAsset.Id)
	if addrListErr != nil {
		return fmt.Errorf("get address list of loginUser failed or address list empty. Please check again!")
	}
	//start estimate fee and size
	unitAmount, unitAmountErr := utils.GetUnitAmount(amountToSend, asset)
	if unitAmountErr != nil {
		return fmt.Errorf("get UnitAmount from amount failed!")
	}
	//Check RPC client
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return fmt.Errorf("create RPC Client failed")
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
		Account:       senderName,
	})
	if err != nil {
		return fmt.Errorf("check transaction cost error. Please check again!")
	}
	//if sender balance less than amount to send
	if senderBalance < amountToSend+feeAndSize.Fee.CoinValue {
		return fmt.Errorf("the balance is not enough to make this transaction. Please try again!")
	}

	//Generate code
	newCode, codeCreated := h.CreateNewUrlCode()
	if !codeCreated {
		return fmt.Errorf("create new code failed. Please try again!")
	}
	//create new TxCode
	newTxCode := &models.TxCode{
		Asset:     asset,
		Code:      newCode,
		OwnerName: senderName,
		Amount:    amountToSend,
		Status:    int(utils.UrlCodeStatusCreated),
		Note:      note,
		Createdt:  time.Now().Unix(),
	}

	tx := h.DB.Begin()
	insertErr := tx.Create(newTxCode).Error
	if insertErr != nil {
		tx.Rollback()
		return fmt.Errorf("creating new URL Code failed")
	}
	tx.Commit()
	return nil
}

func (h *Handler) SyncAssetList(username string, summaryList []*models.Asset, allowAssetList []string) []*models.Asset {
	result := make([]*models.Asset, 0)
	for _, allowAsset := range allowAssetList {
		exist := false
		for _, summary := range summaryList {
			if summary.Type == allowAsset {
				result = append(result, summary)
				exist = true
				break
			}
		}
		if !exist {
			summary := h.CreateNewAsset(allowAsset, username)
			result = append(result, summary)
		}
	}
	return result
}

func (h *Handler) CreateNewAsset(assetType string, username string) *models.Asset {
	assetObject := assets.StringToAssetType(assetType)
	return &models.Asset{
		Sort:          assetObject.AssetSortInt(),
		DisplayName:   assetObject.ToFullName(),
		UserName:      username,
		Type:          assetType,
		Balance:       0,
		LocalReceived: 0,
		LocalSent:     0,
		ChainReceived: 0,
		ChainSent:     0,
		TotalFee:      0,
	}
}

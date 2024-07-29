package handler

import (
	"crassets/logpack"
	"crassets/models"
	"crassets/services"
	"crassets/utils"
	"crassets/walletlib"
	"crassets/walletlib/assets"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
)

func UpdateGlobalVariable() {
	if utils.GlobalItem == nil {
		utils.GlobalItem = &utils.Globals{
			CheckedNet:  false,
			Testnet:     false,
			AssetMgrMap: make(map[string]assets.Asset),
		}
	}

	//If is not verify net param
	if !utils.GlobalItem.CheckedNet {
		utils.GlobalItem.Testnet = models.TestnetFlg
		utils.GlobalItem.CheckedNet = true
	}

	if utils.GlobalItem.AssetsAllow == nil {
		allowAssets, err := walletlib.GetAllowAssetFromSettings()
		if err != nil {
			utils.GlobalItem.AssetsAllow = make([]string, 0)
		} else {
			utils.GlobalItem.AssetsAllow = allowAssets
		}
	}

	//check and create/update asset manager
	if utils.GlobalItem.AssetMgrMap == nil {
		utils.GlobalItem.AssetMgrMap = make(map[string]assets.Asset)
	}

	//Update asset manager
	for _, asset := range utils.GlobalItem.AssetsAllow {
		if asset == assets.USDWalletAsset.String() {
			continue
		}
		_, mgrExist := utils.GlobalItem.AssetMgrMap[asset]
		if !mgrExist {
			//create asset manager
			mgr, mgrErr := walletlib.CreateAssetAndConnectRPC(assets.StringToAssetType(asset))
			if mgrErr != nil {
				logpack.Error(fmt.Sprintf("Create asset manager and RPC client for %s failed", asset), utils.GetFuncName(), mgrErr)
			} else {
				utils.GlobalItem.AssetMgrMap[asset] = mgr
			}
		}
	}
}

func GetTotalDaemonBalance(asset string) float64 {
	if asset == assets.USDWalletAsset.String() {
		return 0
	}
	UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return 0
	}
	daemonBalance, err := assetObj.GetSystemBalance()
	if err != nil {
		logpack.Error(fmt.Sprintf("Get daemon balance of %s failed", asset), utils.GetFuncName(), err)
		return 0
	}
	return daemonBalance
}

func GetSpendableAmount(asset string) float64 {
	if asset == assets.USDWalletAsset.String() {
		return 0
	}
	UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return 0
	}
	daemonBalance := assetObj.GetSpendableAmount()
	return daemonBalance
}

func UpdateAssetManagerByType(assetType string) {
	if utils.GlobalItem.AssetMgrMap == nil {
		utils.GlobalItem.AssetMgrMap = make(map[string]assets.Asset)
	}
	_, mgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !mgrExist {
		mgr, mgrErr := walletlib.CreateAssetAndConnectRPC(assets.StringToAssetType(assetType))
		if mgrErr != nil {
			return
		}
		utils.GlobalItem.AssetMgrMap[assetType] = mgr
	}
	//check wallet loaded
	assetMgr := utils.GlobalItem.AssetMgrMap[assetType]
	assetMgr.MutexLock()
	defer assetMgr.MutexUnlock()
	assetMgr.CheckAndLoadWallet()
}

func SyncTransactionConfirmed() {
	threadLoop := true
	//Every 15 seconds. Update confimed status and
	go func() {
		for threadLoop {
			time.Sleep(15 * time.Second)
			//Get unconfirmed txhistory list
			txHistoryList, txErr := utils.GetUnconfirmedTxHistoryList()
			if txErr != nil {
				continue
			}
			o := orm.NewOrm()
			tx, txOrmerErr := o.Begin()
			if txOrmerErr != nil {
				continue
			}
			//browse txHistory and update confimed status
			for _, txHistory := range txHistoryList {
				if txHistory.Confirmed || utils.IsEmpty(txHistory.Txid) {
					continue
				}
				//Get transaction
				assetObj, exist := GetRPCAssetsObject(txHistory.Currency)
				if !exist {
					continue
				}
				_, count, txError := assetObj.GetTransactionStatus(txHistory.Txid)
				if txError != nil {
					continue
				}
				txUpdate := false
				//check count is confirmed
				if utils.IsConfirmed(count, txHistory.Currency) {
					txHistory.Confirmations = int(count)
					txHistory.Confirmed = true
					txUpdate = true
					//if is receivef from external, update asset of user. Apply for LTC, BTC
					if txHistory.TransType == int(utils.TransTypeChainReceive) && txHistory.Currency != assets.DCRWalletAsset.String() {
						if txHistory.ReceiverId <= 0 {
							continue
						}
						//Get Asset for receiver
						userAsset, assetErr := utils.GetAssetByOwner(txHistory.ReceiverId, o, txHistory.Currency)
						if assetErr != nil {
							continue
						}
						userAsset.ChainReceived += txHistory.Amount
						userAsset.OnChainBalance += txHistory.Amount
						userAsset.Balance += txHistory.Amount
						userAsset.Updatedt = time.Now().Unix()
						_, assetUpdateErr := tx.Update(userAsset)
						if assetUpdateErr != nil {
							continue
						}
					}
				} else {
					if txHistory.Confirmations != int(count) {
						txHistory.Confirmations = int(count)
						txUpdate = true
					}
				}
				if txUpdate {
					_, txHisErr := tx.Update(&txHistory)
					if txHisErr != nil {
						continue
					}
				}
			}
			tx.Commit()
		}
	}()
}

func GetRPCAssetsObject(asset string) (assets.Asset, bool) {
	//Check RPC client
	UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return nil, false
	}
	return assetObj, true
}

func SyncSystemData() {
	// init exist transaction list
	o := orm.NewOrm()
	//Check time for sync (UTC zone)
	timeSync := beego.AppConfig.String("SYSTEM_SYNC_TIME")
	if utils.IsEmpty(timeSync) {
		//if haven't time sync on config, set default to time sync
		timeSync = "00:00"
	}
	hour := int64(0)
	minute := int64(0)
	timeArr := strings.Split(timeSync, ":")
	if len(timeArr) > 1 {
		hourStr := strings.TrimLeft(timeArr[0], "0")
		minuteStr := strings.TrimLeft(timeArr[1], "0")
		hour, _ = strconv.ParseInt(hourStr, 0, 32)
		minute, _ = strconv.ParseInt(minuteStr, 0, 32)
	}
	timeCompare := hour*60 + minute
	// wait group just for test
	threadLoop := true
	todaySync := false
	var lastSyncDate *time.Time
	go func() {
		for threadLoop {
			now := time.Now()
			//check if today differnce last sync data
			if lastSyncDate != nil {
				nowDateCompare := now.Year()*12 + int(now.Month())
				lastDateCompare := lastSyncDate.Year()*12 + int(lastSyncDate.Month())
				//if time moved to new day, reset flag
				if nowDateCompare > lastDateCompare {
					todaySync = false
					lastSyncDate = nil
				}
			}
			nowCompare := int64(now.Hour()*60 + now.Minute())
			compare := nowCompare - timeCompare
			if compare < 0 || compare > 5 || todaySync {
				time.Sleep(3 * time.Minute)
				continue
			}
			//sync data
			SystemSyncHandler(o)
			//set today is true
			todaySync = true
			lastSyncDate = &now
			//Check time sync after every 3 minute
			time.Sleep(3 * time.Minute)
		}
	}()
}

func CheckAssetExistOnArray(assetList []*models.Asset, userId int64, assetType string) *models.Asset {
	for _, assetTemp := range assetList {
		if assetTemp.UserId == userId && assetTemp.Type == assetType {
			return assetTemp
		}
	}
	return nil
}

func GetAssetIndexByIdFromArray(assetList []*models.Asset, id int64) int {
	for index, asset := range assetList {
		if asset.Id == id {
			return index
		}
	}
	return -1
}

// check and sync daemon data with DB data
func SyncDecredHandler(o orm.Ormer, tx orm.TxOrmer, assetMgr assets.Asset) {
	//Get account list with balance
	balanceMap := assetMgr.GetBalanceMapByLabel()
	accountAssetMap := make(map[string]*models.Asset)
	//Address array
	addressArray := make([]*models.Addresses, 0)
	//sync account on DB
	for account, balance := range balanceMap {
		//because account is username. so, check user is exist with username
		user, userErr := utils.GetUserByUsername(account, o)
		if userErr != nil {
			continue
		}
		//then, check account exist on asset
		asset, assetErr := utils.GetAssetByOwner(user, o, assets.DCRWalletAsset.String())
		//if error, continue
		if assetErr != nil && assetErr != orm.ErrNoRows {
			continue
		}
		assetObj := assets.DCRWalletAsset
		//if asset not exist, insert to DB
		if assetErr == orm.ErrNoRows {
			asset = &models.Asset{
				DisplayName:    assetObj.ToFullName(),
				UserId:         user.Id,
				UserName:       user.Username,
				Type:           assetObj.String(),
				Sort:           assetObj.AssetSortInt(),
				Balance:        balance,
				OnChainBalance: balance,
				Status:         int(utils.AssetStatusActive),
				Createdt:       time.Now().Unix(),
				Updatedt:       time.Now().Unix(),
			}
			_, assetInsertErr := tx.Insert(asset)
			if assetInsertErr != nil {
				tx.Rollback()
				continue
			}
		} else {
			if asset.OnChainBalance != balance {
				//update asset
				asset.OnChainBalance = balance
				asset.Updatedt = time.Now().Unix()
				_, assetUpdateErr := tx.Update(asset)
				if assetUpdateErr != nil {
					tx.Rollback()
					continue
				}
			}
		}
		//get address by account
		addressList, addrErr := assetMgr.GetAddressesByLabel(account)
		if addrErr == nil {
			for _, addr := range addressList {
				//check address exist on DB
				checkAddr, checkErr := utils.GetAddress(addr)
				if checkErr != nil && checkErr != orm.ErrNoRows {
					continue
				}
				//if not exist, create new
				if checkAddr == nil && checkErr == orm.ErrNoRows {
					checkAddr = &models.Addresses{
						AssetId:       asset.Id,
						Address:       addr,
						LocalReceived: 0,
						ChainReceived: 0,
						Label:         user.Username,
						Createdt:      time.Now().Unix(),
					}
				}
				addressArray = append(addressArray, checkAddr)
			}
		}
		accountAssetMap[account] = asset
	}
	//get all transactions
	allTransactions, err := assetMgr.GetAllTransactions()
	if err != nil {
		tx.Commit()
		return
	}
	receivedTransactions := make([]assets.ListTransactionsResult, 0)
	for _, transaction := range allTransactions {
		if transaction.Category == assets.CategoryReceive {
			//check account exist on map
			_, accountExist := accountAssetMap[transaction.Account]
			if accountExist {
				receivedTransactions = append(receivedTransactions, transaction)
			}
		}
	}
	//with dcr, only check received transactions
	for _, receivedTransaction := range receivedTransactions {
		asset, exist := accountAssetMap[receivedTransaction.Account]
		if !exist {
			continue
		}

		addrExistOnArray := false
		var existAddressOnArray *models.Addresses
		var existIndex int
		//check from addressArray
		for index, addr := range addressArray {
			if addr.Address == receivedTransaction.Address {
				addrExistOnArray = true
				existAddressOnArray = addr
				existIndex = index
				break
			}
		}

		//if not exist on array
		if !addrExistOnArray {
			//check on DB
			//check address exist on DB
			addressObj, addrErr := utils.GetAddress(receivedTransaction.Address)
			//if get address error
			if addrErr != nil && addrErr != orm.ErrNoRows {
				continue
			}
			if addrErr == orm.ErrNoRows {
				addressObj = &models.Addresses{
					AssetId:       asset.Id,
					Address:       receivedTransaction.Address,
					ChainReceived: receivedTransaction.Amount,
					Transactions:  1,
					Createdt:      time.Now().Unix(),
				}
				addressArray = append(addressArray, addressObj)
			} else {
				existAddressOnArray = addressObj
			}
		}
		if existAddressOnArray != nil {
			//if assetid on address not match with assetId, re update address
			if existAddressOnArray.AssetId != asset.Id {
				existAddressOnArray.AssetId = asset.Id
			}
			existAddressOnArray.ChainReceived += receivedTransaction.Amount
			existAddressOnArray.Transactions++
			if !addrExistOnArray {
				addressArray = append(addressArray, existAddressOnArray)
			} else {
				addressArray[existIndex] = existAddressOnArray
			}
		}
		//update received of asset
		asset.ChainReceived += receivedTransaction.Amount
		accountAssetMap[receivedTransaction.Account] = asset
	}

	//update assets
	for _, updateAsset := range accountAssetMap {
		if updateAsset.Id > 0 {
			//get received
			receivedAmount, receivedErr := assetMgr.GetReceivedByAccount(updateAsset.UserName)
			if receivedErr == nil {
				updateAsset.ChainReceived = receivedAmount
			}
			//update onchain sent amount
			updateAsset.ChainSent = updateAsset.ChainReceived - updateAsset.OnChainBalance
			updateAsset.Balance = updateAsset.LocalReceived - updateAsset.LocalSent + updateAsset.OnChainBalance
			updateAsset.Updatedt = time.Now().Unix()
			//update
			_, assetUpdateErr := tx.Update(updateAsset)
			if assetUpdateErr != nil {
				tx.Rollback()
				continue
			}
		}
	}
	//update or insert address
	for _, updateAddress := range addressArray {
		//if update
		if updateAddress.Id > 0 {
			_, addrUpdateErr := tx.Update(updateAddress)
			if addrUpdateErr != nil {
				tx.Rollback()
				continue
			}
		} else {
			//if create
			_, addrInsertErr := tx.Insert(updateAddress)
			if addrInsertErr != nil {
				tx.Rollback()
				continue
			}
		}
	}
}

// Sync blockchain data with system DB data to check data consistency
func SystemSyncHandler(o orm.Ormer) {
	tx, beginErr := o.Begin()
	if beginErr != nil {
		logpack.Error("An error has occurred", utils.GetFuncName(), beginErr)
	}
	//Get allow asset list
	assetList, err := walletlib.GetAllowAssetFromSettings()
	if err != nil {
		logpack.Error("Sync by date error", utils.GetFuncName(), err)
		return
	}
	for _, asset := range assetList {
		//If is USD, skip
		if asset == assets.USDWalletAsset.String() {
			continue
		}
		//start check and sync
		//Check RPC client
		UpdateAssetManagerByType(asset)
		assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
		if !assetMgrExist {
			logpack.Error(fmt.Sprintf("Create %s RPC Client failed", asset), utils.GetFuncName(), nil)
			continue
		}
		//if is decred, separate handler
		if asset == assets.DCRWalletAsset.String() {
			SyncDecredHandler(o, tx, assetObj)
			continue
		}
		//Sync address and account
		//get label list
		labelList := assetObj.GetLabelList()
		addressUpdateMap := make(map[string]*models.Addresses)
		assetUpdateArr := make([]*models.Asset, 0)

		for _, label := range labelList {
			//check label is label of address in system
			user, existLabel := utils.GetUserFromLabel(label)
			if !existLabel {
				continue
			}
			// get addresses by label
			assetObj.MutexLock()
			addressList, addrErr := assetObj.GetAddressesByLabel(label)
			assetObj.MutexUnlock()
			if addrErr != nil {
				logpack.Error(fmt.Sprintf("Get Address by Label Failed. Label: ", label), utils.GetFuncName(), addrErr)
				continue
			}
			// check address exist on DB
			for _, address := range addressList {
				if utils.IsEmpty(address) {
					continue
				}
				//check address exist on DB
				addressDB, addrErr := utils.GetAddress(address)
				if addrErr != nil && addrErr != orm.ErrNoRows {
					logpack.Error(fmt.Sprintf("Address is not existed on DB. Address: ", address), utils.GetFuncName(), addrErr)
					continue
				}
				//if address is nil, exist is false. if else, exist is true
				addressExist := addressDB != nil
				assetTemp := CheckAssetExistOnArray(assetUpdateArr, user.Id, asset)
				//if addressExist
				if addressExist {
					if assetTemp == nil {
						//get asset
						var getAssetErr error
						assetTemp, getAssetErr = utils.GetAssetById(addressDB.AssetId)
						if getAssetErr != nil {
							logpack.Error(fmt.Sprintf("Get Asset By Id failed. Asset Id: ", addressDB.AssetId), utils.GetFuncName(), getAssetErr)
							continue
						}
						//reset balance and all amount related
						assetTemp.OnChainBalance = 0
						assetTemp.ChainReceived = 0
						assetUpdateArr = append(assetUpdateArr, assetTemp)
						addressDB.ChainReceived = 0
					}
				} else {
					//else, if address does not exist
					//check asset exist on DB
					var assetCheckErr error
					if assetTemp == nil {
						assetTemp, assetCheckErr = utils.GetUserAsset(user.Id, asset)
						if assetCheckErr != nil {
							logpack.Error(fmt.Sprintf("Get user asset failed. UserId: %d, Asset: %s", user.Id, asset), utils.GetFuncName(), assetCheckErr)
							continue
						}
						if assetTemp == nil {
							assetObject := assets.StringToAssetType(asset)
							assetTemp = &models.Asset{
								DisplayName: assetObject.ToFullName(),
								UserId:      user.Id,
								UserName:    user.Username,
								Type:        asset,
								Sort:        assetObject.AssetSortInt(),
								Status:      int(utils.AssetStatusActive),
								Createdt:    time.Now().Unix(),
								Updatedt:    time.Now().Unix(),
							}
							_, insertAssetErr := tx.Insert(assetTemp)
							if insertAssetErr != nil {
								logpack.Error(fmt.Sprintf("Insert asset failed. userid: %d, Asset: %s", user.Id, asset), utils.GetFuncName(), insertAssetErr)
								tx.Rollback()
								return
							}
						}
						assetUpdateArr = append(assetUpdateArr, assetTemp)
					}
					//create new address
					addressDB = &models.Addresses{
						AssetId:      assetTemp.Id,
						Address:      address,
						Transactions: 0,
						Label:        label,
						Createdt:     time.Now().Unix(),
					}
				}
				addressUpdateMap[address] = addressDB
			}
		}
		//Sync transactions
		//get transaction list
		transactionList, transErr := assetObj.GetAllTransactions()
		if transErr != nil {
			logpack.Error(fmt.Sprintf("Get transaction list failed. Asset: %s", asset), utils.GetFuncName(), transErr)
			continue
		}
		//handler transaction
		for _, transaction := range transactionList {
			//if transaction is receive, check receive and update asset
			if transaction.Category == assets.CategoryReceive {
				//check address
				addressFromMap, addrExist := addressUpdateMap[transaction.Address]
				if !addrExist {
					continue
				}
				//get asset by asset id
				assetIndex := GetAssetIndexByIdFromArray(assetUpdateArr, addressFromMap.AssetId)
				if assetIndex < 0 {
					continue
				}
				handlerAsset := assetUpdateArr[assetIndex]
				isConfirmed := utils.IsConfirmed(transaction.Confirmations, asset)
				if isConfirmed {
					handlerAsset.OnChainBalance += transaction.Amount
					handlerAsset.ChainReceived += transaction.Amount
				}

				addressFromMap.ChainReceived += transaction.Amount
				addressFromMap.Transactions++
				handlerAsset.Updatedt = time.Now().Unix()
				addressUpdateMap[transaction.Address] = addressFromMap
				assetUpdateArr[assetIndex] = handlerAsset

				//check on txhistory table and insert
				txHistory, txHistoryErr := utils.GetTxHistoryByTxid(transaction.TxID)
				if txHistoryErr != nil && txHistoryErr != orm.ErrNoRows {
					logpack.Error(fmt.Sprintf("Get TxHistory failed. Txid: %s", transaction.TxID), utils.GetFuncName(), txHistoryErr)
					continue
				}
				//if exist, ignore
				if txHistoryErr == nil {
					continue
				}
				var transactionService services.TransactionService
				price, err := transactionService.GetExchangePrice(asset)
				if err != nil {
					price = 0
				}
				txHistory = &models.TxHistory{
					ReceiverId:    handlerAsset.UserId,
					Receiver:      handlerAsset.UserName,
					ToAddress:     transaction.Address,
					Currency:      asset,
					Amount:        transaction.Amount,
					Rate:          price,
					Txid:          transaction.TxID,
					TransType:     int(utils.TransTypeChainReceive),
					Status:        int(utils.StatusActive),
					Description:   fmt.Sprintf("Received %s from blockchain", strings.ToUpper(asset)),
					Createdt:      transaction.Time,
					Confirmed:     isConfirmed,
					Confirmations: int(transaction.Confirmations),
				}
				_, HistoryErr := tx.Insert(txHistory)
				if HistoryErr != nil {
					logpack.Error(fmt.Sprintf("Insert new history failed: Txid: %s", transaction.TxID), utils.GetFuncName(), HistoryErr)
					continue
				}
			}

			//if transaction is sent, check sent and update asset: updated: 2023/4/22: When use utxo obfusing, can't check sent by comment
			// if transaction.Category == assets.CategorySend {
			// 	//check is external Chain sent
			// 	comment, isChainSent, err := GetSentCommentToChain(transaction.Comment)
			// 	//if is not chain sent
			// 	if err != nil || !isChainSent {
			// 		continue
			// 	}
			// 	//if from on comment is empty, skip
			// 	if utils.IsEmpty(comment.From) || utils.IsEmpty(comment.FromAddress) {
			// 		continue
			// 	}
			// 	//handler for sender
			// 	//check address
			// 	addressFromMap, addrExist := addressUpdateMap[comment.FromAddress]
			// 	if !addrExist {
			// 		continue
			// 	}
			// 	//get asset by asset id
			// 	assetIndex := GetAssetIndexByIdFromArray(assetUpdateArr, addressFromMap.AssetId)
			// 	if assetIndex <= 0 {
			// 		continue
			// 	}
			// 	handlerAsset := assetUpdateArr[assetIndex]
			// 	amount := transaction.Amount
			// 	fee := *transaction.Fee
			// 	if amount < 0 {
			// 		amount = -amount
			// 	}
			// 	if fee < 0 {
			// 		fee = -fee
			// 	}

			// 	handlerAsset.OnChainBalance -= amount + fee
			// 	handlerAsset.ChainSent += amount + fee
			// 	handlerAsset.TotalFee += fee
			// 	handlerAsset.Updatedt = time.Now().Unix()
			// 	assetUpdateArr[assetIndex] = handlerAsset
			// 	continue
			// }
		}

		//handler update or insert asset to DB
		for _, updateAsset := range assetUpdateArr {
			//if update
			if updateAsset.Id > 0 {
				updateAsset.OnChainBalance -= updateAsset.ChainSent
				updateAsset.Balance = updateAsset.LocalReceived - updateAsset.LocalSent + updateAsset.OnChainBalance
				_, updateErr := tx.Update(updateAsset)
				if updateErr != nil {
					tx.Rollback()
					return
				}
				continue
			}
			//else insert asset to DB
			updateAsset.Balance = updateAsset.OnChainBalance
			_, insertErr := tx.Insert(updateAsset)
			if insertErr != nil {
				tx.Rollback()
				return
			}
			continue
		}

		//handler update or insert address to DB
		for _, item := range addressUpdateMap {
			//if update
			if item.Id > 0 {
				_, updateErr := tx.Update(item)
				if updateErr != nil {
					tx.Rollback()
					return
				}
				continue
			}
			//else insert to asset
			_, insertErr := tx.Insert(item)
			if insertErr != nil {
				tx.Rollback()
				return
			}
		}
	}
	tx.Commit()
	logpack.Info("Sync transactions completed", utils.GetFuncName())
}

func IsFromChainReceived(comment string) (bool, error) {
	if utils.IsEmpty(comment) {
		return true, nil
	}
	transNote := assets.TransationNote{}
	err := json.Unmarshal([]byte(comment), &transNote)
	if err != nil {
		return true, err
	}
	return utils.IsEmpty(transNote.From), nil
}

func GetSentCommentToChain(comment string) (*assets.TransationNote, bool, error) {
	if utils.IsEmpty(comment) {
		return nil, false, fmt.Errorf("Failed Transaction")
	}
	transNote := assets.TransationNote{}
	err := json.Unmarshal([]byte(comment), &transNote)
	if err != nil {
		return nil, false, err
	}
	//return comment, is external chain sent, error
	return &transNote, !utils.IsEmpty(transNote.From), nil
}

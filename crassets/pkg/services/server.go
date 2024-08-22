package services

import (
	"context"
	"crassets/pkg/logpack"
	"crassets/pkg/models"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"crassets/pkg/walletlib/assets/btc"
	"crassets/pkg/walletlib/assets/dcr"
	"crassets/pkg/walletlib/assets/ltc"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/rpcclient/v8"
	"gorm.io/gorm"
)

// sync cryptocurrency exchange rate and save to DB
func (s *Server) SyncCryptocurrencyPrice() {
	threadLoop := true
	//Every 7 seconds. Exchange rates of all types are updated once and sent to the session
	go func() {
		for threadLoop {
			//Get allow asset on system settings
			allowCurrencies, allowErr := utils.GetAllowAssetNames(s.Conf.AllowAssets)
			if allowErr != nil {
				continue
			}
			//Get rate of currencies
			usdRateMap, allResultMap, allRateMapErr := s.GetAllMultilPrice(allowCurrencies)
			if allRateMapErr != nil {
				continue
			}
			s.H.WriteRateToDB(usdRateMap, allResultMap)
			time.Sleep(8 * time.Second)
		}
	}()
}

func (s *Server) DecredNotificationsHandler() {
	go func() {
		port := ""
		if utils.GlobalItem.Testnet {
			port = s.Conf.DcrTestnetPort
		} else {
			port = s.Conf.DcrMainnetPort
		}
		host := fmt.Sprintf("%s:%s", s.Conf.DcrRpcHost, port)
		user := s.Conf.DcrRpcUser
		pass := s.Conf.DcrRpcPass
		certFile := filepath.Join(dcrutil.AppDataDir("dcrd", false), "rpc.cert")
		cert, err := os.ReadFile(certFile)
		if err != nil {
			return
		}
		connCfg := &rpcclient.ConnConfig{
			Host:         host,
			Endpoint:     "ws",
			User:         user,
			Pass:         pass,
			Certificates: cert,
		}
		var client *rpcclient.Client
		ntfnHandlers := rpcclient.NotificationHandlers{
			OnTxAccepted: func(hash *chainhash.Hash, amount dcrutil.Amount) {
				if hash == nil {
					return
				}
				time.Sleep(2 * time.Second)
				//get dcr asset manager
				assetMgr, assetExist := utils.GlobalItem.AssetMgrMap[assets.DCRWalletAsset.String()]
				if !assetExist {
					return
				}
				transResult, transErr := assetMgr.GetTransactionByTxhash(hash.String())
				if transErr != nil {
					return
				}
				logpack.Info(fmt.Sprintf("Check txhash from RPC Server: %s", hash.String()), utils.GetFuncName())
				//check txid in DB
				_, txErr := s.H.GetTxHistoryByTxid(hash.String())
				//if txid is existed in txhistory, return
				if txErr == nil {
					logpack.Error("Txid has been processed in txhistory", utils.GetFuncName(), nil)
					return
				}
				//Get received address of transaction
				receivedAddress := ""
				handlerAmount := float64(0)
				for _, detail := range transResult.Details {
					if detail.Category == assets.CategoryReceive {
						receivedAddress = detail.Address
						handlerAmount = detail.Amount
					}
				}
				if utils.IsEmpty(receivedAddress) {
					logpack.Error("Get address of transaction failed", utils.GetFuncName(), nil)
					return
				}
				//get address object
				addressObject, addrErr := s.H.GetAddress(receivedAddress)
				if addrErr != nil {
					logpack.Error("Get address object from receive address failed", utils.GetFuncName(), addrErr)
					return
				}
				//Get asset from receiver address
				asset, assetErr := s.H.GetAssetById(addressObject.AssetId)
				if assetErr != nil {
					logpack.Error("Get asset from receive address failed", utils.GetFuncName(), assetErr)
					return
				}

				asset.Balance += handlerAmount
				asset.OnChainBalance += handlerAmount
				asset.ChainReceived += handlerAmount
				addressObject.ChainReceived += handlerAmount
				asset.Updatedt = time.Now().Unix()
				tx := s.H.DB.Begin()
				//update asset
				assetUpdateErr := tx.Save(asset).Error
				if assetUpdateErr != nil {
					logpack.Error("Update asset error", utils.GetFuncName(), assetUpdateErr)
					tx.Rollback()
					return
				}

				//update address
				addressUpdateErr := tx.Save(addressObject).Error
				if addressUpdateErr != nil {
					logpack.Error("Update address error", utils.GetFuncName(), addressUpdateErr)
					tx.Rollback()
					return
				}
				//Insert to txhistory
				price, err := s.GetExchangePrice(assets.DCRWalletAsset.String())
				if err != nil {
					price = 0
				}
				txHistory := models.TxHistory{}
				txHistory.Receiver = asset.UserName
				txHistory.ToAddress = receivedAddress
				txHistory.Currency = assets.DCRWalletAsset.String()
				txHistory.Amount = handlerAmount
				txHistory.Rate = price
				txHistory.Txid = hash.String()
				txHistory.TransType = int(utils.TransTypeChainReceive)
				txHistory.Status = 1
				txHistory.Description = fmt.Sprintf("Received %s from blockchain", strings.ToUpper(assets.DCRWalletAsset.ToStringUpper()))
				txHistory.Createdt = time.Now().Unix()
				txHistory.Confirmed = utils.IsConfirmed(transResult.Confirmations, assets.DCRWalletAsset.String())
				txHistory.Confirmations = int(transResult.Confirmations)
				HistoryErr := tx.Create(&txHistory).Error
				if HistoryErr != nil {
					logpack.Error("Insert DB txhistory error", utils.GetFuncName(), HistoryErr)
					tx.Rollback()
					return
				}
				tx.Commit()
				logpack.Info("The payment transaction has been processed and stored in the system wallet", utils.GetFuncName())
			},
		}
		client, err = rpcclient.New(connCfg, &ntfnHandlers)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		if err := client.NotifyNewTransactions(ctx, false); err != nil {
			client.Shutdown()
			log.Fatal(err)
		}
	}()
}

func (s *Server) UpdateAssetManagerByType(assetType string) {
	if utils.GlobalItem.AssetMgrMap == nil {
		utils.GlobalItem.AssetMgrMap = make(map[string]assets.Asset)
	}
	_, mgrExist := utils.GlobalItem.AssetMgrMap[assetType]
	if !mgrExist {
		mgr, mgrErr := s.CreateAssetAndConnectRPC(assets.StringToAssetType(assetType))
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

func (s *Server) CreateNewAddressForAsset(username string, isAdmin bool, assetObject assets.AssetType) (*models.Addresses, *models.Asset, error) {
	s.UpdateAssetManagerByType(assetObject.String())
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
		token, _, err := s.H.CheckAndCreateAccountToken(username, role)
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
	userAsset, err := s.H.GetUserAsset(username, assetObject.String())
	if err != nil {
		return nil, nil, fmt.Errorf("Get Asset from DB failed. Please try again!")
	}
	//if user asset is nil, insert new asset to DB
	var assetId int64
	tx := s.H.DB.Begin()
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

func (s *Server) NewAssets(assetType assets.AssetType) (assets.Asset, error) {
	var result assets.Asset
	var err error
	switch assetType {
	case assets.BTCWalletAsset:
		result = s.initBTCAsset()
	case assets.DCRWalletAsset:
		result = s.initDCRAsset()
	case assets.LTCWalletAsset:
		result = s.initLTCAsset()
	default:
		err = fmt.Errorf("Create asset from type failed")
		return nil, err
	}
	return result, nil
}

func (s *Server) initBTCAsset() *btc.Asset {
	asset := &btc.Asset{
		AssetType: assets.BTCWalletAsset.String(),
		Handler:   s.H,
		Config:    s.Conf,
	}
	return asset
}

func (s *Server) initDCRAsset() *dcr.Asset {
	asset := &dcr.Asset{
		AssetType: assets.DCRWalletAsset.String(),
		Handler:   s.H,
		Config:    s.Conf,
	}
	return asset
}

func (s *Server) initLTCAsset() *ltc.Asset {
	asset := &ltc.Asset{
		AssetType: assets.LTCWalletAsset.String(),
		Handler:   s.H,
		Config:    s.Conf,
	}
	return asset
}

func (s *Server) CreateAssetAndConnectRPC(assetType assets.AssetType) (assets.Asset, error) {
	asset, err := s.NewAssets(assetType)
	if err != nil {
		return nil, err
	}
	rpcErr := asset.CreateRPCClient()
	if rpcErr != nil {
		return nil, rpcErr
	}
	asset.SetSystemAddress()
	return asset, nil
}

// Sync blockchain data with system DB data to check data consistency
func (s *Server) SystemSyncHandler() {
	//Get allow asset list
	assetList, err := utils.GetAllowAssetNames(s.Conf.AllowAssets)
	if err != nil {
		logpack.Error("Sync by date error", utils.GetFuncName(), err)
		return
	}
	tx := s.H.DB.Begin()
	for _, asset := range assetList {
		//If is USD, skip
		if asset == assets.USDWalletAsset.String() {
			continue
		}
		//start check and sync
		//Check RPC client
		s.UpdateAssetManagerByType(asset)
		assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
		if !assetMgrExist {
			logpack.Error(fmt.Sprintf("Create %s RPC Client failed", asset), utils.GetFuncName(), nil)
			continue
		}
		//if is decred, separate handler
		if asset == assets.DCRWalletAsset.String() {
			s.SyncDecredHandler(assetObj)
			continue
		}
		//Sync address and account
		//get label list
		labelList := assetObj.GetLabelList()
		addressUpdateMap := make(map[string]*models.Addresses)
		assetUpdateArr := make([]*models.Asset, 0)

		for _, label := range labelList {
			//check label is label of address in system
			account, existLabel := s.H.GetUserFromLabel(label)
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
				addressDB, addrErr := s.H.GetAddress(address)
				if addrErr != nil && addrErr != gorm.ErrRecordNotFound {
					logpack.Error(fmt.Sprintf("Address is not existed on DB. Address: ", address), utils.GetFuncName(), addrErr)
					continue
				}
				//if address is nil, exist is false. if else, exist is true
				addressExist := addressDB != nil
				assetTemp := CheckAssetExistOnArray(assetUpdateArr, account.Username, asset)
				//if addressExist
				if addressExist {
					if assetTemp == nil {
						//get asset
						var getAssetErr error
						assetTemp, getAssetErr = s.H.GetAssetById(addressDB.AssetId)
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
						assetTemp, assetCheckErr = s.H.GetUserAsset(account.Username, asset)
						if assetCheckErr != nil {
							logpack.Error(fmt.Sprintf("Get user asset failed. Username: %s, Asset: %s", account.Username, asset), utils.GetFuncName(), assetCheckErr)
							continue
						}
						if assetTemp == nil {
							assetObject := assets.StringToAssetType(asset)
							assetTemp = &models.Asset{
								DisplayName: assetObject.ToFullName(),
								UserName:    account.Username,
								IsAdmin:     utils.IsSuperAdmin(account.Role),
								Type:        asset,
								Sort:        assetObject.AssetSortInt(),
								Status:      int(utils.AssetStatusActive),
								Createdt:    time.Now().Unix(),
								Updatedt:    time.Now().Unix(),
							}
							insertAssetErr := tx.Create(assetTemp).Error
							if insertAssetErr != nil {
								logpack.Error(fmt.Sprintf("Insert asset failed. username: %s, Asset: %s", account.Username, asset), utils.GetFuncName(), insertAssetErr)
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
				txHistory, txHistoryErr := s.H.GetTxHistoryByTxid(transaction.TxID)
				if txHistoryErr != nil && txHistoryErr != gorm.ErrRecordNotFound {
					logpack.Error(fmt.Sprintf("Get TxHistory failed. Txid: %s", transaction.TxID), utils.GetFuncName(), txHistoryErr)
					continue
				}
				//if exist, ignore
				if txHistoryErr == nil {
					continue
				}
				price, err := s.GetExchangePrice(asset)
				if err != nil {
					price = 0
				}
				txHistory = &models.TxHistory{
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
				historyErr := tx.Create(txHistory).Error
				if historyErr != nil {
					logpack.Error(fmt.Sprintf("Insert new history failed: Txid: %s", transaction.TxID), utils.GetFuncName(), historyErr)
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
				updateErr := tx.Save(updateAsset).Error
				if updateErr != nil {
					tx.Rollback()
					return
				}
				continue
			}
			//else insert asset to DB
			updateAsset.Balance = updateAsset.OnChainBalance
			insertErr := tx.Create(updateAsset).Error
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
				updateErr := tx.Save(item).Error
				if updateErr != nil {
					tx.Rollback()
					return
				}
				continue
			}
			//else insert to asset
			insertErr := tx.Create(item).Error
			if insertErr != nil {
				tx.Rollback()
				return
			}
		}
	}
	tx.Commit()
	logpack.Info("Sync transactions completed", utils.GetFuncName())
}

// check and sync daemon data with DB data
func (s *Server) SyncDecredHandler(assetMgr assets.Asset) {
	// //Get account list with balance
	// balanceMap := assetMgr.GetBalanceMapByLabel()
	// accountAssetMap := make(map[string]*models.Asset)
	// //Address array
	// addressArray := make([]*models.Addresses, 0)
	// //sync account on DB
	// for account, balance := range balanceMap {
	// 	//because account is username. so, check user is exist with username
	// 	user, userErr := GetUserInfoByName(account)
	// 	if userErr != nil {
	// 		continue
	// 	}
	// 	//then, check account exist on asset
	// 	asset, assetErr := utils.GetAssetByOwner(user.Id, o, assets.DCRWalletAsset.String())
	// 	//if error, continue
	// 	if assetErr != nil && assetErr != orm.ErrNoRows {
	// 		continue
	// 	}
	// 	assetObj := assets.DCRWalletAsset
	// 	//if asset not exist, insert to DB
	// 	if assetErr == orm.ErrNoRows {
	// 		asset = &models.Asset{
	// 			DisplayName:    assetObj.ToFullName(),
	// 			UserId:         user.Id,
	// 			UserName:       user.Username,
	// 			IsAdmin:        utils.IsSuperAdmin(user.Role),
	// 			Type:           assetObj.String(),
	// 			Sort:           assetObj.AssetSortInt(),
	// 			Balance:        balance,
	// 			OnChainBalance: balance,
	// 			Status:         int(utils.AssetStatusActive),
	// 			Createdt:       time.Now().Unix(),
	// 			Updatedt:       time.Now().Unix(),
	// 		}
	// 		_, assetInsertErr := tx.Insert(asset)
	// 		if assetInsertErr != nil {
	// 			tx.Rollback()
	// 			continue
	// 		}
	// 	} else {
	// 		if asset.OnChainBalance != balance {
	// 			//update asset
	// 			asset.OnChainBalance = balance
	// 			asset.Updatedt = time.Now().Unix()
	// 			_, assetUpdateErr := tx.Update(asset)
	// 			if assetUpdateErr != nil {
	// 				tx.Rollback()
	// 				continue
	// 			}
	// 		}
	// 	}
	// 	//get address by account
	// 	addressList, addrErr := assetMgr.GetAddressesByLabel(account)
	// 	if addrErr == nil {
	// 		for _, addr := range addressList {
	// 			//check address exist on DB
	// 			checkAddr, checkErr := utils.GetAddress(addr)
	// 			if checkErr != nil && checkErr != orm.ErrNoRows {
	// 				continue
	// 			}
	// 			//if not exist, create new
	// 			if checkAddr == nil && checkErr == orm.ErrNoRows {
	// 				checkAddr = &models.Addresses{
	// 					AssetId:       asset.Id,
	// 					Address:       addr,
	// 					LocalReceived: 0,
	// 					ChainReceived: 0,
	// 					Label:         user.Username,
	// 					Createdt:      time.Now().Unix(),
	// 				}
	// 			}
	// 			addressArray = append(addressArray, checkAddr)
	// 		}
	// 	}
	// 	accountAssetMap[account] = asset
	// }
	// //get all transactions
	// allTransactions, err := assetMgr.GetAllTransactions()
	// if err != nil {
	// 	tx.Commit()
	// 	return
	// }
	// receivedTransactions := make([]assets.ListTransactionsResult, 0)
	// for _, transaction := range allTransactions {
	// 	if transaction.Category == assets.CategoryReceive {
	// 		//check account exist on map
	// 		_, accountExist := accountAssetMap[transaction.Account]
	// 		if accountExist {
	// 			receivedTransactions = append(receivedTransactions, transaction)
	// 		}
	// 	}
	// }
	// //with dcr, only check received transactions
	// for _, receivedTransaction := range receivedTransactions {
	// 	asset, exist := accountAssetMap[receivedTransaction.Account]
	// 	if !exist {
	// 		continue
	// 	}

	// 	addrExistOnArray := false
	// 	var existAddressOnArray *models.Addresses
	// 	var existIndex int
	// 	//check from addressArray
	// 	for index, addr := range addressArray {
	// 		if addr.Address == receivedTransaction.Address {
	// 			addrExistOnArray = true
	// 			existAddressOnArray = addr
	// 			existIndex = index
	// 			break
	// 		}
	// 	}

	// 	//if not exist on array
	// 	if !addrExistOnArray {
	// 		//check on DB
	// 		//check address exist on DB
	// 		addressObj, addrErr := utils.GetAddress(receivedTransaction.Address)
	// 		//if get address error
	// 		if addrErr != nil && addrErr != orm.ErrNoRows {
	// 			continue
	// 		}
	// 		if addrErr == orm.ErrNoRows {
	// 			addressObj = &models.Addresses{
	// 				AssetId:       asset.Id,
	// 				Address:       receivedTransaction.Address,
	// 				ChainReceived: receivedTransaction.Amount,
	// 				Transactions:  1,
	// 				Createdt:      time.Now().Unix(),
	// 			}
	// 			addressArray = append(addressArray, addressObj)
	// 		} else {
	// 			existAddressOnArray = addressObj
	// 		}
	// 	}
	// 	if existAddressOnArray != nil {
	// 		//if assetid on address not match with assetId, re update address
	// 		if existAddressOnArray.AssetId != asset.Id {
	// 			existAddressOnArray.AssetId = asset.Id
	// 		}
	// 		existAddressOnArray.ChainReceived += receivedTransaction.Amount
	// 		existAddressOnArray.Transactions++
	// 		if !addrExistOnArray {
	// 			addressArray = append(addressArray, existAddressOnArray)
	// 		} else {
	// 			addressArray[existIndex] = existAddressOnArray
	// 		}
	// 	}
	// 	//update received of asset
	// 	asset.ChainReceived += receivedTransaction.Amount
	// 	accountAssetMap[receivedTransaction.Account] = asset
	// }

	// //update assets
	// for _, updateAsset := range accountAssetMap {
	// 	if updateAsset.Id > 0 {
	// 		//get received
	// 		receivedAmount, receivedErr := assetMgr.GetReceivedByAccount(updateAsset.UserName)
	// 		if receivedErr == nil {
	// 			updateAsset.ChainReceived = receivedAmount
	// 		}
	// 		//update onchain sent amount
	// 		updateAsset.ChainSent = updateAsset.ChainReceived - updateAsset.OnChainBalance
	// 		updateAsset.Balance = updateAsset.LocalReceived - updateAsset.LocalSent + updateAsset.OnChainBalance
	// 		updateAsset.Updatedt = time.Now().Unix()
	// 		//update
	// 		_, assetUpdateErr := tx.Update(updateAsset)
	// 		if assetUpdateErr != nil {
	// 			tx.Rollback()
	// 			continue
	// 		}
	// 	}
	// }
	// //update or insert address
	// for _, updateAddress := range addressArray {
	// 	//if update
	// 	if updateAddress.Id > 0 {
	// 		_, addrUpdateErr := tx.Update(updateAddress)
	// 		if addrUpdateErr != nil {
	// 			tx.Rollback()
	// 			continue
	// 		}
	// 	} else {
	// 		//if create
	// 		_, addrInsertErr := tx.Insert(updateAddress)
	// 		if addrInsertErr != nil {
	// 			tx.Rollback()
	// 			continue
	// 		}
	// 	}
	// }
}

func CheckAssetExistOnArray(assetList []*models.Asset, username string, assetType string) *models.Asset {
	for _, assetTemp := range assetList {
		if assetTemp.UserName == username && assetTemp.Type == assetType {
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

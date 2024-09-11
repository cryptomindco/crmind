package handler

import (
	"crassets/pkg/db"
	"crassets/pkg/logpack"
	"crassets/pkg/services"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type GlobalHandler struct {
	Server services.Server
}

func (gs *GlobalHandler) UpdateGlobalVariable() {
	if utils.GlobalItem == nil {
		utils.GlobalItem = &utils.Globals{
			CheckedNet:     false,
			Testnet:        false,
			AssetMgrMap:    make(map[string]assets.Asset),
			ExchangeServer: gs.Server.Conf.Exchange,
			PriceSpread:    gs.Server.Conf.PriceSpread,
		}
	}

	//If is not verify net param
	if !utils.GlobalItem.CheckedNet {
		utils.GlobalItem.Testnet = db.TestnetFlg
		utils.GlobalItem.CheckedNet = true
	}

	if utils.GlobalItem.AssetsAllow == nil {
		allowAssets, err := utils.GetAllowAssetNames(gs.Server.Conf.AllowAssets)
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
			mgr, mgrErr := gs.Server.CreateAssetAndConnectRPC(assets.StringToAssetType(asset))
			if mgrErr != nil {
				logpack.Error(fmt.Sprintf("Create asset manager and RPC client for %s failed", asset), utils.GetFuncName(), mgrErr)
			} else {
				utils.GlobalItem.AssetMgrMap[asset] = mgr
			}
		}
	}
}

func (gs *GlobalHandler) GetTotalDaemonBalance(asset string) float64 {
	if asset == assets.USDWalletAsset.String() {
		return 0
	}
	gs.Server.UpdateAssetManagerByType(asset)
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

func (gs *GlobalHandler) GetSpendableAmount(asset string) float64 {
	if asset == assets.USDWalletAsset.String() {
		return 0
	}
	gs.Server.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return 0
	}
	daemonBalance := assetObj.GetSpendableAmount()
	return daemonBalance
}

func (gs *GlobalHandler) SyncTransactionConfirmed() {
	threadLoop := true
	//Every 15 seconds. Update confimed status and
	go func() {
		for threadLoop {
			time.Sleep(15 * time.Second)
			//Get unconfirmed txhistory list
			txHistoryList, txErr := gs.Server.H.GetUnconfirmedTxHistoryList()
			if txErr != nil {
				continue
			}
			tx := gs.Server.H.DB.Begin()
			//browse txHistory and update confimed status
			for _, txHistory := range txHistoryList {
				if txHistory.Confirmed || utils.IsEmpty(txHistory.Txid) {
					continue
				}
				//Get transaction
				assetObj, exist := gs.GetRPCAssetsObject(txHistory.Currency)
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
						if utils.IsEmpty(txHistory.Receiver) {
							continue
						}
						//Get Asset for receiver
						userAsset, assetErr := gs.Server.H.GetUserAsset(txHistory.Receiver, txHistory.Currency)
						if assetErr != nil || userAsset == nil {
							continue
						}
						userAsset.ChainReceived += txHistory.Amount
						userAsset.OnChainBalance += txHistory.Amount
						userAsset.Balance += txHistory.Amount
						userAsset.Updatedt = time.Now().Unix()
						assetUpdateErr := tx.Save(userAsset).Error
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
					txHisErr := tx.Save(&txHistory).Error
					if txHisErr != nil {
						continue
					}
				}
			}
			tx.Commit()
		}
	}()
}

func (gs *GlobalHandler) GetRPCAssetsObject(asset string) (assets.Asset, bool) {
	//Check RPC client
	gs.Server.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return nil, false
	}
	return assetObj, true
}

func (gs *GlobalHandler) SyncSystemData() {
	// init exist transaction list
	//Check time for sync (UTC zone)
	timeSync := gs.Server.Conf.SystemSyncTime
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
			gs.Server.SystemSyncHandler()
			//set today is true
			todaySync = true
			lastSyncDate = &now
			//Check time sync after every 3 minute
			time.Sleep(3 * time.Minute)
		}
	}()
}

// func GetUserInfoByName(username string) (*models.UserInfo, error) {
// 	checkUrl := fmt.Sprintf("%s%s", utils.AuthSite(), "/user-by-name")
// 	req := &dohttp.ReqConfig{
// 		Method:  http.MethodGet,
// 		HttpUrl: checkUrl,
// 		Payload: map[string]string{
// 			"username": username,
// 		},
// 	}
// 	var response utils.ResponseData
// 	err := dohttp.HttpRequest(req, &response)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if response.IsError {
// 		return nil, fmt.Errorf("Get user info by username failed")
// 	}
// 	var userInfo models.UserInfo
// 	err = utils.CatchObject(response.Data, &userInfo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &userInfo, nil
// }

func (gs *GlobalHandler) IsFromChainReceived(comment string) (bool, error) {
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

func (gs *GlobalHandler) GetSentCommentToChain(comment string) (*assets.TransationNote, bool, error) {
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

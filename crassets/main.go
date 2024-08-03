package main

import (
	"context"
	"crassets/handler"
	"crassets/logpack"
	"crassets/models"
	_ "crassets/routers"
	"crassets/services"
	"crassets/utils"
	"crassets/walletlib/assets"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/rpcclient/v8"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	fileName := "logs/assets_op.log"
	var exist bool
	for !exist {
		err := os.MkdirAll("logs", os.ModePerm)
		if err == nil {
			exist = true
		}
	}
	// open log file
	logFile, logErr := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0644)
	if logErr != nil {
		log.Panic(logErr)
	}
	defer logFile.Close()
	// redirect all the output to file
	wrt := io.MultiWriter(os.Stdout, logFile)

	// set log out put
	log.SetOutput(wrt)
	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	beego.BConfig.AppName = "crassets"
	beego.BConfig.Log.AccessLogs = true
	//check and init global variable
	handler.UpdateGlobalVariable()
	SyncCryptocurrencyPrice()
	DecredNotificationsHandler()
	handler.SyncSystemData()
	handler.SyncTransactionConfirmed()
	beego.Run()
}

// sync cryptocurrency exchange rate and save to DB
func SyncCryptocurrencyPrice() {
	threadLoop := true
	//Every 7 seconds. Exchange rates of all types are updated once and sent to the session
	go func() {
		for threadLoop {
			//Get allow asset on system settings
			allowCurrencies, allowErr := utils.GetAllowAssetNames()
			if allowErr != nil {
				continue
			}
			//Get rate of currencies
			var transactionService services.TransactionService
			//usd exchange rate
			usdRateMap, allResultMap, allRateMapErr := transactionService.GetAllMultilPrice(allowCurrencies)
			if allRateMapErr != nil {
				continue
			}
			utils.WriteRateToDB(usdRateMap, allResultMap)
			time.Sleep(7 * time.Second)
		}
	}()
}

func DecredNotificationsHandler() {
	go func() {
		port := ""
		if utils.GlobalItem.Testnet {
			port = beego.AppConfig.String("DCR_TESTNET_PORT")
		} else {
			port = beego.AppConfig.String("DCR_MAINET_PORT")
		}
		host := fmt.Sprintf("%s:%s", beego.AppConfig.String("DCR_RPC_HOST"), port)
		user := beego.AppConfig.String("DCR_RPC_USER")
		pass := beego.AppConfig.String("DCR_RPC_PASS")
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
				o := orm.NewOrm()
				txCount, txErr := o.QueryTable(new(models.TxHistory)).Filter("txid", hash.String()).Count()
				var txExist = txErr == nil && txCount > 0
				//if txid is existed in txhistory, return
				if txExist {
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
				addressObject, addrErr := utils.GetAddress(receivedAddress)
				if addrErr != nil {
					logpack.Error("Get address object from receive address failed", utils.GetFuncName(), addrErr)
					return
				}
				//Get asset from receiver address
				asset, assetErr := utils.GetAssetById(addressObject.AssetId)
				if assetErr != nil {
					logpack.Error("Get asset from receive address failed", utils.GetFuncName(), assetErr)
					return
				}

				asset.Balance += handlerAmount
				asset.OnChainBalance += handlerAmount
				asset.ChainReceived += handlerAmount
				addressObject.ChainReceived += handlerAmount
				asset.Updatedt = time.Now().Unix()
				tx, beginErr := o.Begin()
				if beginErr != nil {
					logpack.Error("Initialize error Transaction DB to insert txhistory", utils.GetFuncName(), beginErr)
					return
				}
				//update asset
				_, assetUpdateErr := tx.Update(asset)
				if assetUpdateErr != nil {
					logpack.Error("Update asset error", utils.GetFuncName(), assetUpdateErr)
					tx.Rollback()
					return
				}

				//update address
				_, addressUpdateErr := tx.Update(addressObject)
				if addressUpdateErr != nil {
					logpack.Error("Update address error", utils.GetFuncName(), addressUpdateErr)
					tx.Rollback()
					return
				}
				//Insert to txhistory
				var transactionService services.TransactionService
				price, err := transactionService.GetExchangePrice(assets.DCRWalletAsset.String())
				if err != nil {
					price = 0
				}
				txHistory := models.TxHistory{}
				txHistory.ReceiverId = asset.UserId
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
				_, HistoryErr := tx.Insert(&txHistory)
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

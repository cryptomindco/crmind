package btc

import (
	"bytes"
	"crassets/pkg/logpack"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcwallet/wallet/txsizes"
	"github.com/jinzhu/copier"
)

type PrivKeyTweaker func(*btcec.PrivateKey) (*btcec.PrivateKey, error)

// Create new RPC Client
func (asset *Asset) CreateRPCClient() error {
	// Connect to local bitcoin core RPC server using HTTP POST mode.
	var port, walletName, walletPassphrase string
	if utils.GlobalItem.Testnet {
		port = asset.Config.BtcTestnetPort
		walletName = asset.Config.BtcSystemTestnetWallet
		walletPassphrase = asset.Config.BtcTestnetWalletPassphrase
		asset.ChainParam = &chaincfg.TestNet3Params
	} else {
		port = asset.Config.BtcMainetPort
		walletName = asset.Config.BtcSystemWallet
		walletPassphrase = asset.Config.BtcWalletPassphrase
		asset.ChainParam = &chaincfg.MainNetParams
	}

	host := asset.Config.BtcRpcHost
	rpcUser := asset.Config.BtcRpcUser
	rpcPass := asset.Config.BtcRpcPass

	connCfg := &rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s:%s", host, port),
		User:         rpcUser,
		Pass:         rpcPass,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return err
	}
	//check wallet loaded
	_, walletErr := client.GetWalletInfo()
	if walletErr != nil {
		_, loadErr := client.LoadWallet(walletName)
		if loadErr != nil {
			//if can't load wallet, create new wallet
			_, createErr := client.CreateWallet(walletName, rpcclient.WithCreateWalletPassphrase(walletPassphrase))
			if createErr != nil {
				logpack.Error(createErr.Error(), utils.GetFuncName(), nil)
				return createErr
			}
			_, loadErr = client.LoadWallet(walletName)
			if loadErr != nil {
				logpack.Error(loadErr.Error(), utils.GetFuncName(), nil)
				return loadErr
			}
		}
		logpack.Info("Load wallet successfully", utils.GetFuncName())
	}
	logpack.Info("Start RPCClient successfully", utils.GetFuncName())
	asset.RpcClient = client
	return nil
}

func (asset *Asset) IsValidAddress(address string) bool {
	_, err := btcutil.DecodeAddress(address, asset.ChainParam)
	return err == nil
}

func (asset *Asset) UpdateLabel(address string, newLabel string) error {
	_, err := asset.CustomRawRequestWithAnyParams("setlabel", address, newLabel)
	if err != nil {
		logpack.Error("Set label error", utils.GetFuncName(), err)
		return err
	}
	return nil
}

func (asset *Asset) GetSystemBalance() (float64, error) {
	amount, err := asset.RpcClient.GetBalance("*")
	if err != nil {
		return 0, err
	}
	return amount.ToBTC(), nil
}

// Shutdown RPC Client
func (asset *Asset) ShutdownRPCClient() {
	asset.RpcClient.Shutdown()
}

func (asset *Asset) MutexLock() {
	asset.RLock()
}

func (asset *Asset) MutexUnlock() {
	asset.RUnlock()
}

func (asset *Asset) GetSpendableAmount() float64 {
	listUnspent, err := asset.ListUnspent()
	if err != nil {
		return float64(0)
	}
	totalAmount := float64(0)
	for _, unspent := range listUnspent {
		totalAmount += unspent.Amount
	}
	return totalAmount
}

// TODO: Create get balance by label
func (asset *Asset) GetBalanceMapByLabel() map[string]float64 {
	return make(map[string]float64)
}

// Get label list
func (asset *Asset) GetLabelList() []string {
	result := make([]string, 0)
	rsl, err := asset.CustomRawRequestWithAnyParams("listlabels")
	if err != nil {
		return result
	}
	json.Unmarshal(rsl, &result)
	return result
}

// Create new address
func (asset *Asset) GetNewAddress(username string) (string, error) {
	res2, err := asset.RpcClient.GetNewAddress("")
	if err != nil {
		return "", err
	}
	return res2.String(), nil
}

// Get Transactions of wallet
func (asset *Asset) GetTransactions(count int, skip int) ([]assets.ListTransactionsResult, error) {
	rawResult, err := asset.CustomRawRequestWithAnyParams("listtransactions", "*", count, skip)
	if err != nil {
		return nil, err
	}

	transList := make([]btcjson.ListTransactionsResult, 0)
	err = json.Unmarshal(rawResult, &transList)
	if err != nil {
		return nil, err
	}

	result := make([]assets.ListTransactionsResult, 0)
	for _, trans := range transList {
		copy := assets.ListTransactionsResult{}
		err := copier.Copy(&copy, trans)
		if err != nil {
			continue
		}
		result = append(result, copy)
	}
	return result, nil
}

func (asset *Asset) GetAllTransactions() ([]assets.ListTransactionsResult, error) {
	//get every 20 transactions
	interval := 20
	from := 0
	result := make([]assets.ListTransactionsResult, 0)
	for {
		//Get transactions
		transList, err := asset.GetTransactions(interval, from)
		if err != nil || len(transList) == 0 {
			logpack.Error(fmt.Sprintf("No transactions from %d to %d", from, from+interval), utils.GetFuncName(), err)
			break
		}
		result = append(result, transList...)
		from += interval
	}
	return result, nil
}

func (asset *Asset) CustomRawRequestWithAnyParams(methodName string, paramStr ...any) (json.RawMessage, error) {
	params := make([]json.RawMessage, 0)
	for _, paramString := range paramStr {
		paramJSON, paramErr := json.Marshal(paramString)
		if paramErr != nil {
			logpack.Error("Parse json for param error", utils.GetFuncName(), paramErr)
			return nil, paramErr
		}
		params = append(params, paramJSON)
	}

	return asset.RpcClient.RawRequest(methodName, params)
}

// Create new address with label settings
func (asset *Asset) CreateNewAddressWithLabel(username, label string) (string, error) {
	//get new address
	newAddr, newErr := asset.RpcClient.GetNewAddress("")
	if newErr != nil {
		return "", newErr
	}

	//Create new label for address
	_, err := asset.CustomRawRequestWithAnyParams("setlabel", newAddr.String(), label)
	if err != nil {
		logpack.Error("Set label error", utils.GetFuncName(), err)
		return "", err
	}
	return newAddr.String(), nil
}

// Get address by label saved on user db
func (asset *Asset) GetAddressesByLabel(label string) ([]string, error) {
	addressList := make([]string, 0)
	//get address by label
	result, err := asset.CustomRawRequestWithAnyParams("getaddressesbylabel", label)
	if err != nil {
		logpack.Error("Get address by label error", utils.GetFuncName(), err)
		return addressList, err
	}
	var resultMap map[string]interface{}
	parseJsonErr := json.Unmarshal(result, &resultMap)
	if parseJsonErr != nil {
		logpack.Error("Parse address result error", utils.GetFuncName(), parseJsonErr)
		return addressList, err
	}
	for k := range resultMap {
		if !utils.IsEmpty(k) {
			addressList = append(addressList, k)
		}
	}
	return addressList, nil
}

func (asset *Asset) DecodeFromAddressString(address string) (btcutil.Address, error) {
	var netParams chaincfg.Params
	if utils.GlobalItem.Testnet {
		netParams = chaincfg.TestNet3Params
	} else {
		netParams = chaincfg.MainNetParams
	}
	btcAddress, err := btcutil.DecodeAddress(address, &netParams)
	if err != nil {
		logpack.Error("Decode address failed. Please try again", utils.GetFuncName(), err)
		return nil, err
	}
	return btcAddress, nil
}

// Get Received By Address
func (asset *Asset) GetReceivedByAddress(address string) (float64, error) {
	addressDecode, decodeErr := asset.DecodeFromAddressString(address)
	if decodeErr != nil {
		return 0, decodeErr
	}
	amount, err := asset.RpcClient.GetReceivedByAddress(addressDecode)
	if err != nil {
		logpack.Error("Get received amount by address failed. Please try again", utils.GetFuncName(), err)
		return 0, err
	}
	return amount.ToBTC(), nil
}

func (asset *Asset) SendToAccountAddress(account, toAddress string, amount float64) (string, error) {
	return "", nil
}

// create new account for user: 1 user - only 1 account
func (asset *Asset) CreateNewAccount(username string) error {
	return nil
}

func (asset *Asset) GetAccount(address string) (string, error) {
	return "", nil
}

func (asset *Asset) GetTransactionFee(txhash string) (float64, error) {
	result, err := asset.GetTransactionByTxhash(txhash)
	if err != nil {
		return 0, err
	}
	return result.Fee, nil
}

func (asset *Asset) GetReceivedByAccount(account string) (float64, error) {
	return 0, nil
}

func (asset *Asset) EstimateFeeAndSize(target *assets.TxTarget) (*assets.TxFeeAndSize, error) {
	sendAmount, err := btcutil.NewAmount(target.Amount)
	if err != nil {
		return nil, err
	}
	if asset.TxAuthoredInfo == nil {
		asset.TxAuthoredInfo = &TxAuthor{}
	}
	unsignedTx, err := asset.unsignedTransaction(target)
	if err != nil {
		return nil, err
	}

	// Since the fee is already calculated when computing the change source out
	// or single destination to send max amount, no need to repeat calculations again.
	feeToSpend := asset.TxAuthoredInfo.TxSpendAmount - sendAmount
	feeAmount := &assets.Amount{
		UnitValue: int64(feeToSpend),
		CoinValue: feeToSpend.ToBTC(),
	}

	var change *assets.Amount
	if unsignedTx.ChangeIndex >= 0 {
		txOut := unsignedTx.Tx.TxOut[unsignedTx.ChangeIndex]
		change = &assets.Amount{
			UnitValue: txOut.Value,
			CoinValue: AmountBTC(txOut.Value),
		}
	}

	// TODO: confirm if the size on UI needs to be in vB to B.
	// This estimation returns size in Bytes (B).
	estimatedSize := txsizes.EstimateSerializeSize(len(unsignedTx.Tx.TxIn), unsignedTx.Tx.TxOut, true)
	// This estimation returns size in virtualBytes (vB).
	// estimatedSize := feeToSpend.ToBTC() / fallBackFeeRate.ToBTC()

	return &assets.TxFeeAndSize{
		FeeRate:             int64(assets.BTCFallbackFeeRatePerkvB),
		EstimatedSignedSize: estimatedSize,
		Fee:                 feeAmount,
		Change:              change,
	}, nil
}

func (asset *Asset) CheckAndLoadWallet() {
	//check wallet loaded
	_, walletErr := asset.RpcClient.GetWalletInfo()
	if walletErr != nil {
		walletName := asset.Config.BtcSystemWallet
		walletPassphrase := asset.Config.BtcWalletPassphrase
		if utils.GlobalItem.Testnet {
			walletName = asset.Config.BtcSystemTestnetWallet
			walletPassphrase = asset.Config.BtcTestnetWalletPassphrase
		}
		_, loadErr := asset.RpcClient.LoadWallet(walletName)
		if loadErr != nil {
			//if can't load wallet, create new wallet
			_, createErr := asset.RpcClient.CreateWallet(walletName, rpcclient.WithCreateWalletPassphrase(walletPassphrase))
			if createErr != nil {
				logpack.Error(createErr.Error(), utils.GetFuncName(), nil)
				return
			}
			_, loadErr = asset.RpcClient.LoadWallet(walletName)
			if loadErr != nil {
				logpack.Error(loadErr.Error(), utils.GetFuncName(), nil)
				return
			}
		}
		logpack.Info("Load wallet successfully", utils.GetFuncName())
	}
}

func (asset *Asset) SendToAddress(from, to, fromAddress, toAddress string, amount float64) (string, error) {
	decodeAddress, decodeErr := asset.DecodeFromAddressString(toAddress)
	if decodeErr != nil {
		return "", decodeErr
	}

	amountObj, amountErr := btcutil.NewAmount(amount)
	if amountErr != nil {
		return "", amountErr
	}
	//input passphrase for wallet. TODO: Handler auth method
	passPhraseErr := asset.RpcClient.WalletPassphrase(asset.GetWalletPassphrase(), 60)
	if passPhraseErr != nil {
		return "", passPhraseErr
	}

	//create transaction Note
	transNote := assets.TransationNote{
		From:        from,
		To:          to,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
	}
	transBytes, err := json.Marshal(transNote)
	if err != nil {
		return "", err

	}

	hash, sendErr := asset.RpcClient.SendToAddressComment(decodeAddress, amountObj, string(transBytes), "")
	if sendErr != nil {
		return "", sendErr
	}

	lockErr := asset.RpcClient.WalletLock()
	if lockErr != nil {
		return "", lockErr
	}
	return hash.String(), nil
}

// Get sending Transaction amount by txID
func (asset *Asset) GetSendTransactionAmount(txhash string) (float64, error) {
	return asset.GetTransactionAmount(txhash, assets.CategorySend)
}

// Get sending Transaction amount by txID
func (asset *Asset) GetReceiveTransactionAmount(txhash string) (float64, error) {
	return asset.GetTransactionAmount(txhash, assets.CategoryReceive)
}

// Get transaction amount by txhash and category. There are both negative and positive values
func (asset *Asset) GetTransactionAmount(txhash string, category string) (float64, error) {
	hash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return 0, err
	}

	transResult, transErr := asset.RpcClient.GetTransaction(hash)
	if transErr != nil {
		logpack.Error(transErr.Error(), utils.GetFuncName(), nil)
		return 0, transErr
	}
	for _, transDetail := range transResult.Details {
		if transDetail.Category == category {
			return transDetail.Amount, nil
		}
	}
	return 0, fmt.Errorf("No sending transaction exists in transaction")
}

// Get wallet passphrase from config
func (asset *Asset) GetWalletPassphrase() string {
	if !utils.GlobalItem.Testnet {
		return asset.Config.BtcWalletPassphrase
	}
	return asset.Config.BtcTestnetWalletPassphrase
}

// Get transaction status from txhash (string) (return: Transation status - Confirmed count - error)
func (asset *Asset) GetTransactionStatus(txhash string) (assets.TransactionStatus, int64, error) {
	hash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return assets.TransactionUnconfirmed, 0, err
	}

	transResult, transErr := asset.RpcClient.GetTransaction(hash)
	if transErr != nil {
		logpack.Error(transErr.Error(), utils.GetFuncName(), nil)
		return assets.TransactionUnconfirmed, 0, transErr
	}

	if transResult.Confirmations >= 6 {
		return assets.TransactionConfirmed, transResult.Confirmations, nil
	}
	return assets.TransactionUnconfirmed, transResult.Confirmations, nil
}

// Get Transaction information by txhash
func (asset *Asset) GetTransactionByTxhash(txhash string) (*assets.TransactionResult, error) {
	hash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return nil, err
	}
	transactionRst, transErr := asset.RpcClient.GetTransaction(hash)
	if transErr != nil {
		return nil, transErr
	}
	result := assets.TransactionResult{}
	copier.Copy(&result, transactionRst)
	return &result, nil
}

// Get Raw Transaction information by txhash
func (asset *Asset) GetRawTransactionByTxhash(txhash string) (*btcjson.TxRawResult, error) {
	hash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return nil, err
	}

	return asset.RpcClient.GetRawTransactionVerbose(hash)
}

// Get received list by address
func (asset *Asset) GetReceivedListByAddress() ([]btcjson.ListReceivedByAddressResult, error) {
	return asset.RpcClient.ListReceivedByAddress()
}

// Create Raw Transactions
func (asset *Asset) CreateRawTransaction(address string, amount float64) (json.RawMessage, error) {
	params := asset.AddToParamArray(nil, asset.CreateMarshalParam(make([]string, 0)))
	mapParam := asset.AddParamToMap(nil, address, amount)
	params = asset.AddToParamArray(params, asset.CreateMarshalParam(mapParam))
	return asset.RpcClient.RawRequest("createrawtransaction", params)
}

// Create Fund Raw Transactions
func (asset *Asset) FundRawTransaction(username string, rawTransactionHex string) (*assets.FundRawTransactionResult, error) {
	params := asset.AddToParamArray(nil, asset.CreateMarshalParam(rawTransactionHex))
	fundRaw, err := asset.RpcClient.RawRequest("fundrawtransaction", params)
	if err != nil {
		return nil, err
	}
	result := assets.FundRawTransactionResult{}
	resErr := json.Unmarshal(fundRaw, &result)
	return &result, resErr
}

func (asset *Asset) AddToParamArray(paramArray []json.RawMessage, param json.RawMessage) []json.RawMessage {
	var resultArray []json.RawMessage
	if paramArray == nil {
		resultArray = make([]json.RawMessage, 0)
	} else {
		resultArray = paramArray
	}
	resultArray = append(resultArray, param)
	return resultArray
}

func (asset *Asset) CreateMarshalParam(value any) json.RawMessage {
	param, err := json.Marshal(value)
	if err != nil {
		param = []byte("")
	}
	return param
}

func (asset *Asset) AddParamToArray(paramArray []any, value any) []any {
	var resultArray []any
	if paramArray == nil {
		resultArray = make([]any, 0)
	} else {
		resultArray = paramArray
	}
	resultArray = append(resultArray, value)
	return resultArray
}

func (asset *Asset) AddParamToMap(paramMap map[string]any, key string, value any) map[string]any {
	var resultMap map[string]any
	if paramMap == nil {
		resultMap = make(map[string]any)
	} else {
		resultMap = paramMap
	}
	resultMap[key] = value
	return resultMap
}

func (asset *Asset) SetSystemAddress() {
	systemAddress, _ := asset.Handler.GetSuperadminSystemAddress(assets.BTCWalletAsset.String())
	asset.SystemAddress = systemAddress
}

func (asset *Asset) GetSystemAddress() string {
	if utils.IsEmpty(asset.SystemAddress) {
		asset.SetSystemAddress()
	}
	return asset.SystemAddress
}

func (asset *Asset) SendToAddressObfuscateUtxos(target *assets.TxTarget) (string, error) {
	//check if system address is empty, return failed
	if utils.IsEmpty(asset.SystemAddress) {
		return "", fmt.Errorf("%s: %s", asset.AssetType, "The system's address could not be found. Please contact superadmin to report errors")
	}
	unsignedTx, err := asset.unsignedTransaction(target)
	if err != nil {
		return "", err
	}
	if unsignedTx.ChangeIndex > 0 {
		unsignedTx.RandomizeChangePosition()
	}
	// Test encode and decode the tx to check its validity after being signed.
	msgTx := unsignedTx.Tx
	lock := make(chan time.Time, 1)
	defer func() {
		lock <- time.Time{}
	}()

	//input passphrase for wallet. TODO: Handler auth method
	passPhraseErr := asset.RpcClient.WalletPassphrase(asset.GetWalletPassphrase(), 60)
	if passPhraseErr != nil {
		return "", passPhraseErr
	}
	bestBlockHeight, bestBlockErr := asset.GetBestBlockHeight()
	if bestBlockErr != nil {
		return "", bestBlockErr
	}
	msgTx.LockTime = uint32(bestBlockHeight)

	//sign transaction
	var complete bool
	msgTx, complete, err = asset.RpcClient.SignRawTransactionWithWallet(msgTx)
	if err != nil {
		return "", err
	}
	if !complete {
		return "", fmt.Errorf("%s", "Not completed transaction")
	}

	var serializedTransaction bytes.Buffer
	serializedTransaction.Grow(msgTx.SerializeSize())
	err = msgTx.Serialize(&serializedTransaction)
	if err != nil {
		logpack.Error(fmt.Sprintf("encoding the tx to test its validity failed: %v", err), utils.GetFuncName(), nil)
		return "", err
	}
	err = msgTx.Deserialize(bytes.NewReader(serializedTransaction.Bytes()))
	if err != nil {
		// Invalid tx
		logpack.Error(fmt.Sprintf("decoding the tx to test its validity failed: %v", err), utils.GetFuncName(), nil)
		return "", err
	}
	hash, err := asset.RpcClient.SendRawTransaction(msgTx, false)
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (asset *Asset) fetchOutputAddr(script []byte) (*btcutil.WIF, error) {
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(script, asset.ChainParam)
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		dumpKey, err := asset.RpcClient.DumpPrivKey(addr)
		if err != nil {
			continue
		}
		return dumpKey, nil
	}

	return nil, fmt.Errorf("Fetch Output Address failed")
}

func (asset *Asset) GetBestBlockHeight() (int32, error) {
	hash, err := asset.RpcClient.GetBestBlockHash()
	if err != nil {
		logpack.Error(fmt.Sprintf("GetBestBlockHash for BTC failed, Err: ", err), utils.GetFuncName(), nil)
		return 0, err
	}

	msgBlock, blockErr := asset.RpcClient.GetBlockVerbose(hash)
	if blockErr != nil {
		logpack.Error(fmt.Sprintf("GetBlock Info for BTC failed, Err: ", err), utils.GetFuncName(), nil)
		return 0, err
	}
	return int32(msgBlock.Height), nil
}

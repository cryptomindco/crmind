package dcr

import (
	"bytes"
	"context"
	"crassets/pkg/logpack"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"decred.org/dcrwallet/v3/rpc/client/dcrwallet"
	"decred.org/dcrwallet/v3/rpc/jsonrpc/types"
	"decred.org/dcrwallet/v4/wallet/txrules"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
	"github.com/jinzhu/copier"
)

// Create new RPC Client
func (asset *Asset) CreateRPCClient() error {
	certFile := filepath.Join(dcrutil.AppDataDir("dcrwallet", false), "rpc.cert")
	cert, err := os.ReadFile(certFile)
	if err != nil {
		return err
	}
	port := ""
	if utils.GlobalItem.Testnet {
		port = asset.Config.DcrWalletTestnetPort
		asset.ChainParam = chaincfg.TestNet3Params()
	} else {
		port = asset.Config.DcrWalletMainnetPort
		asset.ChainParam = chaincfg.MainNetParams()
	}
	user := asset.Config.DcrRpcUser
	pass := asset.Config.DcrRpcPass
	connCfg := &rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s:%s", asset.Config.DcrRpcHost, port),
		Endpoint:     "ws",
		User:         user,
		Pass:         pass,
		Certificates: cert,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return err
	}
	walletClient := dcrwallet.NewClient(dcrwallet.RawRequestCaller(client), asset.ChainParam)
	asset.RpcClient = client
	asset.WalletClient = walletClient
	ctx, _ := asset.GetContextForClient()
	asset.Ctx = ctx
	logpack.Info("Start RPCClient successfully", utils.GetFuncName())
	return nil
}

func (asset *Asset) MutexLock() {
	asset.RLock()
}

func (asset *Asset) MutexUnlock() {
	asset.RUnlock()
}

func (asset *Asset) IsValidAddress(address string) bool {
	_, err := stdaddr.DecodeAddress(address, asset.ChainParam)
	return err == nil
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

// Get balance of system wallet
func (asset *Asset) GetSystemBalance() (float64, error) {
	rawResult, err := asset.CustomRawRequestWithAnyParams("getbalance")
	if err != nil {
		return 0, err
	}
	var balanceRst types.GetBalanceResult
	err = json.Unmarshal(rawResult, &balanceRst)
	if err != nil {
		return 0, err
	}
	totalAmount := float64(0)
	for _, balance := range balanceRst.Balances {
		totalAmount += balance.Total
	}
	return totalAmount, nil
}

func (asset *Asset) GetContextForClient() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// Shutdown RPC Client
func (asset *Asset) ShutdownRPCClient() {
	asset.RpcClient.Shutdown()
}

func (asset *Asset) EstimateFeeAndSize(target *assets.TxTarget) (*assets.TxFeeAndSize, error) {
	if asset.TxAuthoredInfo == nil {
		asset.TxAuthoredInfo = &TxAuthor{}
	}
	unsignedTx, err := asset.unsignedTransaction(target)
	if err != nil {
		return nil, err
	}

	feeToSendTx := txrules.FeeForSerializeSize(txrules.DefaultRelayFeePerKb, unsignedTx.EstimatedSignedSerializeSize)
	feeAmount := &assets.Amount{
		UnitValue: int64(feeToSendTx),
		CoinValue: feeToSendTx.ToCoin(),
	}

	var change *assets.Amount
	if unsignedTx.ChangeIndex >= 0 {
		txOut := unsignedTx.Tx.TxOut[unsignedTx.ChangeIndex]
		change = &assets.Amount{
			UnitValue: txOut.Value,
			CoinValue: Amount(txOut.Value).ToCoin(),
		}
	}

	return &assets.TxFeeAndSize{
		EstimatedSignedSize: unsignedTx.EstimatedSignedSerializeSize,
		Fee:                 feeAmount,
		Change:              change,
	}, nil
}

// Create new address
func (asset *Asset) GetNewAddress(username string) (string, error) {
	res2, err := asset.WalletClient.GetNewAddress(asset.Ctx, username)
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

	transList := make([]types.ListTransactionsResult, 0)
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

	return asset.RpcClient.RawRequest(asset.Ctx, methodName, params)
}

// Create new address with account name (account is username)
func (asset *Asset) CreateNewAddressWithLabel(username, label string) (string, error) {
	unlockErr := asset.UnlockedWallet()
	if unlockErr != nil {
		return "", unlockErr
	}
	//if get account failed, create account
	//check account by check balance of account
	_, balanceErr := asset.WalletClient.GetBalance(asset.Ctx, username)
	if balanceErr != nil {
		//if get balance of account error , create account
		//create new account for username
		accErr := asset.CreateNewAccount(username)
		if accErr != nil {
			return "", accErr
		}
	}
	//get new address
	newAddr, newErr := asset.GetNewAddress(username)
	if newErr != nil {
		return "", newErr
	}
	return newAddr, nil
}

// Get addresses by label saved on user db, with dcr, label is account name
func (asset *Asset) GetAddressesByLabel(label string) ([]string, error) {
	result := make([]string, 0)
	addressList, err := asset.WalletClient.GetAddressesByAccount(asset.Ctx, label)
	if err != nil {
		return nil, err
	}
	for _, addr := range addressList {
		result = append(result, addr.String())
	}
	return result, nil
}

func (asset *Asset) GetReceivedByAccount(account string) (float64, error) {
	amount, err := asset.WalletClient.GetReceivedByAccount(asset.Ctx, account)
	if err != nil {
		return 0, err
	}
	return amount.ToCoin(), nil
}

func (asset *Asset) DecodeFromAddressString(address string) (stdaddr.Address, error) {
	param := chaincfg.MainNetParams()
	if utils.GlobalItem.Testnet {
		param = chaincfg.TestNet3Params()
	}
	addressRst, err := stdaddr.DecodeAddress(address, param)
	if err != nil {
		logpack.Error("Decode address failed. Please try again", utils.GetFuncName(), err)
		return nil, err
	}
	return addressRst, nil
}

// create new account for user: 1 user - only 1 account
func (asset *Asset) CreateNewAccount(username string) error {
	return asset.WalletClient.CreateNewAccount(asset.Ctx, username)
}

func (asset *Asset) GetAccount(address string) (string, error) {
	addr, addrErr := asset.DecodeFromAddressString(address)
	if addrErr != nil {
		return "", addrErr
	}
	account, accErr := asset.WalletClient.GetAccount(asset.Ctx, addr)
	if accErr != nil {
		return "", accErr
	}
	return account, nil
}

func (asset *Asset) CheckAndLoadWallet() {
}

// Get Received By Address
func (asset *Asset) GetReceivedByAddress(address string) (float64, error) {
	// addressDecode, decodeErr := asset.DecodeFromAddressString(address)
	// if decodeErr != nil {
	// 	return 0, decodeErr
	// }
	// amount, err := asset.rpcClient.GetReceivedByAddress(addressDecode)
	// if err != nil {
	// 	logpack.Error("Get received amount by address failed. Please try again", utils.GetFuncName())
	// 	return 0, err
	// }
	// return amount.ToBTC(), nil
	return 0, nil
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

	transResult, transErr := asset.WalletClient.GetTransaction(asset.Ctx, hash)
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
		return asset.Config.DcrWalletPassphrase
	}
	return asset.Config.DcrWalletTestnetPassphrase
}

// Get transaction status from txhash (string) (return: Transation status - Confirmed count - error)
func (asset *Asset) GetTransactionStatus(txhash string) (assets.TransactionStatus, int64, error) {
	hash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return assets.TransactionUnconfirmed, 0, err
	}

	transResult, transErr := asset.WalletClient.GetTransaction(asset.Ctx, hash)
	if transErr != nil {
		logpack.Error(transErr.Error(), utils.GetFuncName(), nil)
		return assets.TransactionUnconfirmed, 0, transErr
	}

	if transResult.Confirmations >= 2 {
		return assets.TransactionConfirmed, transResult.Confirmations, nil
	}
	return assets.TransactionUnconfirmed, transResult.Confirmations, nil
}

func (asset *Asset) UnlockedWallet() error {
	rstl, _ := asset.RpcClient.RawRequest(asset.Ctx, "walletislocked", make([]json.RawMessage, 0))
	isLocked := true
	json.Unmarshal(rstl, &isLocked)
	if isLocked {
		//input passphrase for wallet
		passPhraseErr := asset.WalletClient.WalletPassphrase(asset.Ctx, asset.GetWalletPassphrase(), 60)
		if passPhraseErr != nil {
			return passPhraseErr
		}
	}
	return nil
}

func (asset *Asset) SendToAccountAddress(account, toAddress string, amount float64) (string, error) {
	decodeAddress, decodeErr := asset.DecodeFromAddressString(toAddress)
	if decodeErr != nil {
		return "", decodeErr
	}

	amountObj, amountErr := dcrutil.NewAmount(amount)
	if amountErr != nil {
		return "", amountErr
	}
	//input passphrase for wallet
	passPhraseErr := asset.WalletClient.WalletPassphrase(asset.Ctx, asset.GetWalletPassphrase(), 60)
	if passPhraseErr != nil {
		return "", passPhraseErr
	}

	hash, sendErr := asset.WalletClient.SendFrom(asset.Ctx, account, decodeAddress, amountObj)
	if sendErr != nil {
		return "", sendErr
	}

	lockErr := asset.WalletClient.WalletLock(asset.Ctx)
	if lockErr != nil {
		return "", lockErr
	}
	return hash.String(), nil
}

func (asset *Asset) SendToAddress(from, to, fromAddress, toAddress string, amount float64) (string, error) {
	decodeAddress, decodeErr := asset.DecodeFromAddressString(toAddress)
	if decodeErr != nil {
		return "", decodeErr
	}

	amountObj, amountErr := dcrutil.NewAmount(amount)
	if amountErr != nil {
		return "", amountErr
	}
	//input passphrase for wallet
	passPhraseErr := asset.WalletClient.WalletPassphrase(asset.Ctx, asset.GetWalletPassphrase(), 60)
	if passPhraseErr != nil {
		return "", passPhraseErr
	}
	hash, sendErr := asset.WalletClient.SendToAddress(asset.Ctx, decodeAddress, amountObj)
	if sendErr != nil {
		return "", sendErr
	}

	lockErr := asset.WalletClient.WalletLock(asset.Ctx)
	if lockErr != nil {
		return "", lockErr
	}
	return hash.String(), nil
}

func (asset *Asset) GetTransactionFee(txhash string) (float64, error) {
	result, err := asset.GetTransactionByTxhash(txhash)
	if err != nil {
		return 0, err
	}
	return result.Fee, nil
}

// Get Transaction information by txhash
func (asset *Asset) GetTransactionByTxhash(txhash string) (*assets.TransactionResult, error) {
	hash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return nil, err
	}
	transactionRst, transErr := asset.WalletClient.GetTransaction(asset.Ctx, hash)
	if transErr != nil {
		return nil, transErr
	}
	result := assets.TransactionResult{}
	copier.Copy(&result, transactionRst)
	return &result, nil
}

// Get Raw Transaction information by txhash
func (asset *Asset) GetRawTransactionByTxhash(txhash string) (*dcrjson.TxRawResult, error) {
	_, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		return nil, err
	}

	return nil, nil
}

// with decred, balance by account
func (asset *Asset) GetBalanceMapByLabel() map[string]float64 {
	accountList, err := asset.WalletClient.ListAccounts(asset.Ctx)
	result := make(map[string]float64)
	if err != nil {
		return result
	}
	for key, value := range accountList {
		result[key] = value.ToCoin()
	}
	return result
}

// Get label list (with decred, label is account name)
func (asset *Asset) GetLabelList() []string {
	accountList, err := asset.WalletClient.ListAccounts(asset.Ctx)
	result := make([]string, 0)
	if err != nil {
		return result
	}
	for k, _ := range accountList {
		result = append(result, k)
	}
	return result
}

func (asset *Asset) UpdateLabel(address string, newLabel string) error {
	return nil
}

// Get received list by address
func (asset *Asset) GetReceivedListByAddress() ([]dcrjson.ListReceivedByAddressResult, error) {
	return nil, nil //asset.WalletClient.ListReceivedByAddress(asset.Ctx)
}

// Create Raw Transactions
func (asset *Asset) CreateRawTransaction(address string, amount float64) (json.RawMessage, error) {
	params := asset.AddToParamArray(nil, asset.CreateMarshalParam(make([]string, 0)))
	mapParam := asset.AddParamToMap(nil, address, amount)
	params = asset.AddToParamArray(params, asset.CreateMarshalParam(mapParam))
	return asset.RpcClient.RawRequest(asset.Ctx, "createrawtransaction", params)
}

// Create Fund Raw Transactions
func (asset *Asset) FundRawTransaction(username, rawTransactionHex string) (*assets.FundRawTransactionResult, error) {
	params := asset.AddToParamArray(nil, asset.CreateMarshalParam(rawTransactionHex))
	params = asset.AddToParamArray(params, asset.CreateMarshalParam(username))
	fundRaw, err := asset.RpcClient.RawRequest(asset.Ctx, "fundrawtransaction", params)
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

	var txBuf bytes.Buffer
	txBuf.Grow(unsignedTx.Tx.SerializeSize())
	err = unsignedTx.Tx.Serialize(&txBuf)
	if err != nil {
		return "", err
	}

	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(txBuf.Bytes()))
	if err != nil {
		// Bytes do not represent a valid raw transaction
		return "", err
	}

	lock := make(chan time.Time, 1)
	defer func() {
		lock <- time.Time{}
	}()

	//input passphrase for wallet. TODO: Handler auth method
	unlockErr := asset.UnlockedWallet()
	if unlockErr != nil {
		return "", unlockErr
	}

	ctx, _ := asset.ShutdownContextWithCancel()
	var completed bool
	var signedMsgTx *wire.MsgTx
	signedMsgTx, completed, err = asset.WalletClient.SignRawTransaction(ctx, &msgTx)
	if err != nil {
		return "", err
	}

	if !completed {
		return "", fmt.Errorf("%s", "Not completed transaction")
	}

	var serializedTransaction bytes.Buffer
	serializedTransaction.Grow(signedMsgTx.SerializeSize())
	err = signedMsgTx.Serialize(&serializedTransaction)
	if err != nil {
		return "", err
	}

	err = signedMsgTx.Deserialize(bytes.NewReader(serializedTransaction.Bytes()))
	if err != nil {
		// Invalid tx
		return "", err
	}

	hash, err := asset.RpcClient.SendRawTransaction(ctx, signedMsgTx, false)
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (asset *Asset) SetSystemAddress() {
	systemAddress, _ := asset.Handler.GetSuperadminSystemAddress(assets.DCRWalletAsset.String())
	asset.SystemAddress = systemAddress
}

func (asset *Asset) GetSystemAddress() string {
	if utils.IsEmpty(asset.SystemAddress) {
		asset.SetSystemAddress()
	}
	return asset.SystemAddress
}

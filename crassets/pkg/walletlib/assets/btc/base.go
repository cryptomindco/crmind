package btc

import (
	"crassets/pkg/logpack"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"crassets/pkg/walletlib/assets/helper"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/btcsuite/btcwallet/wallet/txrules"
)

type InputSource func(target btcutil.Amount) (total btcutil.Amount, inputs []*wire.TxIn,
	inputValues []btcutil.Amount, scripts [][]byte, err error)

func (asset *Asset) unsignedTransaction(target *assets.TxTarget) (*txauthor.AuthoredTx, error) {
	unsignedTx, err := asset.constructTransaction(target)
	if err != nil {
		return nil, err
	}
	asset.TxAuthoredInfo.needsConstruct = false
	asset.TxAuthoredInfo.unsignedTx = unsignedTx
	return asset.TxAuthoredInfo.unsignedTx, nil
}

// AmountBTC converts a satoshi amount to a BTC amount.
func AmountBTC(amount int64) float64 {
	return btcutil.Amount(amount).ToBTC()
}

func (asset *Asset) constructTransaction(target *assets.TxTarget) (*txauthor.AuthoredTx, error) {
	var err error
	outputs := make([]*wire.TxOut, 0)
	var changeSource *txauthor.ChangeSource
	output, err := helper.MakeBTCTxOutput(target.ToAddress, target.UnitAmount, asset.ChainParam)
	if err != nil {
		logpack.Error("constructTransaction: error preparing tx output: %v", utils.GetFuncName(), err)
		return nil, fmt.Errorf("make tx output error: %v", err)
	}
	// confirm that the txout will not be rejected on hitting the mempool.
	if err = txrules.CheckOutput(output, assets.BTCFallbackFeeRatePerkvB); err != nil {
		return nil, fmt.Errorf("main txOut validation failed %v", err)
	}
	outputs = append(outputs, output)
	changeSource, err = asset.changeSource(target)
	if err != nil {
		return nil, err
	}
	unspents, currentUserUnspents, err := asset.UnspentOutputs(target)
	if err != nil {
		return nil, err
	}

	inputSource := asset.makeInputSource(unspents, currentUserUnspents)
	unsignedTx, err := txauthor.NewUnsignedTransaction(outputs, assets.BTCFallbackFeeRatePerkvB, inputSource, changeSource)
	if err != nil {
		return nil, fmt.Errorf("creating unsigned tx failed: %v", err)
	}

	if unsignedTx.ChangeIndex == -1 {
		// The change amount is zero or the Txout is likely to be considered as dust
		// if sent to the mempool the whole tx will be rejected.
		return nil, errors.New("adding the change txOut or sendMax tx failed")
	}

	// Confirm that the change output is valid too.
	if err = txrules.CheckOutput(unsignedTx.Tx.TxOut[unsignedTx.ChangeIndex], assets.BTCFallbackFeeRatePerkvB); err != nil {
		return nil, fmt.Errorf("change txOut validation failed %v", err)
	}

	return unsignedTx, nil
}

func (asset *Asset) changeSource(target *assets.TxTarget) (*txauthor.ChangeSource, error) {
	changeSource, err := asset.MakeTxChangeSource(target, asset.ChainParam)
	if err != nil {
		logpack.Error("constructTransaction: error preparing change source", utils.GetFuncName(), err)
		return nil, fmt.Errorf("change source error: %v", err)
	}

	return changeSource, nil
}

// Make refund fee with superadmin address
func (asset *Asset) MakeTxChangeSource(target *assets.TxTarget, net *chaincfg.Params) (*txauthor.ChangeSource, error) {
	var pkScript []byte
	changeSource := &txauthor.ChangeSource{
		NewScript: func() ([]byte, error) {
			pkScript, err := helper.BTCPkScript(asset.SystemAddress, net)
			if err != nil {
				return nil, err
			}
			return pkScript, nil
		},
		ScriptSize: len(pkScript),
	}
	return changeSource, nil
}

func (asset *Asset) UnspentOutputs(target *assets.TxTarget) ([]*assets.UnspentOutput, []*assets.UnspentOutput, error) {
	// Only return UTXOs with the required number of confirmations.
	unspents, err := asset.ListUnspent()
	if err != nil {
		return nil, nil, err
	}
	resp := make([]*assets.UnspentOutput, 0)
	currentUserResp := make([]*assets.UnspentOutput, 0)
	for _, utxo := range unspents {
		// error returned is ignored because the amount value is from upstream
		// and doesn't require an extra layer of validation.
		amount, _ := btcutil.NewAmount(utxo.Amount)
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid TxID %v : error: %v", utxo.TxID, err)
		}
		txInfo, err := asset.RpcClient.GetTransaction(hash)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid TxID %v : error: %v", utxo.TxID, err)
		}

		unspent := &assets.UnspentOutput{
			TxID:          utxo.TxID,
			Vout:          utxo.Vout,
			Address:       utxo.Address,
			ScriptPubKey:  utxo.ScriptPubKey,
			RedeemScript:  utxo.RedeemScript,
			Amount:        Amount(amount),
			Confirmations: int32(utxo.Confirmations),
			Spendable:     utxo.Spendable,
			ReceiveTime:   time.Unix(txInfo.Time, 0),
		}
		//if utxo to address of sender, or spendable is false, ignore
		if utils.ExistStringInArray(target.FromAddresses, utxo.Address) || !utxo.Spendable {
			currentUserResp = append(currentUserResp, unspent)
		} else {
			resp = append(resp, unspent)
		}
	}

	return resp, currentUserResp, nil
}

func (asset *Asset) ListUnspent() ([]btcjson.ListUnspentResult, error) {
	return asset.RpcClient.ListUnspent()
}

func (asset *Asset) makeInputSource(outputs []*assets.UnspentOutput, currentOutputs []*assets.UnspentOutput) txauthor.InputSource {
	var (
		sourceErr       error
		totalInputValue btcutil.Amount
		inputs          = make([]*wire.TxIn, 0, len(outputs))
		inputValues     = make([]btcutil.Amount, 0, len(outputs))
		pkScripts       = make([][]byte, 0, len(outputs))
	)

	// Sorts the outputs in the descending order (utxo with largest amount start)
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Amount.ToCoin() > outputs[j].Amount.ToCoin() })
	sort.Slice(currentOutputs, func(i, j int) bool { return currentOutputs[i].Amount.ToCoin() > currentOutputs[j].Amount.ToCoin() })
	outputs = append(outputs, currentOutputs...)
	// validates the utxo amounts and if an invalid amount is discovered an
	// error is returned.
	for _, output := range outputs {
		// Ignore unspendable utxos
		if !output.Spendable {
			continue
		}

		if output.Amount == nil || output.Amount.ToCoin() == 0 {
			continue
		}

		if !saneOutputValue(output.Amount.(Amount)) {
			sourceErr = fmt.Errorf("impossible output amount `%v` in listunspent result", output.Amount)
			break
		}

		previousOutPoint, err := parseOutPoint(output)
		if err != nil {
			sourceErr = fmt.Errorf("invalid TxIn data found: %v", err)
			break
		}

		script, err := hex.DecodeString(output.ScriptPubKey)
		if err != nil {
			sourceErr = fmt.Errorf("invalid TxIn pkScript data found: %v", err)
			break
		}

		// Determine whether this transaction output is considered dust
		if txrules.IsDustOutput(wire.NewTxOut(output.Amount.ToInt(), script), txrules.DefaultRelayFeePerKb) {
			logpack.Error(fmt.Sprintf("transaction contains a dust output with value: %v", output.Amount.String()), utils.GetFuncName(), nil)
			continue
		}

		totalInputValue += btcutil.Amount(output.Amount.(Amount))
		pkScripts = append(pkScripts, script)
		inputValues = append(inputValues, btcutil.Amount(output.Amount.(Amount)))
		inputs = append(inputs, wire.NewTxIn(previousOutPoint, nil, nil))
	}

	if sourceErr == nil && totalInputValue == 0 {
		// Constructs an error describing the possible reasons why the
		// wallet balance cannot be spent.
		sourceErr = fmt.Errorf("inputs not spendable or have less than %d confirmations")
	}

	return func(target btcutil.Amount) (btcutil.Amount, []*wire.TxIn, []btcutil.Amount, [][]byte, error) {
		// If an error was found return it first.
		if sourceErr != nil {
			return 0, nil, nil, nil, sourceErr
		}
		// This sets the amount the tx will spend if utxos to balance it exists.
		// This spend amount will be crucial in calculating the projected tx fee.
		asset.TxAuthoredInfo.TxSpendAmount = target

		var index int
		var totalUtxo btcutil.Amount
		for _, utxoAmount := range inputValues {
			if totalUtxo < target {
				// Found some utxo(s) we can spend in the current tx.
				index++

				totalUtxo += utxoAmount
				continue
			}
			break
		}
		asset.TxAuthoredInfo.Inputs = inputs[:index]
		asset.TxAuthoredInfo.InputValues = inputValues[:index]
		return totalUtxo, inputs[:index], inputValues[:index], pkScripts[:index], nil
	}
}

func saneOutputValue(amount Amount) bool {
	return amount >= 0 && amount <= btcutil.MaxSatoshi
}

func parseOutPoint(input *assets.UnspentOutput) (*wire.OutPoint, error) {
	txHash, err := chainhash.NewHashFromStr(input.TxID)
	if err != nil {
		return nil, err
	}
	return wire.NewOutPoint(txHash, input.Vout), nil
}

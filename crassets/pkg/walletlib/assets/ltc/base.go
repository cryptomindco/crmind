package ltc

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

	"github.com/ltcsuite/ltcd/btcjson"
	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/chaincfg/chainhash"
	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/ltcsuite/ltcd/wire"
	"github.com/ltcsuite/ltcwallet/wallet/txrules"
)

type InputSource func(target ltcutil.Amount) (total ltcutil.Amount, inputs []*wire.TxIn,
	inputValues []ltcutil.Amount, scripts [][]byte, err error)

func (asset *Asset) unsignedTransaction(target *assets.TxTarget) (*AuthoredTx, error) {
	unsignedTx, err := asset.constructTransaction(target)
	if err != nil {
		return nil, err
	}
	asset.TxAuthoredInfo.needsConstruct = false
	asset.TxAuthoredInfo.unsignedTx = unsignedTx
	return asset.TxAuthoredInfo.unsignedTx, nil
}

// AmountLTC converts a satoshi amount to a BTC amount.
func AmountLTC(amount int64) float64 {
	return ltcutil.Amount(amount).ToBTC()
}

func (asset *Asset) constructTransaction(target *assets.TxTarget) (*AuthoredTx, error) {
	var err error
	outputs := make([]*wire.TxOut, 0)
	var changeSource *ChangeSource
	output, err := helper.MakeLTCTxOutput(target.ToAddress, target.UnitAmount, asset.ChainParam)
	if err != nil {
		logpack.Error("constructTransaction: error preparing tx output: %v", utils.GetFuncName(), err)
		return nil, fmt.Errorf("make tx output error: %v", err)
	}
	// confirm that the txout will not be rejected on hitting the mempool.
	if err = txrules.CheckOutput(output, assets.LTCFallbackFeeRatePerkvB); err != nil {
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
	unsignedTx, err := NewUnsignedTransaction(outputs, assets.LTCFallbackFeeRatePerkvB, inputSource, changeSource)
	if err != nil {
		return nil, fmt.Errorf("creating unsigned tx failed: %v", err)
	}

	if unsignedTx.ChangeIndex == -1 {
		// The change amount is zero or the Txout is likely to be considered as dust
		// if sent to the mempool the whole tx will be rejected.
		return nil, errors.New("adding the change txOut or sendMax tx failed")
	}

	// Confirm that the change output is valid too.
	if err = txrules.CheckOutput(unsignedTx.Tx.TxOut[unsignedTx.ChangeIndex], assets.LTCFallbackFeeRatePerkvB); err != nil {
		return nil, fmt.Errorf("change txOut validation failed %v", err)
	}

	return unsignedTx, nil
}

func (asset *Asset) changeSource(target *assets.TxTarget) (*ChangeSource, error) {
	changeSource, err := asset.MakeTxChangeSource(target, asset.ChainParam)
	if err != nil {
		logpack.Error("constructTransaction: error preparing change source", utils.GetFuncName(), err)
		return nil, fmt.Errorf("change source error: %v", err)
	}

	return changeSource, nil
}

// Make refund fee with superadmin address
func (asset *Asset) MakeTxChangeSource(target *assets.TxTarget, net *chaincfg.Params) (*ChangeSource, error) {
	var pkScript []byte
	changeSource := &ChangeSource{
		NewScript: func() ([]byte, error) {
			pkScript, err := helper.LTCPkScript(asset.SystemAddress, net)
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
		amount, _ := ltcutil.NewAmount(utxo.Amount)
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

func (asset *Asset) makeInputSource(outputs []*assets.UnspentOutput, currentOutputs []*assets.UnspentOutput) InputSource {
	var (
		sourceErr       error
		totalInputValue ltcutil.Amount
		inputs          = make([]*wire.TxIn, 0, len(outputs))
		inputValues     = make([]ltcutil.Amount, 0, len(outputs))
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

		totalInputValue += ltcutil.Amount(output.Amount.(Amount))
		pkScripts = append(pkScripts, script)
		inputValues = append(inputValues, ltcutil.Amount(output.Amount.(Amount)))
		inputs = append(inputs, wire.NewTxIn(previousOutPoint, nil, nil))
	}

	if sourceErr == nil && totalInputValue == 0 {
		// Constructs an error describing the possible reasons why the
		// wallet balance cannot be spent.
		sourceErr = fmt.Errorf("inputs not spendable or have less than %d confirmations")
	}

	return func(target ltcutil.Amount) (ltcutil.Amount, []*wire.TxIn, []ltcutil.Amount, [][]byte, error) {
		// If an error was found return it first.
		if sourceErr != nil {
			return 0, nil, nil, nil, sourceErr
		}
		// This sets the amount the tx will spend if utxos to balance it exists.
		// This spend amount will be crucial in calculating the projected tx fee.
		asset.TxAuthoredInfo.TxSpendAmount = target

		var index int
		var totalUtxo ltcutil.Amount
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
	return amount >= 0 && amount <= ltcutil.MaxSatoshi
}

func parseOutPoint(input *assets.UnspentOutput) (*wire.OutPoint, error) {
	txHash, err := chainhash.NewHashFromStr(input.TxID)
	if err != nil {
		return nil, err
	}
	return wire.NewOutPoint(txHash, input.Vout), nil
}

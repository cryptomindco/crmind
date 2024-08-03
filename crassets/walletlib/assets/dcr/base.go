package dcr

import (
	"context"
	"crassets/logpack"
	"crassets/utils"
	"crassets/walletlib/assets"
	"crassets/walletlib/assets/helper"
	"encoding/hex"
	"fmt"
	"time"

	"decred.org/dcrwallet/v3/rpc/jsonrpc/types"
	"decred.org/dcrwallet/v4/errors"
	"decred.org/dcrwallet/v4/wallet/txauthor"
	"decred.org/dcrwallet/v4/wallet/txrules"
	"decred.org/dcrwallet/v4/wallet/txsizes"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/wire"
)

func (asset *Asset) unsignedTransaction(target *assets.TxTarget) (*txauthor.AuthoredTx, error) {
	if asset.TxAuthoredInfo.needsConstruct || asset.TxAuthoredInfo.unsignedTx == nil {
		unsignedTx, err := asset.constructTransaction(target)
		if err != nil {
			return nil, err
		}

		asset.TxAuthoredInfo.needsConstruct = false
		asset.TxAuthoredInfo.unsignedTx = unsignedTx
	}

	return asset.TxAuthoredInfo.unsignedTx, nil
}

func (wallet *Asset) ShutdownContextWithCancel() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	wallet.cancelFuncs = append(wallet.cancelFuncs, cancel)
	return ctx, cancel
}

// validateSendAmount validate the amount to send to a destination address
func (asset *Asset) validateSendAmount(atomAmount int64) error {
	if atomAmount <= 0 || atomAmount > dcrutil.MaxAmount {
		return errors.E(errors.Invalid, "invalid amount")
	}
	return nil
}

func (asset *Asset) constructTransaction(target *assets.TxTarget) (*txauthor.AuthoredTx, error) {
	var err error
	outputs := make([]*wire.TxOut, 0)
	var changeSource txauthor.ChangeSource
	//check valid amount
	if err := asset.validateSendAmount(target.UnitAmount); err != nil {
		return nil, err
	}

	output, err := helper.MakeDCRTxOutput(target.ToAddress, target.UnitAmount, asset.ChainParam)
	if err != nil {
		logpack.Error(fmt.Sprintf("constructTransaction: error preparing tx output: %v", err), utils.GetFuncName(), err)
		return nil, fmt.Errorf("make tx output error: %v", err)
	}

	outputs = append(outputs, output)
	changeSource, err = helper.MakeDCRTxChangeSource(asset.SystemAddress, asset.ChainParam)
	if err != nil {
		return nil, err
	}

	// if preset with a selected list of UTXOs exists, use them instead.
	unspents, currentUserUnspents, err := asset.UnspentOutputs(target)
	if err != nil {
		return nil, err
	}

	// Use the custom input source function instead of querying the same data from the
	// db for every utxo.
	inputsSourceFunc := asset.makeInputSource(unspents, currentUserUnspents)
	unsignedTx, err := txauthor.NewUnsignedTransaction(outputs, dcrutil.Amount(assets.DCRFallbackFeeRatePerkvB), inputsSourceFunc, changeSource, asset.ChainParam.MaxTxSize)
	if err != nil {
		return nil, fmt.Errorf("creating unsigned tx failed: %v", err)
	}

	if unsignedTx.ChangeIndex == -1 {
		// The change amount is zero or the Txout is likely to be considered as dust
		// if sent to the mempool the whole tx will be rejected.
		return nil, errors.New("adding the change txOut or sendMax tx failed")
	}

	// Confirm that the change output is valid too.
	if err = txrules.CheckOutput(unsignedTx.Tx.TxOut[unsignedTx.ChangeIndex], assets.DCRFallbackFeeRatePerkvB); err != nil {
		return nil, fmt.Errorf("change txOut validation failed %v", err)
	}

	return unsignedTx, nil
}

func (asset *Asset) makeInputSource(utxos []*assets.UnspentOutput, currentOutputs []*assets.UnspentOutput) txauthor.InputSource {
	var (
		sourceErr       error
		totalInputValue dcrutil.Amount

		inputs            = make([]*wire.TxIn, 0, len(utxos))
		pkScripts         = make([][]byte, 0, len(utxos))
		redeemScriptSizes = make([]int, 0, len(utxos))
	)

	utxos = append(utxos, currentOutputs...)
	for _, output := range utxos {
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

		totalInputValue += dcrutil.Amount(output.Amount.(Amount))
		pkScripts = append(pkScripts, script)
		redeemScriptSizes = append(redeemScriptSizes, txsizes.RedeemP2PKHSigScriptSize)
		inputs = append(inputs, wire.NewTxIn(&previousOutPoint, output.Amount.ToInt(), nil))
	}

	if sourceErr == nil && totalInputValue == 0 {
		// Constructs an error describing the possible reasons why the
		// wallet balance cannot be spent.
		sourceErr = fmt.Errorf("inputs have less than %d confirmations", assets.DefaultDCRRequiredConfirmations)
	}

	return func(target dcrutil.Amount) (*txauthor.InputDetail, error) {
		// If an error was found return it first.
		if sourceErr != nil {
			return nil, sourceErr
		}

		inputDetails := &txauthor.InputDetail{}
		var index int
		var currentTotal dcrutil.Amount

		for _, utxoAmount := range inputs {
			if currentTotal < target || target == 0 {
				// Found some utxo(s) we can spend in the current tx.
				index++

				currentTotal += dcrutil.Amount(utxoAmount.ValueIn)
				continue
			}
			break
		}

		inputDetails.Amount = currentTotal
		inputDetails.Inputs = inputs[:index]
		inputDetails.Scripts = pkScripts[:index]
		inputDetails.RedeemScriptSizes = redeemScriptSizes[:index]
		return inputDetails, nil
	}
}

func saneOutputValue(amount Amount) bool {
	return amount >= 0 && amount <= dcrutil.MaxAmount
}

func parseOutPoint(input *assets.UnspentOutput) (wire.OutPoint, error) {
	txHash, err := chainhash.NewHashFromStr(input.TxID)
	if err != nil {
		return wire.OutPoint{}, err
	}
	return wire.OutPoint{Hash: *txHash, Index: input.Vout, Tree: input.Tree}, nil
}

func (asset *Asset) UnspentOutputs(target *assets.TxTarget) ([]*assets.UnspentOutput, []*assets.UnspentOutput, error) {
	// Only return UTXOs with the required number of confirmations.
	unspents, err := asset.ListUnspent()
	if err != nil {
		return nil, nil, err
	}
	resp := make([]*assets.UnspentOutput, 0, len(unspents))
	currentUserResp := make([]*assets.UnspentOutput, 0)
	for _, utxo := range unspents {
		// error returned is ignored because the amount value is from upstream
		// and doesn't require an extra layer of validation.
		amount, _ := dcrutil.NewAmount(utxo.Amount)
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid TxID %v : error: %v", utxo.TxID, err)
		}
		txInfo, err := asset.WalletClient.GetTransaction(asset.Ctx, hash)
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

// Get list unspent. TODO fix support mincof and maxconf
func (asset *Asset) ListUnspent() ([]types.ListUnspentResult, error) {
	return asset.WalletClient.ListUnspent(asset.Ctx)
}

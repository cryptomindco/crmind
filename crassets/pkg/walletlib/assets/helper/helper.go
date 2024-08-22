package helper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
	btctxscript "github.com/btcsuite/btcd/txscript"
	btcWire "github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
	ltcchaincfg "github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/ltcutil"
	ltctxscript "github.com/ltcsuite/ltcd/txscript"
	ltcWire "github.com/ltcsuite/ltcd/wire"
)

const scriptVersion = 0

type TxChangeSource struct {
	// Shared fields.
	script []byte

	// DCR fields.
	version uint16
}

func (src *TxChangeSource) Script() ([]byte, uint16, error) {
	return src.script, src.version, nil
}

func (src *TxChangeSource) ScriptSize() int {
	return len(src.script)
}

func MakeDCRTxChangeSource(destAddr string, net dcrutil.AddressParams) (*TxChangeSource, error) {
	pkScript, err := DCRPkScript(destAddr, net)
	if err != nil {
		return nil, err
	}
	changeSource := &TxChangeSource{
		script:  pkScript,
		version: scriptVersion,
	}
	return changeSource, nil
}

func MakeDCRTxOutput(address string, amountInAtom int64, net dcrutil.AddressParams) (output *wire.TxOut, err error) {
	pkScript, err := DCRPkScript(address, net)
	if err != nil {
		return
	}

	output = &wire.TxOut{
		Value:    amountInAtom,
		Version:  scriptVersion,
		PkScript: pkScript,
	}
	return
}

// returns the public key payment script for the given address.
func DCRPkScript(address string, net dcrutil.AddressParams) ([]byte, error) {
	addr, err := stdaddr.DecodeAddress(address, net)
	if err != nil {
		return nil, fmt.Errorf("error decoding address '%s': %s", address, err.Error())
	}

	_, pkScript := addr.PaymentScript()
	return pkScript, nil
}

func MakeBTCTxOutput(address string, amountInSatoshi int64, net *btcchaincfg.Params) (output *btcWire.TxOut, err error) {
	pkScript, err := BTCPkScript(address, net)
	if err != nil {
		return
	}

	output = &btcWire.TxOut{
		Value:    amountInSatoshi,
		PkScript: pkScript,
	}
	return
}

// returns the public key payment script for the given address.
func BTCPkScript(address string, net *btcchaincfg.Params) ([]byte, error) {
	addr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return nil, fmt.Errorf("error decoding address '%s': %s", address, err.Error())
	}

	// Create a public key script that pays to the address.
	pkScript, err := btctxscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}

	return pkScript, nil
}

func MakeLTCTxOutput(address string, amountInLitoshi int64, net *ltcchaincfg.Params) (output *ltcWire.TxOut, err error) {
	pkScript, err := LTCPkScript(address, net)
	if err != nil {
		return
	}

	output = &ltcWire.TxOut{
		Value:    amountInLitoshi,
		PkScript: pkScript,
	}
	return
}

// returns the public key payment script for the given address.
func LTCPkScript(address string, net *ltcchaincfg.Params) ([]byte, error) {
	addr, err := ltcutil.DecodeAddress(address, net)
	if err != nil {
		return nil, fmt.Errorf("error decoding address '%s': %s", address, err.Error())
	}

	// Create a public key script that pays to the address.
	pkScript, err := ltctxscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}

	return pkScript, nil
}

package dcr

// import (
// 	"fmt"
// 	"io/ioutil"

// 	"github.com/decred/dcrd/rpcclient"
// 	"github.com/prometheus/common/log"
// )

// var requiredChainServerAPI = semver{major: 3, minor: 1, patch: 0}
// var requiredWalletAPI = semver{major: 4, minor: 1, patch: 0}

// func connectWalletRPC(cfg *config) (*rpcclient.Client, semver, error) {
// 	var dcrwCerts []byte
// 	var err error
// 	var walletVer semver
// 	if !cfg.DisableWalletTLS {
// 		dcrwCerts, err = ioutil.ReadFile(cfg.DcrwCert)
// 		if err != nil {
// 			log.Errorf("Failed to read dcrwallet cert file at %s: %s\n",
// 				cfg.DcrwCert, err.Error())
// 			return nil, walletVer, err
// 		}
// 	}

// 	log.Debugf("Attempting to connect to dcrwallet RPC %s as user %s "+
// 		"using certificate located in %s",
// 		cfg.DcrwServ, cfg.DcrwUser, cfg.DcrwCert)

// 	connCfgWallet := &rpcclient.ConnConfig{
// 		Host:         cfg.DcrwServ,
// 		Endpoint:     "ws",
// 		User:         cfg.DcrwUser,
// 		Pass:         cfg.DcrwPass,
// 		Certificates: dcrwCerts,
// 		DisableTLS:   cfg.DisableWalletTLS,
// 	}

// 	ntfnHandlers := getWalletNtfnHandlers(cfg)
// 	dcrwClient, err := rpcclient.New(connCfgWallet, ntfnHandlers)
// 	if err != nil {
// 		log.Errorf("Failed to start dcrwallet RPC client: %s\nPerhaps you"+
// 			" wanted to start with --nostakeinfo?\n", err.Error())
// 		log.Errorf("Verify that rpc.cert is for your wallet:\n\t%v",
// 			cfg.DcrwCert)
// 		return nil, walletVer, err
// 	}

// 	// Ensure the wallet RPC server has a compatible API version.
// 	ver, err := dcrwClient.Version()
// 	if err != nil {
// 		log.Error("Unable to get RPC version: ", err)
// 		return nil, walletVer, fmt.Errorf("Unable to get wallet RPC version")
// 	}

// 	dcrwVer := ver["dcrwalletjsonrpcapi"]
// 	walletVer = semver{dcrwVer.Major, dcrwVer.Minor, dcrwVer.Patch}

// 	if !semverCompatible(requiredWalletAPI, walletVer) {
// 		return nil, walletVer, fmt.Errorf("Wallet JSON-RPC server does not have "+
// 			"a compatible API version. Advertises %v but require %v",
// 			walletVer, requiredWalletAPI)
// 	}

// 	return dcrwClient, walletVer, nil
// }

// func connectNodeRPC(cfg *config) (*rpcclient.Client, semver, error) {
// 	var dcrdCerts []byte
// 	var err error
// 	var nodeVer semver
// 	if !cfg.DisableDaemonTLS {
// 		dcrdCerts, err = ioutil.ReadFile(cfg.DcrdCert)
// 		if err != nil {
// 			log.Errorf("Failed to read dcrd cert file at %s: %s\n",
// 				cfg.DcrdCert, err.Error())
// 			return nil, nodeVer, err
// 		}
// 	}

// 	log.Debugf("Attempting to connect to dcrd RPC %s as user %s "+
// 		"using certificate located in %s",
// 		cfg.DcrdServ, cfg.DcrdUser, cfg.DcrdCert)

// 	connCfgDaemon := &rpcclient.ConnConfig{
// 		Host:         cfg.DcrdServ,
// 		Endpoint:     "ws", // websocket
// 		User:         cfg.DcrdUser,
// 		Pass:         cfg.DcrdPass,
// 		Certificates: dcrdCerts,
// 		DisableTLS:   cfg.DisableDaemonTLS,
// 	}

// 	ntfnHandlers := getNodeNtfnHandlers(cfg)
// 	dcrdClient, err := rpcclient.New(connCfgDaemon, ntfnHandlers)
// 	if err != nil {
// 		log.Errorf("Failed to start dcrd RPC client: %s\n", err.Error())
// 		return nil, nodeVer, err
// 	}

// 	// Ensure the RPC server has a compatible API version.
// 	ver, err := dcrdClient.Version()
// 	if err != nil {
// 		log.Error("Unable to get RPC version: ", err)
// 		return nil, nodeVer, fmt.Errorf("Unable to get node RPC version")
// 	}

// 	dcrdVer := ver["dcrdjsonrpcapi"]
// 	nodeVer = semver{dcrdVer.Major, dcrdVer.Minor, dcrdVer.Patch}

// 	if !semverCompatible(requiredChainServerAPI, nodeVer) {
// 		return nil, nodeVer, fmt.Errorf("Node JSON-RPC server does not have "+
// 			"a compatible API version. Advertises %v but require %v",
// 			nodeVer, requiredChainServerAPI)
// 	}

// 	return dcrdClient, nodeVer, nil
// }

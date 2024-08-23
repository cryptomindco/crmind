package utils

import (
	"crassets/pkg/models"
	"crassets/pkg/walletlib/assets"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	btc_chaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	ltc_chaincfg "github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/ltcutil"
)

var AuthHost, AuthPort string

func IsEmpty(x interface{}) bool {
	switch value := x.(type) {
	case string:
		return value == ""
	case int32:
		return value == 0
	case int:
		return value == 0
	case uint32:
		return value == 0
	case uint64:
		return value == 0
	case int64:
		return value == 0
	case float64:
		return value == 0
	case bool:
		return false
	default:
		return true
	}
}

func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return fmt.Sprintf("%s", runtime.FuncForPC(pc).Name())
}

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ObjectToJsonString(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(b)
}

func ConvertToJsonString(value any) (string, error) {
	outputBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(outputBytes), nil
}

func ExistStringInArray(arrayString []string, checkValue string) bool {
	for _, item := range arrayString {
		if item == checkValue {
			return true
		}
	}
	return false
}

func IsConfirmed(confirmations int64, asset string) bool {
	switch asset {
	case assets.BTCWalletAsset.String(), assets.LTCWalletAsset.String():
		return confirmations >= 6
	case assets.DCRWalletAsset.String():
		return confirmations >= 2
	default:
		return false
	}
}

func CreateDefaultAddressLabelPostfix(assetType string) string {
	return fmt.Sprintf("_%s_address", assetType)
}

func GetAssetRelatedTablePrefix() string {
	if !GlobalItem.Testnet {
		return ""
	}
	return "t_"
}

func GetDateTimeDisplay(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006/01/02, 15:04:05")
}

func CheckValidAddress(assetType, address string) bool {
	var err error
	switch assetType {
	case assets.BTCWalletAsset.String():
		_, err = btcutil.DecodeAddress(address, GetBTCChainParam())
	case assets.LTCWalletAsset.String():
		_, err = ltcutil.DecodeAddress(address, GetLTCChainParam())
	case assets.DCRWalletAsset.String():
		_, err = stdaddr.DecodeAddress(address, GetDCRChainParam())
	default:
		return false
	}
	return err == nil
}

func GetBTCChainParam() *btc_chaincfg.Params {
	if GlobalItem.Testnet {
		return &btc_chaincfg.TestNet3Params
	}
	return &btc_chaincfg.MainNetParams
}

func GetLTCChainParam() *ltc_chaincfg.Params {
	if GlobalItem.Testnet {
		return &ltc_chaincfg.TestNet4Params
	}
	return &ltc_chaincfg.MainNetParams
}

func GetDCRChainParam() *chaincfg.Params {
	if GlobalItem.Testnet {
		return chaincfg.TestNet3Params()
	}
	return chaincfg.MainNetParams()
}

func GetUnitAmount(amount float64, asset string) (int64, error) {
	switch asset {
	case assets.DCRWalletAsset.String():
		amount, err := dcrutil.NewAmount(amount)
		return int64(amount), err
	case assets.BTCWalletAsset.String():
		amount, err := btcutil.NewAmount(amount)
		return int64(amount), err
	case assets.LTCWalletAsset.String():
		amount, err := ltcutil.NewAmount(amount)
		return int64(amount), err
	default:
		return 0, fmt.Errorf("No Asset was found")
	}
}

func CatchObject(from interface{}, to interface{}) error {
	jsonBytes, err := json.Marshal(from)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBytes, &to)
	if err != nil {
		return err
	}
	return nil
}

func CheckUserExistOnContactList(username string, contactList []models.ContactItem) bool {
	if len(contactList) == 0 {
		return false
	}
	for _, contact := range contactList {
		if contact.UserName == username {
			return true
		}
	}
	return false
}

func IsSuperAdmin(role int) bool {
	return role == int(RoleSuperAdmin)
}

func GetAllowAssetNames(allowAssets string) ([]string, error) {
	return GetAssetsNameFromStr(allowAssets), nil
}

func GetAssetsNameFromStr(input string) []string {
	if IsEmpty(input) {
		return []string{assets.USDWalletAsset.String()}
	}
	result := make([]string, 0)
	assetArr := strings.Split(input, ",")
	for _, asset := range assetArr {
		assetObj := assets.StringToAssetType(strings.TrimSpace(asset))
		if assetObj != assets.NilAsset {
			result = append(result, assetObj.String())
		}
	}
	if len(result) == 0 {
		result = append(result, assets.USDWalletAsset.String())
	}
	return result
}

func CheckUsernameExistOnContactList(username string, contactList []models.ContactItem) bool {
	if len(contactList) == 0 {
		return false
	}
	for _, contact := range contactList {
		if contact.UserName == username {
			return true
		}
	}
	return false
}

func IsCryptoCurrency(assetType string) bool {
	var lowercaseType = strings.ToLower(assetType)
	return lowercaseType == assets.BTCWalletAsset.String() || lowercaseType == assets.DCRWalletAsset.String() || lowercaseType == assets.LTCWalletAsset.String()
}

func JsonStringToObject(jsonString string, to interface{}) error {
	err := json.Unmarshal([]byte(jsonString), &to)
	if err != nil {
		return err
	}
	return nil
}

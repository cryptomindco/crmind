package utils

import (
	"crmind/models"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/adapter"
)

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

func RequestBodyToString(body io.ReadCloser) string {
	b, err := io.ReadAll(body)
	if err != nil {
		return ""
	}
	return string(b)
}

func ObjectToJsonString(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func JsonStringToObject(jsonString string, to interface{}) error {
	err := json.Unmarshal([]byte(jsonString), &to)
	if err != nil {
		return err
	}
	return nil
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

func GetAuthHost() string {
	if !IsEmpty(AuthHost) {
		return AuthHost
	}
	AuthHost = beego.AppConfig.String("AUTH_SERVICE_URL")
	return AuthHost
}

func GetChatHost() string {
	if !IsEmpty(ChatHost) {
		return ChatHost
	}
	ChatHost = beego.AppConfig.String("CHAT_SERVICE_URL")
	return ChatHost
}

func GetAssetsHost() string {
	if !IsEmpty(AssetsHost) {
		return AssetsHost
	}
	AssetsHost = beego.AppConfig.String("ASSETS_SERVICE_URL")
	return AssetsHost
}

func GetUrlCodeStatusFromValue(status int) (UrlCodeStatus, error) {
	switch status {
	case int(UrlCodeStatusCreated):
		return UrlCodeStatusCreated, nil
	case int(UrlCodeStatusConfirmed):
		return UrlCodeStatusConfirmed, nil
	case int(UrlCodeStatusCancelled):
		return UrlCodeStatusCancelled, nil
	default:
		return -1, fmt.Errorf("Get Url Code error with status: %d", status)
	}
}

func (urlCodeStatus UrlCodeStatus) CodeStatusColor() string {
	switch urlCodeStatus {
	case UrlCodeStatusCreated:
		return "#dbc272"
	case UrlCodeStatusConfirmed:
		return "#008000"
	case UrlCodeStatusCancelled:
		return "#dc3545"
	default:
		return ""
	}
}

func ConvertToJsonString(value any) (string, error) {
	outputBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(outputBytes), nil
}

func StringToAssetType(assetType string) AssetType {
	switch assetType {
	case "usd":
		return USDWalletAsset
	case "btc":
		return BTCWalletAsset
	case "dcr":
		return DCRWalletAsset
	case "ltc":
		return LTCWalletAsset
	default:
		return NilAsset
	}
}

func GetAssetsNameFromStr(input string) []string {
	if IsEmpty(input) {
		return []string{"usd"}
	}
	assetArr := strings.Split(input, ",")
	return assetArr
}

func CreateNewAsset(assetType string, username string) *models.Asset {
	return &models.Asset{
		Sort:          AssetSortInt(assetType),
		DisplayName:   GetAssetFullName(assetType),
		UserName:      username,
		Type:          assetType,
		Balance:       0,
		LocalReceived: 0,
		LocalSent:     0,
		ChainReceived: 0,
		ChainSent:     0,
		TotalFee:      0,
	}
}

func GetAllowAssets() string {
	if IsEmpty(AllowAssets) {
		allowAssetsStr, err := GetAssetStrFromSettings()
		if err != nil {
			return "usd"
		}
		return allowAssetsStr
	}
	return AllowAssets
}

func GetDateTimeDisplay(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006/01/02, 15:04:05")
}

func CheckActiveService(checkService string) bool {
	if IsEmpty(ActiveServices) {
		ActiveServices, _ = GetServicesStrFromSettings()
	}
	services := HandlerActiveServiceStr(ActiveServices)
	for _, service := range services {
		if service == checkService {
			return true
		}
	}
	return false
}

func IsAuthActive() bool {
	return CheckActiveService(string(AuthService))
}

func IsChatActive() bool {
	return CheckActiveService(string(ChatService))
}

func IsAssetsActive() bool {
	return CheckActiveService(string(AssetsService))
}

func GetServiceUrl(service string) []string {
	switch service {
	case string(ChatService):
		return ChatUrl
	case string(AssetsService):
		return AssetUrl
	default:
		return []string{}
	}
}

func CheckServiceValidUrl(url string) bool {
	for _, service := range ServiceList {
		urlList := GetServiceUrl(service)
		if len(urlList) == 0 {
			continue
		}
		isUrl := false
		for _, tmpUrl := range urlList {
			if tmpUrl == url {
				isUrl = true
				break
			}
		}
		if isUrl {
			return CheckActiveService(service)
		}
	}
	return true
}

// handler error from rpc
func HandlerRPCError(err error) error {
	if err == nil {
		return nil
	}
	errString := err.Error()
	if strings.HasPrefix(errString, "rpc error:") {
		errSplitArr := strings.Split(errString, "desc =")
		if len(errSplitArr) < 2 {
			return err
		}
		errMsg := strings.TrimSpace(errSplitArr[1])
		return fmt.Errorf("%s", errMsg)
	}
	return err
}

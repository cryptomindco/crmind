package utils

import (
	"encoding/json"
	"fmt"
	"runtime"

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

func ObjectToJsonString(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(b)
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
	AuthHost = beego.AppConfig.String("authhost")
	return AuthHost
}

func GetAuthPort() string {
	if !IsEmpty(AuthPort) {
		return AuthPort
	}
	AuthPort = beego.AppConfig.String("authport")
	return AuthPort
}

func GetAssetsHost() string {
	if !IsEmpty(AssetsHost) {
		return AssetsHost
	}
	AssetsHost = beego.AppConfig.String("assethost")
	return AssetsHost
}

func AuthSite() string {
	return fmt.Sprintf("%s:%s", GetAuthHost(), GetAuthPort())
}

func GetAssetsPort() string {
	if !IsEmpty(AssetsPort) {
		return AssetsPort
	}
	AssetsPort = beego.AppConfig.String("assetport")
	return AssetsPort
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

package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"

	beego "github.com/beego/beego/v2/adapter"
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

func GetConfValue(key string) string {
	return beego.AppConfig.String(key)
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

func AuthSite() string {
	return fmt.Sprintf("%s:%s", GetAuthHost(), GetAuthPort())
}

func ConvertToJsonString(value any) (string, error) {
	outputBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(outputBytes), nil
}

package logpack

import (
	"crchat/pkg/utils"
	"fmt"
	"log"
)

// Simple Info Msg
func Info(msg string, funcName string) {
	log.Println("INFO:", funcName, "-", msg)
}

// Simple Info Msg
func Error(msg string, funcName string, err error) {
	errStr := ""
	if err != nil {
		errStr = fmt.Sprintf("\n%s", err.Error())
	}
	log.Println("ERROR:", funcName, "-", msg, errStr)
}

// Simple Info Msg
func Warn(msg string, funcName string) {
	log.Println("WARNING:", funcName, "-", msg)
}

// Simple Info Msg
func Fatal(msg string, funcName string) {
	log.Println("FATAL:", funcName, "-", msg)
}

func FInfo(msg string, loginName string, funcName string) {
	log.Println("INFO:", funcName, GetLoginUsername(loginName), "-", msg)
}

func FError(msg string, loginName string, funcName string, err error) {
	errStr := ""
	if err != nil {
		errStr = fmt.Sprintf("\n%s", err.Error())
	}
	log.Println("ERROR:", funcName, GetLoginUsername(loginName), "-", msg, errStr)
}

func FWarn(msg string, loginName string, funcName string) {
	log.Println("WARNING:", funcName, GetLoginUsername(loginName), "-", msg)
}

func FFatal(msg string, loginName string, funcName string) {
	log.Println("FATAL:", funcName, GetLoginUsername(loginName), "-", msg)
}

func GetLoginUsername(loginName string) string {
	if utils.IsEmpty(loginName) {
		return ""
	}
	return fmt.Sprintf(", LoginName: %s", loginName)
}

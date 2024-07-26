package logpack

import (
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

func FInfo(msg string, loginId int64, funcName string) {
	log.Println("INFO:", funcName, GetLoginUserId(loginId), "-", msg)
}

func FError(msg string, loginId int64, funcName string, err error) {
	errStr := ""
	if err != nil {
		errStr = fmt.Sprintf("\n%s", err.Error())
	}
	log.Println("ERROR:", funcName, GetLoginUserId(loginId), "-", msg, errStr)
}

func FWarn(msg string, loginId int64, funcName string) {
	log.Println("WARNING:", funcName, GetLoginUserId(loginId), "-", msg)
}

func FFatal(msg string, loginId int64, funcName string) {
	log.Println("FATAL:", funcName, GetLoginUserId(loginId), "-", msg)
}

func GetLoginUserId(loginId int64) string {
	if loginId <= 0 {
		return ""
	}
	return fmt.Sprintf(", LoginId: %d", loginId)
}

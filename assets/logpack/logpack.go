package logpack

import (
	"assets/models"
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

func FInfo(msg string, loginUser *models.User, funcName string) {
	log.Println("INFO:", funcName, GetLoginUserId(loginUser), "-", msg)
}

func FError(msg string, loginUser *models.User, funcName string, err error) {
	errStr := ""
	if err != nil {
		errStr = fmt.Sprintf("\n%s", err.Error())
	}
	log.Println("ERROR:", funcName, GetLoginUserId(loginUser), "-", msg, errStr)
}

func FWarn(msg string, loginUser *models.User, funcName string) {
	log.Println("WARNING:", funcName, GetLoginUserId(loginUser), "-", msg)
}

func FFatal(msg string, loginUser *models.User, funcName string) {
	log.Println("FATAL:", funcName, GetLoginUserId(loginUser), "-", msg)
}

func GetLoginUserId(loginUser *models.User) string {
	if loginUser == nil {
		return ""
	}
	return fmt.Sprintf(", LoginId: %d", loginUser.Id)
}

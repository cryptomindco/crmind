package main

import (
	"auth/logpack"
	"auth/passkey"
	_ "auth/routers"
	"auth/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
	"github.com/go-webauthn/webauthn/webauthn"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	fileName := "logs/auth_op.log"
	var exist bool
	for !exist {
		err := os.MkdirAll("logs", os.ModePerm)
		if err == nil {
			exist = true
		}
	}
	// open log file
	logFile, logErr := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0644)
	if logErr != nil {
		log.Panic(logErr)
	}
	defer logFile.Close()
	// redirect all the output to file
	wrt := io.MultiWriter(os.Stdout, logFile)

	// set log out put
	log.SetOutput(wrt)
	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	beego.BConfig.AppName = "crauth"
	beego.BConfig.Log.AccessLogs = true
	InitForWebAuthn()
	//handler for web authn
	beego.Run()
}

func InitForWebAuthn() {
	origin := beego.AppConfig.String("passkeyhost")
	host := beego.AppConfig.String("originalHost")
	logpack.Info("make webauthn config", utils.GetFuncName())
	wconfig := &webauthn.Config{
		RPDisplayName: "Cryptomind authentication", // Display Name for your site
		RPID:          host,                        // Generally the FQDN for your site
		RPOrigins:     []string{origin},            // The origin URLs allowed for WebAuthn
	}
	logpack.Info("create webauthn", utils.GetFuncName())
	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		fmt.Printf("[FATA] %s", err.Error())
		os.Exit(1)
	}

	passkey.WebAuthn = webAuthn

	logpack.Info("create datastore", utils.GetFuncName())
	passkey.Datastore = passkey.NewInMem()
	//get list user with credential info
	hasCredUserList, err := utils.GetHasCredUserList()
	if err != nil && err != orm.ErrNoRows {
		//Get Credential info failed, panic
		fmt.Printf("[FATA] %s", err.Error())
		os.Exit(1)
	}
	//save data for Datastore
	for _, credUser := range hasCredUserList {
		username := credUser.Username
		credsJson := credUser.CredsArrJson
		if utils.IsEmpty(username) || utils.IsEmpty(credsJson) {
			continue
		}
		var creds []webauthn.Credential
		err := json.Unmarshal([]byte(credsJson), &creds)
		if err != nil {
			fmt.Printf("[FATA] %s", err.Error())
			os.Exit(1)
		}
		insertUserPasskey := &passkey.User{
			ID:       []byte(username),
			Username: username,
		}
		insertUserPasskey.SetCredential(creds)
		passkey.Datastore.InsertPasskeyUser(username, *insertUserPasskey)
	}
}

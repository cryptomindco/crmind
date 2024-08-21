package main

import (
	"crauth/pkg/config"
	"crauth/pkg/db"
	"crauth/pkg/logpack"
	"crauth/pkg/passkey"
	"crauth/pkg/pb"
	"crauth/pkg/services"
	"crauth/pkg/utils"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	"log"
	"net"

	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc"
	"gorm.io/gorm"
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
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("failed at config ", err)
	}
	logpack.Info(fmt.Sprintf("DB URL: %s", c.DBUrl), utils.GetFuncName())
	h := db.Init(c)
	aliveSessionHourStr := c.AliveSessionHours
	aliveSessionHours, err := strconv.ParseInt(aliveSessionHourStr, 0, 32)
	if err != nil {
		aliveSessionHours = utils.AliveSessionHours
	}
	jwt := utils.JWTWrapper{
		SecretKey:       c.JWTSecretKey,
		Issuer:          "go-grpc-crauth",
		ExpirationHours: aliveSessionHours,
	}

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("failed at listening : ", err)
	}
	fmt.Println("Auth svc on ", c.Port)
	s := services.Server{
		H:   h,
		Jwt: jwt,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
	InitForWebAuthn(c, h)
}

func InitForWebAuthn(c config.Config, hanlder db.Handler) {
	origin := c.PasskeyHost
	host := c.OriginalHost
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
	hasCredUserList, err := hanlder.GetHasCredUserList()
	if err != nil && err != gorm.ErrRecordNotFound {
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

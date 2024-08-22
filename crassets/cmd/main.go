package main

import (
	"crassets/pkg/config"
	"crassets/pkg/db"
	"crassets/pkg/handler"
	"crassets/pkg/logpack"
	"crassets/pkg/pb"
	"crassets/pkg/services"
	"crassets/pkg/utils"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"log"
	"net"

	"google.golang.org/grpc"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	fileName := "logs/assets_op.log"
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
	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("failed at listening : ", err)
	}
	fmt.Println("Auth svc on ", c.Port)
	s := services.Server{
		H:    h,
		Conf: c,
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAssetsServiceServer(grpcServer, &s)
	handler := handler.GlobalHandler{
		Server: s,
	}
	//handler some task
	handler.UpdateGlobalVariable()
	s.SyncCryptocurrencyPrice()
	s.DecredNotificationsHandler()
	handler.SyncSystemData()
	handler.SyncTransactionConfirmed()
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}

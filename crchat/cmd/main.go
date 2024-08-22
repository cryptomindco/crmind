package main

import (
	"crchat/pkg/config"
	"crchat/pkg/db"
	"crchat/pkg/logpack"
	"crchat/pkg/pb"
	"crchat/pkg/services"
	"crchat/pkg/utils"
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
	fileName := "logs/chat_op.log"
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
		H: h,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterChatServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}

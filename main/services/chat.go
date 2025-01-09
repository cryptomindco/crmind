package services

import (
	"context"
	"crmind/pb/chatpb"
	"crmind/utils"
	"fmt"

	"google.golang.org/grpc"
)

var ChatCli ChatClient

type ChatClient struct {
	Client *chatpb.ChatServiceClient
}

func InitChatClient() *chatpb.ChatServiceClient {
	chatURL := utils.GetChatHost()
	fmt.Println("API Gateway :  InitChatClient")
	//	using WithInsecure() because no SSL running
	cc, err := grpc.Dial(chatURL, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect:", err)
		return nil
	}
	client := chatpb.NewChatServiceClient(cc)
	return &client
}

func CheckAndInitChatClient() error {
	if ChatCli.Client != nil {
		return nil
	}
	ChatCli.Client = InitChatClient()
	if ChatCli.Client == nil {
		return fmt.Errorf("Init chat client failed")
	}
	return nil
}

func UpdateUnreadForChatHandler(ctx context.Context, req *chatpb.UpdateUnreadForChatRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).UpdateUnreadForChat(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func DeleteChatHandler(ctx context.Context, req *chatpb.DeleteChatRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).DeleteChat(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func CheckAndCreateChatHandler(ctx context.Context, req *chatpb.CheckAndCreateChatRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).CheckAndCreateChat(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func SendChatMessageHandler(ctx context.Context, req *chatpb.SendChatMessageRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).SendChatMessage(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func CheckChatExistHandler(ctx context.Context, req *chatpb.CheckChatExistRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).CheckChatExist(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func GetChatMsgDisplayListHandler(ctx context.Context, req *chatpb.CommonRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).GetChatMsgDisplayList(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func GetChatMsgHandler(ctx context.Context, req *chatpb.GetChatMsgRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).GetChatMsg(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func CreateHelloChatHandler(ctx context.Context, req *chatpb.CreateHelloChatRequest) (*chatpb.ResponseData, error) {
	err := CheckAndInitChatClient()
	if err != nil {
		return nil, err
	}
	res, err := (*ChatCli.Client).CreateHelloChat(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

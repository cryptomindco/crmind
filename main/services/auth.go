package services

import (
	"context"
	"crmind/pb/authpb"
	"crmind/utils"
	"fmt"

	"google.golang.org/grpc"
)

var AuthCli AuthClient

type AuthClient struct {
	Client *authpb.AuthServiceClient
}

func InitAuthClient() *authpb.AuthServiceClient {
	authURL := utils.GetAuthHost()
	fmt.Println("API Gateway :  InitAuthClient")
	//	using WithInsecure() because no SSL running
	cc, err := grpc.Dial(authURL, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect to auth service:", err)
		return nil
	}
	client := authpb.NewAuthServiceClient(cc)
	return &client
}

func CheckAndInitAuthClient() error {
	if AuthCli.Client != nil {
		return nil
	}
	AuthCli.Client = InitAuthClient()
	if AuthCli.Client == nil {
		return fmt.Errorf("init auth client failed")
	}
	return nil
}

func CheckMiddlewareLogin(ctx context.Context, req *authpb.CommonRequest) (bool, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return false, err
	}
	_, err = (*AuthCli.Client).IsLoggingOn(ctx, req)
	if err != nil {
		return false, err
	}
	return true, nil
}

func BeginRegistrationHandler(ctx context.Context, req *authpb.WithUsernameRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).BeginRegistration(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CancelRegisterHandler(ctx context.Context, req *authpb.CancelRegisterRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).CancelRegister(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func BeginUpdatePasskeyHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).BeginUpdatePasskey(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FinishUpdatePasskeyHandler(ctx context.Context, req *authpb.FinishUpdatePasskeyRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).FinishUpdatePasskey(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FinishRegistrationHandler(ctx context.Context, req *authpb.SessionKeyAndHttpRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).FinishRegistration(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func AssertionOptionsHandler(ctx context.Context) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).AssertionOptions(ctx, &authpb.CommonRequest{})
	if err != nil {
		return res, err
	}
	return res, nil
}

func AssertionResultHandler(ctx context.Context, req *authpb.SessionKeyAndHttpRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).AssertionResult(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func BeginConfirmPasskeyHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).BeginConfirmPasskey(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FinishConfirmPasskeyHandler(ctx context.Context, req *authpb.SessionKeyAndHttpRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).FinishConfirmPasskey(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func ChangeUsernameFinishHandler(ctx context.Context, req *authpb.ChangeUsernameFinishRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).ChangeUsernameFinish(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func SyncUsernameDBHandler(ctx context.Context, req *authpb.SyncUsernameDBRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).SyncUsernameDB(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetAdminUserListHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).GetAdminUserList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetUserInfoByUsernameHandler(ctx context.Context, req *authpb.WithUsernameRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).GetUserInfoByUsername(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetAdminUserInfoHandler(ctx context.Context, req *authpb.WithUserIdRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).GetAdminUserInfo(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetExcludeLoginUserNameListHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).GetExcludeLoginUserNameList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func ChangeUserStatusHandler(ctx context.Context, req *authpb.ChangeUserStatusRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).ChangeUserStatus(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func IsLoggingOnHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).IsLoggingOn(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GenRandomUsernameHandler(ctx context.Context) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).GenRandomUsername(ctx, &authpb.CommonRequest{})
	if err != nil {
		return res, err
	}
	return res, nil
}

func CheckUserHandler(ctx context.Context, req *authpb.WithUsernameRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).CheckUser(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func RegisterByPassword(ctx context.Context, req *authpb.WithPasswordRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).RegisterByPassword(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func LoginByPassword(ctx context.Context, req *authpb.WithPasswordRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).LoginByPassword(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func UpdatePassword(ctx context.Context, req *authpb.WithPasswordRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).UpdatePassword(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func UpdateUsername(ctx context.Context, req *authpb.WithPasswordRequest) (*authpb.ResponseData, error) {
	err := CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AuthCli.Client).UpdateUsername(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

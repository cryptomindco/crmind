package services

import (
	"context"
	"crmind/pb/assetspb"
	"crmind/utils"
	"fmt"

	"google.golang.org/grpc"
)

var AssetsCli AssetsClient

type AssetsClient struct {
	Client *assetspb.AssetsServiceClient
}

func InitAssetsClient() *assetspb.AssetsServiceClient {
	assetsURL := utils.GetAssetsHost()
	fmt.Println("API Gateway :  InitAssetsClient")
	//	using WithInsecure() because no SSL running
	cc, err := grpc.Dial(assetsURL, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect to assets service:", err)
		return nil
	}
	client := assetspb.NewAssetsServiceClient(cc)
	return &client
}

func CheckAndInitAssetsClient() error {
	if AssetsCli.Client != nil {
		return nil
	}
	AssetsCli.Client = InitAssetsClient()
	if AssetsCli.Client == nil {
		return fmt.Errorf("Init assets client failed")
	}
	return nil
}

func CreateNewAddressHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CreateNewAddress(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func SyncTransactionsHandler(ctx context.Context, req *assetspb.CommonRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).SyncTransactions(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func SendTradingRequestHandler(ctx context.Context, req *assetspb.SendTradingDataRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).SendTradingRequest(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func TransferAmountHandler(ctx context.Context, req *assetspb.TransferAmountRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).TransferAmount(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func AddToContactHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).AddToContact(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FilterTxHistoryHandler(ctx context.Context, req *assetspb.FilterTxHistoryRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).FilterTxHistory(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetCodeListDataHandler(ctx context.Context, req *assetspb.GetCodeListRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetCodeListData(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetAddressListDataWithStatusHandler(ctx context.Context, req *assetspb.GetAddressListRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetAddressListDataWithStatus(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func ConfirmAmountHandler(ctx context.Context, req *assetspb.ConfirmAmountRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).ConfirmAmount(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func ConfirmWithdrawalHandler(ctx context.Context, req *assetspb.ConfirmWithdrawalRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).ConfirmWithdrawal(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func UpdateNewLabelHandler(ctx context.Context, req *assetspb.UpdateLabelRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).UpdateNewLabel(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetBalanceSummaryHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetBalanceSummary(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetAssetDBListHandler(ctx context.Context, req *assetspb.GetAssetDBListRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetAssetDBList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetUserAssetDBHandler(ctx context.Context, req *assetspb.GetUserAssetDBRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetUserAssetDB(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetAddressListHandler(ctx context.Context, req *assetspb.OneIntegerRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetAddressList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CountAddressHandler(ctx context.Context, req *assetspb.CountAddressRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CountAddress(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CheckHasCodeListHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CheckHasCodeList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetContactListHandler(ctx context.Context, req *assetspb.CommonRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetContactList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FilterTxCodeHandler(ctx context.Context, req *assetspb.FilterTxCodeRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).FilterTxCode(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetTxHistoryHandler(ctx context.Context, req *assetspb.OneIntegerRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetTxHistory(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FilterAddressListHandler(ctx context.Context, req *assetspb.FilterAddressListRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).FilterAddressList(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CheckAndCreateAccountTokenHandler(ctx context.Context, req *assetspb.CheckAndCreateAccountTokenRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CheckAndCreateAccountToken(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func FetchRateHandler(ctx context.Context) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).FetchRate(ctx, &assetspb.CommonRequest{})
	if err != nil {
		return res, err
	}
	return res, nil
}

func CheckAssetMatchWithUserHandler(ctx context.Context, req *assetspb.OneIntegerRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CheckAssetMatchWithUser(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CheckAddressMatchWithUserHandler(ctx context.Context, req *assetspb.CheckAddressMatchWithUserRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CheckAddressMatchWithUser(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetAddressHandler(ctx context.Context, req *assetspb.OneIntegerRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetAddress(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func ConfirmAddressActionHandler(ctx context.Context, req *assetspb.ConfirmAddressActionRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).ConfirmAddressAction(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CancelUrlCodeHandler(ctx context.Context, req *assetspb.OneIntegerRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CancelUrlCode(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CheckContactUserHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CheckContactUser(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func WalletSocketHandler(ctx context.Context, req *assetspb.WalletNotifyRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).WalletSocket(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetTxCodeHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).GetTxCode(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func AdminUpdateBalanceHandler(ctx context.Context, req *assetspb.AdminBalanceUpdateRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).AdminUpdateBalance(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func TransactionDetailHandler(ctx context.Context, req *assetspb.OneIntegerRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).TransactionDetail(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func HandlerURLCodeWithdrawlWithAccountHandler(ctx context.Context, req *assetspb.URLCodeWithdrawWithAccountRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).HandlerURLCodeWithdrawlWithAccount(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CreateNewAssetHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).CreateNewAsset(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

func UpdateExchangeRateServerHandler(ctx context.Context, req *assetspb.OneStringRequest) (*assetspb.ResponseData, error) {
	err := CheckAndInitAssetsClient()
	if err != nil {
		return nil, err
	}
	res, err := (*AssetsCli.Client).UpdateExchangeRateServer(ctx, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

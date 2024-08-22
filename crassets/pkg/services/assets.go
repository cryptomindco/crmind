package services

import (
	"context"
	"crassets/pkg/logpack"
	"crassets/pkg/models"
	"crassets/pkg/pb"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"fmt"
)

func (s *Server) GetBalanceSummary(ctx context.Context, reqData *pb.OneStringRequest) (*pb.ResponseData, error) {
	allowassets := reqData.Data
	if !utils.IsEmpty(allowassets) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	//only superadmin has permission access this feature
	if !utils.IsSuperAdmin(int(reqData.Common.Role)) {
		return ResponseError("No access to this feature", utils.GetFuncName(), fmt.Errorf("No access to this featurer"))
	}
	allowList := utils.GetAssetsNameFromStr(allowassets)
	assetList := make([]*models.AssetDisplay, 0)
	for _, asset := range allowList {
		assetDisp := &models.AssetDisplay{
			Type:          asset,
			TypeDisplay:   assets.StringToAssetType(asset).ToFullName(),
			Balance:       s.H.GetTotalUserBalance(asset),
			DaemonBalance: s.GetTotalDaemonBalance(asset),
			SpendableFund: s.GetSpendableAmount(asset),
		}
		assetList = append(assetList, assetDisp)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Balance summary successfully", utils.GetFuncName(), assetList)
}

func (s *Server) GetSpendableAmount(asset string) float64 {
	if asset == assets.USDWalletAsset.String() {
		return 0
	}
	s.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return 0
	}
	daemonBalance := assetObj.GetSpendableAmount()
	return daemonBalance
}

func (s *Server) GetTotalDaemonBalance(asset string) float64 {
	if asset == assets.USDWalletAsset.String() {
		return 0
	}
	s.UpdateAssetManagerByType(asset)
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[asset]
	if !assetMgrExist {
		return 0
	}
	daemonBalance, err := assetObj.GetSystemBalance()
	if err != nil {
		logpack.Error(fmt.Sprintf("Get daemon balance of %s failed", asset), utils.GetFuncName(), err)
		return 0
	}
	return daemonBalance
}

func (s *Server) GetAddress(ctx context.Context, reqData *pb.OneIntegerRequest) (*pb.ResponseData, error) {
	addressid := reqData.Data
	if addressid < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	address, err := s.H.GetAddressById(addressid)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}

	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get address successfully", utils.GetFuncName(), address)
}

func (s *Server) GetUserAssetDB(ctx context.Context, reqData *pb.GetUserAssetDBRequest) (*pb.ResponseData, error) {
	username := reqData.Username
	assetType := reqData.Type
	if utils.IsEmpty(username) || utils.IsEmpty(assetType) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	asset, err := s.H.GetUserAsset(username, assetType)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	type TempoRes struct {
		Exist bool          `json:"exist"`
		Asset *models.Asset `json:"asset"`
	}

	res := &TempoRes{
		Exist: asset != nil,
		Asset: asset,
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get User Asset successfully", utils.GetFuncName(), res)
}

func (s *Server) GetAddressList(ctx context.Context, reqData *pb.OneIntegerRequest) (*pb.ResponseData, error) {
	assetId := reqData.Data
	if assetId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	addressList, err := s.H.GetAddressListByAssetId(assetId)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Address List successfully", utils.GetFuncName(), addressList)
}

func (s *Server) CheckHasCodeList(ctx context.Context, reqData *pb.OneStringRequest) (*pb.ResponseData, error) {
	assetType := reqData.Data
	if utils.IsEmpty(assetType) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	hasCode := s.H.CheckHasCodeList(assetType, reqData.Common.LoginName)
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Check code list successfully", utils.GetFuncName(), hasCode)
}

func (s *Server) GetContactList(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	contacts := s.H.GetContactListOfUser(reqData.LoginName)
	return ResponseSuccessfullyWithAnyData(reqData.LoginName, "Check code list successfully", utils.GetFuncName(), contacts)
}

func (s *Server) CountAddress(ctx context.Context, reqData *pb.CountAddressRequest) (*pb.ResponseData, error) {
	assetId := reqData.AssetId
	activeFlg := reqData.ActiveFlg
	if assetId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	countAddress := s.H.CountAddressesWithStatus(assetId, activeFlg)
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Count address successfully", utils.GetFuncName(), countAddress)
}

func (s *Server) FilterTxCode(ctx context.Context, reqData *pb.FilterTxCodeRequest) (*pb.ResponseData, error) {
	assetType := reqData.AssetType
	status := reqData.Status
	if utils.IsEmpty(assetType) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	txCodeList, err := s.H.FilterUrlCodeList(assetType, status, reqData.Common.LoginName)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Count address successfully", utils.GetFuncName(), txCodeList)
}

func (s *Server) CheckAddressMatchWithUser(ctx context.Context, reqData *pb.CheckAddressMatchWithUserRequest) (*pb.ResponseData, error) {
	assetId := reqData.AssetId
	addressId := reqData.AddressId
	archived := reqData.Archived
	if assetId < 1 || addressId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	isMatch := s.H.CheckMatchAddressWithUser(assetId, addressId, reqData.Common.LoginName, archived)
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Check address match with user successfully", utils.GetFuncName(), isMatch)
}

func (s *Server) CheckAssetMatchWithUser(ctx context.Context, reqData *pb.OneIntegerRequest) (*pb.ResponseData, error) {
	assetId := reqData.Data
	if assetId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	isMatch := s.H.CheckAssetMatchWithUser(assetId, reqData.Common.LoginName)

	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Check asset match with user successfully", utils.GetFuncName(), isMatch)
}

func (s *Server) CheckAndCreateAccountToken(ctx context.Context, reqData *pb.CheckAndCreateAccountTokenRequest) (*pb.ResponseData, error) {
	username := reqData.Username
	role := reqData.Role
	if utils.IsEmpty(username) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	token, _, err := s.H.CheckAndCreateAccountToken(username, int(role))
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Check or create account token successfully", utils.GetFuncName(), token)
}

func (s *Server) FilterAddressList(ctx context.Context, reqData *pb.FilterAddressListRequest) (*pb.ResponseData, error) {
	assetId := reqData.AssetId
	status := reqData.Status
	if assetId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	addressList, err := s.H.FilterAddressList(assetId, status)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Filter Address List successfully", utils.GetFuncName(), addressList)
}

func (s *Server) GetTxHistory(ctx context.Context, reqData *pb.OneIntegerRequest) (*pb.ResponseData, error) {
	txHistoryId := reqData.Data
	if txHistoryId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}

	txHistory, err := s.H.GetTxHistoryById(txHistoryId)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get transaction history successfully", utils.GetFuncName(), txHistory)
}

func (s *Server) GetAssetDBList(ctx context.Context, reqData *pb.OneStringRequest) (*pb.ResponseData, error) {
	allowAssets := reqData.Data
	if utils.IsEmpty(allowAssets) {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}

	allowList := utils.GetAssetsNameFromStr(allowAssets)
	assetList, err := s.H.GetAssetList(reqData.Common.LoginName, allowList)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, err.Error(), utils.GetFuncName(), err)
	}
	assetList = s.H.SyncAssetList(reqData.Common.LoginName, assetList, allowList)
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Get Asset list successfully", utils.GetFuncName(), assetList)
}

func (s *Server) FetchRate(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	rateMap, err := s.H.ReadRateFromDB()
	if err != nil {
		return ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyDataNoLog(utils.ObjectToJsonString(rateMap))
}

func (s *Server) ConfirmAddressAction(ctx context.Context, reqData *pb.ConfirmAddressActionRequest) (*pb.ResponseData, error) {
	assetId := reqData.AssetId
	addressId := reqData.AddressId
	action := reqData.Action
	if assetId < 1 || addressId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	//check valid assetId, addressId
	if !s.H.CheckMatchAddressWithUser(assetId, addressId, reqData.Common.LoginName, action == "reuse") {
		return ResponseLoginError(reqData.Common.LoginName, "The user login information and assets do not match", utils.GetFuncName(), nil)
	}

	//Get address object
	address, addressErr := s.H.GetAddressById(addressId)
	if addressErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get address from DB failed", utils.GetFuncName(), addressErr)
	}
	address.Archived = action != "reuse"

	tx := s.H.DB.Begin()
	updateErr := tx.Save(address).Error
	if updateErr != nil {
		tx.Rollback()
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Update Address failed", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	//return successfully
	return ResponseSuccessfully(reqData.Common.LoginName, "Update address from DB successfully", utils.GetFuncName())
}

func (s *Server) CancelUrlCode(ctx context.Context, reqData *pb.OneIntegerRequest) (*pb.ResponseData, error) {
	codeId := reqData.Data
	if codeId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	cancelErr := s.H.CancelTxCodeByUsername(reqData.Common.LoginName, codeId)
	if cancelErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, cancelErr.Error(), utils.GetFuncName(), nil)
	}
	return ResponseSuccessfully(reqData.Common.LoginName, "Cancel Withdraw Code successfully!", utils.GetFuncName())
}

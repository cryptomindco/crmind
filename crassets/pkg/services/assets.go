package services

import (
	"crassets/pkg/logpack"
	"crassets/pkg/models"
	"crassets/pkg/pb"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"fmt"
)

func (s *Server) GetBalanceSummary(reqData *pb.RequestData) *pb.ResponseData {
	allowassetsAny, allowassetsExist := reqData.DataMap["allowassets"]
	if !allowassetsExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	allowassets := allowassetsAny.(string)
	//only superadmin has permission access this feature
	if !utils.IsSuperAdmin(reqData.Role) {
		return pb.ResponseError("No access to this feature", utils.GetFuncName(), fmt.Errorf("No access to this featurer"))
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
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get Balance summary successfully", utils.GetFuncName(), assetList)
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

func (s *Server) GetAddress(reqData *pb.RequestData) *pb.ResponseData {
	addressidAny, addressidExist := reqData.DataMap["addressid"]
	if !addressidExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	addressid := addressidAny.(int64)
	address, err := s.H.GetAddressById(addressid)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}

	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get address successfully", utils.GetFuncName(), address)
}

func (s *Server) GetUserAssetDB(reqData *pb.RequestData) *pb.ResponseData {
	usernameAny, usernameExist := reqData.DataMap["username"]
	typeAny, typeExist := reqData.DataMap["type"]
	if !usernameExist || !typeExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	username := usernameAny.(string)
	assetType := typeAny.(string)
	asset, err := s.H.GetUserAsset(username, assetType)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	type TempoRes struct {
		Exist bool          `json:"exist"`
		Asset *models.Asset `json:"asset"`
	}

	res := &TempoRes{
		Exist: asset != nil,
		Asset: asset,
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get User Asset successfully", utils.GetFuncName(), res)
}

func (s *Server) GetAddressList(reqData *pb.RequestData) *pb.ResponseData {
	assetidAny, assetidExist := reqData.DataMap["assetid"]
	if !assetidExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetidAny.(int64)
	addressList, err := s.H.GetAddressListByAssetId(assetId)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get Address List successfully", utils.GetFuncName(), addressList)
}

func (s *Server) CheckHasCodeList(reqData *pb.RequestData) *pb.ResponseData {
	typeAny, typeExist := reqData.DataMap["assetType"]
	if !typeExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetType := typeAny.(string)

	hasCode := s.H.CheckHasCodeList(assetType, reqData.LoginName)
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Check code list successfully", utils.GetFuncName(), hasCode)
}

func (s *Server) GetContactList(reqData *pb.RequestData) *pb.ResponseData {
	contacts := s.H.GetContactListOfUser(reqData.LoginName)
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Check code list successfully", utils.GetFuncName(), contacts)
}

func (s *Server) CountAddress(reqData *pb.RequestData) *pb.ResponseData {
	assetidAny, assetidExist := reqData.DataMap["assetid"]
	activeflgAny, activeflgExist := reqData.DataMap["activeflg"]
	if !assetidExist || !activeflgExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetidAny.(int64)
	activeFlg := activeflgAny.(bool)
	countAddress := s.H.CountAddressesWithStatus(assetId, activeFlg)
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Count address successfully", utils.GetFuncName(), countAddress)
}

func (s *Server) FilterTxCode(reqData *pb.RequestData) *pb.ResponseData {
	assettypeAny, assettypeExist := reqData.DataMap["assettype"]
	statusAny, statusExist := reqData.DataMap["status"]
	if !assettypeExist || !statusExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetType := assettypeAny.(string)
	status := statusAny.(string)
	txCodeList, err := s.H.FilterUrlCodeList(assetType, status, reqData.LoginName)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Count address successfully", utils.GetFuncName(), txCodeList)
}

func (s *Server) CheckAddressMatchWithUser(reqData *pb.RequestData) *pb.ResponseData {
	assetidAny, assetidExist := reqData.DataMap["assetid"]
	addressidAny, addressidExist := reqData.DataMap["addressid"]
	archivedAny, archivedExist := reqData.DataMap["archived"]
	if !assetidExist || !addressidExist || !archivedExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetidAny.(int64)
	addressId := addressidAny.(int64)
	archived := archivedAny.(bool)
	isMatch := s.H.CheckMatchAddressWithUser(assetId, addressId, reqData.LoginName, archived)
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Check address match with user successfully", utils.GetFuncName(), isMatch)
}

func (s *Server) CheckAssetMatchWithUser(reqData *pb.RequestData) *pb.ResponseData {
	assetidAny, assetidExist := reqData.DataMap["assetid"]
	if !assetidExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetidAny.(int64)
	isMatch := s.H.CheckAssetMatchWithUser(assetId, reqData.LoginName)

	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Check asset match with user successfully", utils.GetFuncName(), isMatch)
}

func (s *Server) CheckAndCreateAccountToken(reqData *pb.RequestData) *pb.ResponseData {
	usernameAny, usernameExist := reqData.DataMap["username"]
	roleAny, roleExist := reqData.DataMap["role"]
	if !usernameExist || !roleExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	username := usernameAny.(string)
	role := roleAny.(int)
	token, _, err := s.H.CheckAndCreateAccountToken(username, int(role))
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Check or create account token successfully", utils.GetFuncName(), token)
}

func (s *Server) FilterAddressList(reqData *pb.RequestData) *pb.ResponseData {
	assetidAny, assetidExist := reqData.DataMap["assetid"]
	statusAny, statusExist := reqData.DataMap["status"]
	if !assetidExist || !statusExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetidAny.(int64)
	status := statusAny.(string)
	addressList, err := s.H.FilterAddressList(assetId, status)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Filter Address List successfully", utils.GetFuncName(), addressList)
}

func (s *Server) GetTxHistory(reqData *pb.RequestData) *pb.ResponseData {
	txhistoryidAny, txhistoryidExist := reqData.DataMap["txhistoryid"]
	if !txhistoryidExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	txHistoryId := txhistoryidAny.(int64)
	txHistory, err := s.H.GetTxHistoryById(txHistoryId)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get transaction history successfully", utils.GetFuncName(), txHistory)
}

func (s *Server) GetAssetDBList(reqData *pb.RequestData) *pb.ResponseData {
	allowassetsAny, allowassetsExist := reqData.DataMap["allowassets"]
	if !allowassetsExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	allowAssets := allowassetsAny.(string)
	allowList := utils.GetAssetsNameFromStr(allowAssets)
	assetList, err := s.H.GetAssetList(reqData.LoginName, allowList)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	assetList = s.H.SyncAssetList(reqData.LoginName, assetList, allowList)
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get Asset list successfully", utils.GetFuncName(), assetList)
}

func (s *Server) FetchRate(reqData *pb.RequestData) *pb.ResponseData {
	rateMap, err := s.H.ReadRateFromDB()
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyDataNoLog(utils.ObjectToJsonString(rateMap))
}

func (s *Server) ConfirmAddressAction(reqData *pb.RequestData) *pb.ResponseData {
	assetIdAny, assetIdExist := reqData.DataMap["assetId"]
	addressIdAny, addressIdExist := reqData.DataMap["addressId"]
	actionAny, actionExist := reqData.DataMap["action"]
	if !assetIdExist || !addressIdExist || !actionExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	assetId := assetIdAny.(int64)
	addressId := addressIdAny.(int64)
	action := actionAny.(string)
	//check valid assetId, addressId
	if !s.H.CheckMatchAddressWithUser(assetId, addressId, reqData.LoginName, action == "reuse") {
		return pb.ResponseLoginError(reqData.LoginId, "The user login information and assets do not match", utils.GetFuncName(), nil)
	}

	//Get address object
	address, addressErr := s.H.GetAddressById(addressId)
	if addressErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, "Get address from DB failed", utils.GetFuncName(), addressErr)
	}
	address.Archived = action != "reuse"

	tx := s.H.DB.Begin()
	updateErr := tx.Save(address).Error
	if updateErr != nil {
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Update Address failed", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	//return successfully
	return pb.ResponseSuccessfully(reqData.LoginId, "Update address from DB successfully", utils.GetFuncName())
}

func (s *Server) CancelUrlCode(reqData *pb.RequestData) *pb.ResponseData {
	codeIdAny, codeIdExist := reqData.DataMap["codeId"]
	if !codeIdExist {
		return pb.ResponseLoginError(reqData.LoginId, "Param failed. Please try again!", utils.GetFuncName(), nil)
	}
	codeId := codeIdAny.(int64)
	cancelErr := s.H.CancelTxCodeById(reqData.LoginName, codeId)
	if cancelErr != nil {
		return pb.ResponseLoginError(reqData.LoginId, cancelErr.Error(), utils.GetFuncName(), nil)
	}
	return pb.ResponseSuccessfully(reqData.LoginId, "Cancel Withdraw Code successfully!", utils.GetFuncName())
}

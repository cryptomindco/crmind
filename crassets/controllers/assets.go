package controllers

import (
	"crassets/handler"
	"crassets/models"
	"crassets/utils"
	"crassets/walletlib/assets"
	"fmt"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

type AssetsController struct {
	BaseController
}

func (this *AssetsController) GetBalanceSummary() {
	//check login
	loginUser, err := this.AuthCheck()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	//only superadmin has permission access this feature
	if !utils.IsSuperAdmin(loginUser.Role) {
		this.ResponseError("No access to this feature", utils.GetFuncName(), fmt.Errorf("No access to this featurer"))
		return
	}
	allowAssets := this.Ctx.Request.Header.Get("AllowAssets")
	allowList := utils.GetAssetsNameFromStr(allowAssets)
	assetList := make([]*models.AssetDisplay, 0)
	for _, asset := range allowList {
		assetDisp := &models.AssetDisplay{
			Type:          asset,
			TypeDisplay:   assets.StringToAssetType(asset).ToFullName(),
			Balance:       utils.GetTotalUserBalance(asset),
			DaemonBalance: handler.GetTotalDaemonBalance(asset),
			SpendableFund: handler.GetSpendableAmount(asset),
		}
		assetList = append(assetList, assetDisp)
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get Balance summary successfully", utils.GetFuncName(), assetList)
}

func (this *AssetsController) GetAddress() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	addressIdStr := this.Ctx.Input.Query("addressid")
	addressId, err := strconv.ParseInt(addressIdStr, 0, 32)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	address, err := utils.GetAddressById(addressId)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get address successfully", utils.GetFuncName(), address)
}

func (this *AssetsController) GetUserAssetDB() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	userIdStr := this.Ctx.Input.Query("userid")
	assetType := this.Ctx.Input.Query("type")
	userId, err := strconv.ParseInt(userIdStr, 0, 32)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	asset, err := utils.GetUserAsset(userId, assetType)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	type TempoRes struct {
		Exist bool          `json:"exist"`
		Asset *models.Asset `json:"asset"`
	}

	res := &TempoRes{
		Exist: asset != nil,
		Asset: asset,
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get User Asset successfully", utils.GetFuncName(), res)
}

func (this *AssetsController) GetAddressList() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetIdStr := this.Ctx.Input.Query("assetid")
	assetId, err := strconv.ParseInt(assetIdStr, 0, 32)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	addressList, err := this.GetAddressListByAssetId(assetId)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get Address List successfully", utils.GetFuncName(), addressList)
}

func (this *AssetsController) CheckHasCodeList() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetType := this.Ctx.Input.Query("assetType")
	hasCode := utils.CheckHasCodeList(assetType, loginUser.Id)
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check code list successfully", utils.GetFuncName(), hasCode)
}

func (this *AssetsController) GetContactList() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	contacts := utils.GetContactListOfUser(loginUser.Id)
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check code list successfully", utils.GetFuncName(), contacts)
}

func (this *AssetsController) CountAddress() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetIdStr := this.Ctx.Input.Query("assetid")
	activeFlagStr := this.Ctx.Input.Query("activeflg")
	assetId, err := strconv.ParseInt(assetIdStr, 0, 32)
	activeFlg, activeErr := strconv.ParseBool(activeFlagStr)
	if err != nil || activeErr != nil {
		this.ResponseError("Param failed", utils.GetFuncName(), fmt.Errorf("Param failed"))
		return
	}
	countAddress := utils.CountAddressesWithStatus(assetId, activeFlg)
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Count address successfully", utils.GetFuncName(), countAddress)
}

func (this *AssetsController) FilterTxCode() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetType := this.Ctx.Input.Query("assettype")
	status := this.Ctx.Input.Query("status")

	txCodeList, err := utils.FilterUrlCodeList(assetType, status, loginUser.Id)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Count address successfully", utils.GetFuncName(), txCodeList)
}

func (this *AssetsController) CheckAddressMatchWithUser() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetIdStr := this.Ctx.Input.Query("assetid")
	addressIdStr := this.Ctx.Input.Query("addressid")
	archivedStr := this.Ctx.Input.Query("archived")
	assetId, err := strconv.ParseInt(assetIdStr, 0, 32)
	addressId, addressErr := strconv.ParseInt(addressIdStr, 0, 32)
	archived, archivedErr := strconv.ParseBool(archivedStr)
	if err != nil || addressErr != nil || archivedErr != nil {
		this.ResponseError("Param failed", utils.GetFuncName(), fmt.Errorf("Param failed"))
		return
	}
	isMatch := utils.CheckMatchAddressWithUser(assetId, addressId, loginUser.Id, archived)

	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check address match with user successfully", utils.GetFuncName(), isMatch)
}

func (this *AssetsController) CheckAssetMatchWithUser() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetIdStr := this.Ctx.Input.Query("assetid")
	assetId, err := strconv.ParseInt(assetIdStr, 0, 32)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	isMatch := utils.CheckAssetMatchWithUser(assetId, loginUser.Id)

	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check asset match with user successfully", utils.GetFuncName(), isMatch)
}

func (this *AssetsController) CheckAndCreateAccountToken() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	userIdStr := this.Ctx.Input.Query("userid")
	userName := this.Ctx.Input.Query("username")
	roleStr := this.Ctx.Input.Query("role")
	userId, err := strconv.ParseInt(userIdStr, 0, 32)
	role, roleErr := strconv.ParseInt(roleStr, 0, 32)
	if err != nil || roleErr != nil {
		this.ResponseError("Param error", utils.GetFuncName(), fmt.Errorf("Param error"))
		return
	}
	token, _, err := utils.CheckAndCreateAccountToken(userId, userName, int(role))
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Check or create account token successfully", utils.GetFuncName(), token)
}

func (this *AssetsController) FilterAddressList() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetIdStr := this.Ctx.Input.Query("assetid")
	assetId, err := strconv.ParseInt(assetIdStr, 0, 32)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	status := this.Ctx.Input.Query("status")
	addressList, err := utils.FilterAddressList(assetId, status)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Filter Address List successfully", utils.GetFuncName(), addressList)
}

func (this *AssetsController) GetTxHistory() {
	authToken := this.Ctx.Input.Query("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	txHistoryIdStr := this.Ctx.Input.Query("txhistoryid")
	txHistoryId, err := strconv.ParseInt(txHistoryIdStr, 0, 32)

	txHistory, err := utils.GetTxHistoryById(txHistoryId)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get transaction history successfully", utils.GetFuncName(), txHistory)
}

func (this *AssetsController) GetAssetDBList() {
	//check login
	loginUser, err := this.AuthCheck()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	allowAssets := this.Ctx.Request.Header.Get("AllowAssets")
	allowList := utils.GetAssetsNameFromStr(allowAssets)
	assetList, err := this.GetAssetList(loginUser.Id, allowList)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetList = this.SyncAssetList(loginUser, assetList, allowList)
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get Asset list successfully", utils.GetFuncName(), assetList)
}

func (this *AssetsController) FetchRate() {
	rateMap, err := utils.ReadRateFromDB()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyDataNoLog(utils.ObjectToJsonString(rateMap))
}

func (this *AssetsController) ConfirmAddressAction() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	assetId, parseAssetErr := this.GetInt64("assetId")
	addressId, parseAddrErr := this.GetInt64("addressId")
	if parseAddrErr != nil || parseAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Check param error. Please try again!", utils.GetFuncName(), nil)
		return
	}
	action := this.GetString("action")
	//check valid assetId, addressId
	if !utils.CheckMatchAddressWithUser(assetId, addressId, loginUser.Id, action == "reuse") {
		this.ResponseLoginError(loginUser.Id, "The user login information and assets do not match", utils.GetFuncName(), nil)
		return
	}

	//Get address object
	address, addressErr := utils.GetAddressById(addressId)
	if addressErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get address from DB failed", utils.GetFuncName(), addressErr)
		return
	}
	address.Archived = action != "reuse"

	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	_, updateErr := tx.Update(address)
	if updateErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Update Address failed", utils.GetFuncName(), updateErr)
		return
	}
	tx.Commit()
	//return successfully
	this.ResponseSuccessfully(loginUser.Id, "Update address from DB successfully", utils.GetFuncName())
}

func (this *AssetsController) CancelUrlCode() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	//get codeid
	codeIdStr := strings.TrimSpace(this.GetString("codeId"))
	codeId, err := strconv.ParseInt(codeIdStr, 0, 32)
	if err != nil {
		this.ResponseLoginError(loginUser.Id, "Parse codeId to cancel failed. Please try again!", utils.GetFuncName(), nil)
		return
	}
	cancelErr := utils.CancelTxCodeById(loginUser.Id, codeId)
	if cancelErr != nil {
		this.ResponseLoginError(loginUser.Id, cancelErr.Error(), utils.GetFuncName(), nil)
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Cancel Withdraw Code successfully!", utils.GetFuncName())
}

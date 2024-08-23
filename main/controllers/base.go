package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/pb/assetspb"
	"crmind/pb/authpb"
	"crmind/pb/chatpb"
	"crmind/services"
	"crmind/utils"
	"encoding/json"
	"fmt"
	"net/http"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
)

var (
	leftTreeResultMap = make(map[int][]orm.Params)
)

type BaseController struct {
	beego.Controller
}

func (this *BaseController) ResponseRollbackError(tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseError(msg, funcName, err)
}

func (this *BaseController) ResponseLoginRollbackError(loginId int64, tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseLoginError(loginId, msg, funcName, err)
}

func (this *BaseController) ResponseError(msg string, funcName string, err error) {
	logpack.Error(msg, funcName, err)
	this.Data["json"] = utils.ResponseData{
		IsError: true,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseLoginError(loginId int64, msg string, funcName string, err error) {
	if loginId <= 0 {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginId, funcName, err)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: true,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseSuccessfully(loginId int64, msg string, funcName string) {
	if loginId <= 0 {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginId, funcName)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseSuccessfullyWithAnyData(loginId int64, msg, funcName string, result any) {
	if loginId <= 0 {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginId, funcName)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Msg:     msg,
		Data:    result,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseSuccessfullyWithAnyDataNoLog(result any) {
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Data:    result,
	}
	this.ServeJSON()
}

func (this *BaseController) AdminAuthCheck() (*models.AuthClaims, error) {
	authClaim, err := this.AuthCheck()
	if err != nil {
		return nil, err
	}
	if authClaim.Role != int(utils.RoleSuperAdmin) {
		return nil, fmt.Errorf("Login user is not superadmin")
	}
	return authClaim, nil
}

func (this *BaseController) AuthCheck() (*models.AuthClaims, error) {
	authClaim, err := this.GetLoginUser()
	if err != nil {
		this.Redirect("/login", http.StatusFound)
		return nil, err
	}
	this.Data["LoginUser"] = authClaim
	this.Data["IsSuperAdmin"] = this.IsSuperAdmin(*authClaim)

	//get user list json from session
	userListJson := this.GetSession(utils.UserListSessionKey)
	userInfoList := make([]models.UserInfo, 0)
	//if userList is empty, get userList
	if utils.IsEmpty(userListJson) {
		userInfoList = this.GetUsernameListExcludeId()
	} else {
		usernamesJsonBytes, err := json.Marshal(userListJson)
		if err == nil {
			json.Unmarshal(usernamesJsonBytes, &userInfoList)
		}
	}
	this.Data["UserInfoList"] = userInfoList
	isChatActive := utils.IsChatActive()
	if isChatActive {
		//user chat list initialization
		chatMsgList, hasUnreadChatCount := this.GetChatMsgDisplayList(authClaim.Username)
		//content chatMsgList data to json
		chatMsgJsonStr, chatJsonErr := utils.ConvertToJsonString(chatMsgList)
		if chatJsonErr != nil {
			chatMsgJsonStr = "[]"
		}
		this.Data["ChatMsgList"] = chatMsgJsonStr
		this.Data["ItemUnreadChatCount"] = hasUnreadChatCount

	}
	this.Data["LoginToken"] = this.GetLoginToken()
	this.Data["AuthActive"] = utils.IsAuthActive()
	this.Data["ChatActive"] = utils.IsChatActive()
	this.Data["AssetsActive"] = utils.IsAssetsActive()
	return authClaim, nil
}

func (this *BaseController) GetChatMsgDisplayList(loginName string) ([]*models.ChatDisplay, int) {
	res, err := services.GetChatMsgDisplayListHandler(this.Ctx.Request.Context(), &chatpb.CommonRequest{
		LoginName: loginName,
	})
	if err != nil {
		return nil, 0
	}
	type ResData struct {
		ChatList    []*models.ChatDisplay `json:"chatList"`
		UnreadCount int                   `json:"unreadCount"`
	}
	var result ResData
	parseErr := utils.JsonStringToObject(res.Data, &result)
	if parseErr != nil {
		return nil, 0
	}

	return result.ChatList, result.UnreadCount

}

func (this *BaseController) GetAdminAssetsBalance(loginName string, role int64) ([]*models.AssetDisplay, error) {
	res, err := services.GetBalanceSummaryHandler(this.Ctx.Request.Context(), &assetspb.OneStringRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
			Role:      role,
		},
		Data: utils.GetAllowAssets(),
	})
	var result []*models.AssetDisplay
	if err != nil {
		return nil, err
	}

	err = utils.JsonStringToObject(res.Data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (this *BaseController) GetAddressListByAssetId(loginName string, assetId int64) ([]string, error) {
	res, err := services.GetAddressListHandler(this.Ctx.Request.Context(), &assetspb.OneIntegerRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Data: assetId,
	})
	if err != nil {
		return nil, err
	}
	var addressList []string
	err = utils.JsonStringToObject(res.Data, &addressList)
	if err != nil {
		return nil, err
	}
	return addressList, nil
}

func (this *BaseController) GetAddressById(loginName string, addressId int64) (*models.Addresses, error) {
	res, err := services.GetAddressHandler(this.Ctx.Request.Context(), &assetspb.OneIntegerRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Data: addressId,
	})
	if err != nil {
		return nil, err
	}

	var addr models.Addresses
	err = utils.JsonStringToObject(res.Data, &addr)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (this *BaseController) GetAssetByUser(loginName, userName string, assetType string) (*utils.TempoRes, error) {
	res, err := services.GetUserAssetDBHandler(this.Ctx.Request.Context(), &assetspb.GetUserAssetDBRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Username: userName,
		Type:     assetType,
	})
	if err != nil {
		return nil, err
	}
	var result utils.TempoRes
	err = utils.JsonStringToObject(res.Data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *BaseController) CheckMatchAddressWithUser(loginName string, assetId, addressId int64, archived bool) bool {
	res, err := services.CheckAddressMatchWithUserHandler(this.Ctx.Request.Context(), &assetspb.CheckAddressMatchWithUserRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		AssetId:   assetId,
		AddressId: addressId,
		Archived:  archived,
	})
	if err != nil {
		return false
	}
	var resMap map[string]bool
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		return false
	}

	isMatch, isOk := resMap["isMatch"]
	if !isOk {
		return false
	}
	return isMatch
}

func (this *BaseController) CheckAssetMatchUser(loginName string, assetId int64) bool {
	res, err := services.CheckAssetMatchWithUserHandler(this.Ctx.Request.Context(), &assetspb.OneIntegerRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Data: assetId,
	})
	if err != nil {
		return false
	}
	var resMap map[string]bool
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		return false
	}
	isMatch := resMap["isMatch"]
	return isMatch
}

func (this *BaseController) FilterAddressList(loginName string, assetId int64, status string) ([]models.Addresses, error) {
	res, err := services.FilterAddressListHandler(this.Ctx.Request.Context(), &assetspb.FilterAddressListRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		AssetId: assetId,
		Status:  status,
	})

	if err != nil {
		return nil, err
	}
	var addrListRes []models.Addresses
	err = utils.JsonStringToObject(res.Data, &addrListRes)
	if err != nil {
		return nil, err
	}
	return addrListRes, nil
}

func (this *BaseController) CheckAndCreateAccountToken(loginName, username string, role int) (string, error) {
	res, err := services.CheckAndCreateAccountTokenHandler(this.Ctx.Request.Context(), &assetspb.CheckAndCreateAccountTokenRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Username: username,
		Role:     int64(role),
	})

	if err != nil {
		return "", err
	}

	var resMap map[string]string
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		return "", parseErr
	}

	token, ok := resMap["token"]
	if !ok {
		return "", fmt.Errorf("Get token failed")
	}
	return token, nil
}

func (this *BaseController) GetTxHistory(loginName string, txHistoryId int64) (*models.TxHistory, error) {
	res, err := services.GetTxHistoryHandler(this.Ctx.Request.Context(), &assetspb.OneIntegerRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Data: txHistoryId,
	})

	if err != nil {
		return nil, err
	}
	var txRes models.TxHistory
	err = utils.JsonStringToObject(res.Data, &txRes)
	if err != nil {
		return nil, err
	}
	return &txRes, nil
}

func (this *BaseController) FilterUrlCodeList(loginName, assetType string, status string) ([]models.TxCode, error) {
	res, err := services.FilterTxCodeHandler(this.Ctx.Request.Context(), &assetspb.FilterTxCodeRequest{
		Common:    &assetspb.CommonRequest{LoginName: loginName},
		AssetType: assetType,
		Status:    status,
	})

	if err != nil {
		return nil, err
	}
	var resultData []models.TxCode
	parseErr := utils.JsonStringToObject(res.Data, &resultData)
	if parseErr != nil {
		return nil, parseErr
	}
	return resultData, nil
}

func (this *BaseController) CountAddressesWithStatus(loginName string, assetId int64, activeFlg bool) int64 {
	res, err := services.CountAddressHandler(this.Ctx.Request.Context(), &assetspb.CountAddressRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		AssetId:   assetId,
		ActiveFlg: activeFlg,
	})

	if err != nil {
		return 0
	}
	var resMap map[string]int64
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		return 0
	}
	count, ok := resMap["count"]
	if ok {
		return 0
	}
	return count
}

func (this *BaseController) CheckHasCodeList(loginName string, assetType string) bool {
	res, err := services.CheckHasCodeListHandler(this.Ctx.Request.Context(), &assetspb.OneStringRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Data: assetType,
	})

	if err != nil {
		return false
	}

	var resMap map[string]bool
	parseErr := utils.JsonStringToObject(res.Data, &resMap)
	if parseErr != nil {
		return false
	}
	check, ok := resMap["hasCode"]
	if !ok {
		return false
	}
	return check
}

func (this *BaseController) GetContactList(loginName string) []string {
	res, err := services.GetContactListHandler(this.Ctx.Request.Context(), &assetspb.CommonRequest{
		LoginName: loginName,
	})
	if err != nil {
		return []string{}
	}
	var result []string
	parseErr := utils.JsonStringToObject(res.Data, &result)
	if parseErr != nil {
		return []string{}
	}
	return result
}

func (this *BaseController) GetUserByUsername(username string) (*models.UserInfo, error) {
	res, err := services.GetUserInfoByUsernameHandler(this.Ctx.Request.Context(), &authpb.WithUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	var result models.UserInfo
	parseErr := utils.JsonStringToObject(res.Data, &result)
	if parseErr != nil {
		return nil, parseErr
	}
	return &result, nil
}

func (this *BaseController) GetUserAssetList(loginName, username string) ([]*models.Asset, error) {
	res, err := services.GetAssetDBListHandler(this.Ctx.Request.Context(), &assetspb.GetAssetDBListRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginName,
		},
		Allowassets: utils.GetAllowAssets(),
		Username:    username,
	})
	var result []*models.Asset
	if err != nil {
		return result, err
	}
	parseErr := utils.JsonStringToObject(res.Data, &result)
	if parseErr != nil {
		return result, parseErr
	}
	return result, nil
}

func (this *BaseController) GetUsernameListExcludeId() []models.UserInfo {
	res, err := services.GetExcludeLoginUserNameListHandler(this.Ctx.Request.Context(), &authpb.CommonRequest{
		AuthToken: this.GetLoginToken(),
	})
	result := make([]models.UserInfo, 0)
	if err != nil {
		return result
	}
	utils.JsonStringToObject(res.Data, &result)
	return result
}

func (this *BaseController) SimpleAdminAuthCheck() (*models.AuthClaims, error) {
	authClaim, err := this.GetLoginUser()
	if err != nil || authClaim.Role != int(utils.RoleSuperAdmin) {
		return nil, fmt.Errorf("Check admin login failed")
	}
	return authClaim, nil
}

func (this *BaseController) GetLoginUser() (*models.AuthClaims, error) {
	authClaimObj := this.GetSession(utils.LoginUserKey)
	var authClaim models.AuthClaims
	if authClaimObj == nil {
		return nil, fmt.Errorf("Login session not exist")
	}
	userJson, err := json.Marshal(authClaimObj)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(userJson, &authClaim)
	if err != nil {
		return nil, err
	}
	return &authClaim, nil
}

// Check user is superadmin
func (this *BaseController) IsSuperAdmin(user models.AuthClaims) bool {
	return user.Role == int(utils.RoleSuperAdmin)
}

func (this *BaseController) SyncUsernameDB(userId int64, oldUsername, newUsername string) {
	this.SyncUsernameOnUserTable(userId, oldUsername, newUsername)
}

func (this *BaseController) SyncUsernameOnUserTable(userId int64, oldUsername, newUsername string) {
	_, err := services.SyncUsernameDBHandler(this.Ctx.Request.Context(), &authpb.SyncUsernameDBRequest{
		Common: &authpb.CommonRequest{
			AuthToken: fmt.Sprintf("%s%s", "Bearer ", this.GetSession(utils.Tokenkey).(string)),
		},
		NewUsername: newUsername,
		OldUsername: oldUsername,
	})

	if err != nil {
		logpack.FError("Sync user data failed", userId, utils.GetFuncName(), err)
		return
	}
	logpack.FInfo("Sync user data successfully", userId, utils.GetFuncName())
}

func (this *BaseController) GetLoginToken() string {
	return fmt.Sprintf("%s%s", "Bearer ", this.GetSession(utils.Tokenkey).(string))
}

func (this *BaseController) CheckSettingsExist() (*models.Settings, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(settingsModel).Limit(1).One(&settings)
	if queryErr != nil {
		if queryErr == orm.ErrNoRows {
			return nil, nil
		}
		return nil, queryErr
	}
	return &settings, nil
}

func (this *BaseController) GetAssetNamesFromAssetList(assetList []*models.Asset) []string {
	result := make([]string, 0)
	for _, asset := range assetList {
		result = append(result, asset.Type)
	}
	return result
}

func (this *BaseController) CheckUserExist(username string) (bool, error) {
	res, err := services.CheckUserHandler(this.Ctx.Request.Context(), &authpb.WithUsernameRequest{
		Username: username,
	})
	if err != nil {
		return false, fmt.Errorf("check user exist failed")
	}
	var resData map[string]bool
	err = utils.JsonStringToObject(res.Data, &resData)
	if err != nil {
		return false, fmt.Errorf("parse res data failed")
	}
	exist := resData["exist"]
	return exist, nil
}

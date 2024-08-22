package controllers

import (
	"crmind/logpack"
	"crmind/models"
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
	//user chat list initialization
	chatMsgList, hasUnreadChatCount := this.GetChatMsgDisplayList(authClaim.Id)
	//content chatMsgList data to json
	chatMsgJsonStr, chatJsonErr := utils.ConvertToJsonString(chatMsgList)
	if chatJsonErr != nil {
		chatMsgJsonStr = "[]"
	}
	this.Data["ChatMsgList"] = chatMsgJsonStr
	this.Data["LoginToken"] = this.GetLoginToken()
	this.Data["ItemUnreadChatCount"] = hasUnreadChatCount
	return authClaim, nil
}

func (this *BaseController) GetChatMsgDisplayList(userId int64) ([]*models.ChatDisplay, int) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.ChatSite(), "/get-chat-msg"),
		Header: map[string]string{
			"Authorization": this.GetLoginToken(),
			"UserId":        fmt.Sprintf("%d", userId)},
	}
	type ResData struct {
		ChatList    []*models.ChatDisplay `json:"chatList"`
		UnreadCount int                   `json:"unreadCount"`
	}
	var result ResData
	err := services.HttpRequest(req, &response)
	if err == nil && !response.IsError {
		jsonBytes, err := json.Marshal(response.Data)
		if err == nil {
			err = json.Unmarshal(jsonBytes, &result)
			if err == nil {
				return result.ChatList, result.UnreadCount
			}
		}
	}
	return nil, 0
}

func (this *BaseController) GetAdminAssetsBalance() ([]*models.AssetDisplay, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-balance-summary"),
		Header: map[string]string{
			"Authorization": this.GetLoginToken(),
			"AllowAssets":   utils.GetAllowAssets()},
	}

	var result []*models.AssetDisplay
	err := services.HttpRequest(req, &response)
	if err != nil {
		return nil, err
	}
	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	err = utils.CatchObject(response.Data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (this *BaseController) GetAddressListByAssetId(assetId int64) ([]string, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-address-list"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assetid":       fmt.Sprintf("%d", assetId),
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return nil, err
	}
	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	var addressList []string
	err := utils.CatchObject(response.Data, &addressList)
	if err != nil {
		return nil, err
	}
	return addressList, nil
}

func (this *BaseController) GetAddressById(addressId int64) (*models.Addresses, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-address"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"addressid":     fmt.Sprintf("%d", addressId),
		},
	}
	err := services.HttpRequest(req, &response)
	if err != nil {
		return nil, err
	}
	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	var res models.Addresses
	err = utils.CatchObject(response.Data, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (this *BaseController) GetAssetByUser(userId int64, assetType string) (*utils.TempoRes, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-user-asset"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"userid":        fmt.Sprintf("%d", userId),
			"type":          assetType,
		},
	}
	var result utils.TempoRes

	err := services.HttpRequest(req, &response)
	if err != nil {
		return nil, err
	}
	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	err = utils.CatchObject(response.Data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *BaseController) CheckMatchAddressWithUser(assetId, addressId int64, archived bool) bool {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/address-match-user"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assetid":       fmt.Sprintf("%d", assetId),
			"addressid":     fmt.Sprintf("%d", addressId),
			"archived":      fmt.Sprintf("%v", archived),
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return false
	}

	if response.IsError {
		return false
	}
	isMatch, isOk := response.Data.(bool)
	if !isOk {
		return false
	}
	return isMatch
}

func (this *BaseController) CheckAssetMatchUser(assetId int64) bool {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/asset-match-user"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assetid":       fmt.Sprintf("%d", assetId),
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return false
	}

	if response.IsError {
		return false
	}
	isMatch, isOk := response.Data.(bool)
	if !isOk {
		return false
	}
	return isMatch
}

func (this *BaseController) FilterAddressList(assetId int64, status string) ([]models.Addresses, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/filter-address-list"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assetid":       fmt.Sprintf("%d", assetId),
			"status":        status,
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return nil, err
	}

	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	var addrListRes []models.Addresses
	err := utils.CatchObject(response.Data, &addrListRes)
	if err != nil {
		return nil, err
	}
	return addrListRes, nil
}

func (this *BaseController) CheckAndCreateAccountToken(userId int64, username string, role int) (string, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/create-account-token"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"userid":        fmt.Sprintf("%d", userId),
			"username":      username,
			"role":          fmt.Sprintf("%d", role),
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return "", err
	}

	if response.IsError {
		return "", fmt.Errorf(response.Msg)
	}
	token, ok := response.Data.(string)
	if !ok {
		return "", fmt.Errorf("Get token failed")
	}
	return token, nil
}

func (this *BaseController) GetTxHistory(txHistoryId int64) (*models.TxHistory, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-txhistory"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"txhistoryid":   fmt.Sprintf("%d", txHistoryId),
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return nil, err
	}

	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	var txRes models.TxHistory
	err := utils.CatchObject(response.Data, &txRes)
	if err != nil {
		return nil, err
	}
	return &txRes, nil
}

func (this *BaseController) FilterUrlCodeList(assetType string, status string) ([]models.TxCode, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/filter-txcode"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assettype":     assetType,
			"status":        status,
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return nil, err
	}

	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}
	var resultData []models.TxCode
	err := utils.CatchObject(response.Data, &resultData)
	if err != nil {
		return nil, err
	}
	return resultData, nil
}

func (this *BaseController) CountAddressesWithStatus(assetId int64, activeFlg bool) int64 {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/count-address"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assetid":       fmt.Sprintf("%d", assetId),
			"activeflg":     fmt.Sprintf("%v", activeFlg),
		},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		return 0
	}

	if response.IsError {
		return 0
	}
	count, ok := response.Data.(int64)
	if ok {
		return 0
	}
	return count
}

func (this *BaseController) CheckHasCodeList(assetType string) bool {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/has-txcodes"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
			"assetType":     assetType,
		},
	}
	err := services.HttpRequest(req, &response)

	if err != nil || response.IsError {
		return false
	}
	check, ok := response.Data.(bool)
	if !ok {
		return false
	}
	return check
}

func (this *BaseController) GetContactList() []string {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-contacts"),
		Payload: map[string]string{
			"authorization": this.GetLoginToken(),
		},
	}
	err := services.HttpRequest(req, &response)

	if err != nil || response.IsError {
		return []string{}
	}
	var res []string
	err = utils.CatchObject(response.Data, &res)
	if err != nil {
		return []string{}
	}
	return res
}

func (this *BaseController) GetUserAssetList() ([]*models.Asset, error) {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/get-asset-list"),
		Header: map[string]string{
			"Authorization": this.GetLoginToken(),
			"AllowAssets":   utils.GetAllowAssets()},
	}

	var result []*models.Asset
	err := services.HttpRequest(req, &response)
	if err != nil {
		return result, err
	}
	if response.IsError {
		return result, fmt.Errorf(response.Msg)
	}
	err = utils.CatchObject(response.Data, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (this *BaseController) GetUsernameListExcludeId() []models.UserInfo {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AuthSite(), "/username-list"),
		Header:  map[string]string{"Authorization": this.GetLoginToken()},
	}
	usernameList := make([]models.UserInfo, 0)
	err := services.HttpRequest(req, &response)
	if err == nil && !response.IsError {
		jsonBytes, err := json.Marshal(response.Data)
		if err == nil {
			json.Unmarshal(jsonBytes, &usernameList)
		}
	}
	return usernameList
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

func (this *BaseController) AuthSite() string {
	return fmt.Sprintf("%s:%s", utils.GetAuthHost(), utils.GetAuthPort())
}

func (this *BaseController) ChatSite() string {
	return fmt.Sprintf("%s:%s", utils.GetChatHost(), utils.GetChatPort())
}

func (this *BaseController) AssetsSite() string {
	return fmt.Sprintf("%s:%s", utils.GetAssetsHost(), utils.GetAssetsPort())
}

// Check user is superadmin
func (this *BaseController) IsSuperAdmin(user models.AuthClaims) bool {
	return user.Role == int(utils.RoleSuperAdmin)
}

func (this *BaseController) SyncUsernameDB(userId int64, oldUsername, newUsername string) {
	this.SyncUsernameOnUserTable(userId, oldUsername, newUsername)
}

func (this *BaseController) SyncUsernameOnUserTable(userId int64, oldUsername, newUsername string) {
	var response utils.ResponseData
	if err := services.HttpFullPost(fmt.Sprintf("%s%s", this.AuthSite(), "/auth/syncChangeUsername"), this.Ctx.Request.Body, map[string]string{
		"Authorization": fmt.Sprintf("%s%s", "Bearer ", this.GetSession(utils.Tokenkey).(string)),
		"OldUsername":   oldUsername,
		"NewUsername":   newUsername,
	}, &response); err != nil {
		logpack.FError("Sync user data failed", userId, utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		logpack.FError(response.Msg, userId, utils.GetFuncName(), nil)
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

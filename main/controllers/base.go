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

func (this *BaseController) ResponseError(msg string, funcName string, err error) {
	logpack.Error(msg, funcName, err)
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

func (this *BaseController) AuthCheck() (*models.AuthClaims, error) {
	fmt.Println("check hereeeee222")
	authClaim, err := this.GetLoginUser()
	if err != nil {
		fmt.Println("check hereeeee")
		this.Redirect("/login", http.StatusFound)
		return nil, err
	}
	fmt.Println("333333333333333333")
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

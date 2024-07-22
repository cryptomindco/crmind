package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/services"
	"crmind/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type AdminController struct {
	BaseController
}

func (this *AdminController) Get() {
	authClaim, err := this.AuthCheck()
	if err != nil || authClaim.Role != int(utils.RoleSuperAdmin) {
		this.TplName = "err_403.html"
		return
	}
	var userList = make([]models.User, 0)
	//Get user list
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		Payload: map[string]string{},
		HttpUrl: fmt.Sprintf("%s%s", this.AuthSite(), "/admin/get-users"),
		Header:  map[string]string{"Authorization": this.GetLoginToken()},
	}
	err = services.HttpRequest(req, &response)
	if err == nil && !response.IsError {
		jsonBytes, err := json.Marshal(response.Data)
		if err == nil {
			json.Unmarshal(jsonBytes, &userList)
		}
	}
	this.Data["UserList"] = userList
	this.TplName = "admin/users.html"
}

func (this *AdminController) UserDetail() {
	authClaim, err := this.AuthCheck()
	if err != nil || authClaim.Role != int(utils.RoleSuperAdmin) {
		this.TplName = "err_403.html"
		return
	}
	var userId int64
	if err := this.Ctx.Input.Bind(&userId, "id"); err != nil {
		logpack.FError(err.Error(), userId, utils.GetFuncName(), nil)
		this.Redirect("/", http.StatusFound)
		return
	}
	//Get user info
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method: http.MethodGet,
		Payload: map[string]string{
			"userId": fmt.Sprintf("%d", userId),
		},
		HttpUrl: fmt.Sprintf("%s%s", this.AuthSite(), "/admin/user-info"),
		Header:  map[string]string{"Authorization": this.GetLoginToken()},
	}
	var targetUser models.User
	err = services.HttpRequest(req, &response)
	if err == nil && !response.IsError {
		jsonBytes, err := json.Marshal(response.Data)
		if err == nil {
			json.Unmarshal(jsonBytes, &targetUser)
		}
	}
	this.Data["User"] = targetUser
	logpack.Info(fmt.Sprintf("User Detail, Useid: %d", userId), utils.GetFuncName())
	this.TplName = "admin/user_detail.html"
}

func (this *AdminController) ChangeUserStatus() {
	_, err := this.SimpleAdminAuthCheck()
	if err != nil {
		this.ResponseError("Check login session failed", utils.GetFuncName(), err)
		return
	}
	userIdParam := this.GetString("userId")
	activeFlg := this.GetString("active")

	if utils.IsEmpty(userIdParam) || utils.IsEmpty(activeFlg) {
		this.ResponseError("Get user info param failed", utils.GetFuncName(), fmt.Errorf("Get user info param failed"))
		return
	}

	var response utils.ResponseData
	formData := url.Values{
		"userId":        {userIdParam},
		"active":        {activeFlg},
		"authorization": {this.GetLoginToken()},
	}
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/admin/change-user-status"), formData, &response); err != nil {
		this.ResponseError("Request change user status failed", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

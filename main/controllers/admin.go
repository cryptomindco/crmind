package controllers

import (
	"crmind/models"
	"crmind/services"
	"crmind/utils"
	"encoding/json"
	"fmt"
	"net/http"
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
	fmt.Println("Check heereeeee111")
	var userList = make([]models.User, 0)
	//Get user list
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		Payload: map[string]string{},
		HttpUrl: fmt.Sprintf("%s%s", this.AuthSite(), "/get-users"),
		Header:  map[string]string{"Authorization": this.GetLoginToken()},
	}
	err = services.HttpRequest(req, &response)
	if err == nil && !response.IsError {
		jsonBytes, err := json.Marshal(response.Data)
		if err == nil {
			json.Unmarshal(jsonBytes, &userList)
		}
	}
	fmt.Println("No error: ", len(userList))
	this.Data["UserList"] = userList
	this.TplName = "admin/users.html"
}

package controllers

import (
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/url"
)

type AuthController struct {
	BaseController
}

func (this *AuthController) BeginRegistration() {
	username := this.GetString("username")
	formData := url.Values{
		"username": {username},
	}

	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/registerStart"), formData, &response); err != nil {
		this.ResponseError("can't begin registration", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) FinishRegistration() {
}

func (this *AuthController) AuthSite() string {
	return fmt.Sprintf("%s:%s", utils.GetAuthHost(), utils.GetAuthPort())
}

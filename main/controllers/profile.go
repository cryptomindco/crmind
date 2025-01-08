package controllers

import (
	"crmind/pb/authpb"
	"crmind/services"
	"crmind/utils"
)

type ProfileController struct {
	BaseController
}

func (this *ProfileController) Get() {
	_, err := this.AuthCheck()
	if err != nil {
		return
	}
	this.TplName = "profile/profile.html"
}

func (this *ProfileController) UpdatePassword() {
	_, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed", utils.GetFuncName(), err)
		return
	}
	password := this.GetString("newpassword")
	if utils.IsEmpty(password) {
		this.ResponseError("Password cannot be blank", utils.GetFuncName(), nil)
		return
	}
	_, err = services.UpdatePassword(this.Ctx.Request.Context(), &authpb.WithPasswordRequest{
		Common: &authpb.CommonRequest{
			AuthToken: this.GetLoginToken(),
		},
		Password: password,
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(0, "Update password successfully", utils.GetFuncName())
}

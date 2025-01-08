package controllers

import (
	"crmind/models"
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

func (this *ProfileController) UpdateUsername() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("Check login session failed", utils.GetFuncName(), err)
		return
	}
	newUsername := this.GetString("newUsername")
	if utils.IsEmpty(newUsername) {
		this.ResponseError("New username cannot be blank", utils.GetFuncName(), nil)
		return
	}
	resData, err := services.UpdateUsername(this.Ctx.Request.Context(), &authpb.WithPasswordRequest{
		Common: &authpb.CommonRequest{
			AuthToken: this.GetLoginToken(),
		},
		Username: newUsername,
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(resData.Data, &data)
	if err != nil {
		this.ResponseError("Parse res data failed", utils.GetFuncName(), err)
	}
	var authClaim models.AuthClaims
	tokenString := ""
	tokenString = data["token"].(string)
	err = utils.CatchObject(data["user"], &authClaim)
	if err != nil {
		this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
		return
	}
	//set token on session
	this.SetSession(utils.LoginUserKey, authClaim)
	this.SetSession(utils.Tokenkey, tokenString)
	this.SyncUsernameDB(authClaim.Id, loginUser.Username, authClaim.Username)
	this.ResponseSuccessfully(0, "Update username successfully", utils.GetFuncName())
}

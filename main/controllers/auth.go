package controllers

import (
	"crmind/models"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/url"
	"strings"
)

type AuthController struct {
	BaseController
}

func (this *AuthController) Get() {
	this.TplName = "login.html"
}

func (this *AuthController) CheckUser() {
	username := strings.TrimSpace(this.GetString("username"))
	checkUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/check-user")
	var response utils.ResponseData
	if err := services.HttpGet(checkUrl, map[string]string{
		"username": username,
	}, &response); err != nil {
		this.ResponseError("Check user failed", utils.GetFuncName(), err)
		return
	}

	this.Data["json"] = response
	this.ServeJSON()
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
	reqUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/registerFinish")
	Headers := map[string]string{
		"Content-Type": "application/json",
		"Session-Key":  this.Ctx.Request.Header.Get("Session-Key"),
	}

	var response utils.ResponseData
	if err := services.HttpFullPost(reqUrl, this.Ctx.Request.Body, Headers, &response); err != nil {
		this.ResponseError("Register new user failed", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	data, isOk := response.Data.(map[string]any)
	tokenString := ""
	var authClaim models.AuthClaims
	if isOk {
		tokenString, _ = data["token"].(string)
		err := utils.CatchObject(data["user"], &authClaim)
		if err != nil {
			this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
			return
		}
	} else {
		this.ResponseError("Get login token failed", utils.GetFuncName(), fmt.Errorf("Get login token failed"))
		return
	}
	//set token on session
	this.SetSession(utils.Tokenkey, tokenString)
	this.SetSession(utils.LoginUserKey, authClaim)
	this.ResponseSuccessfully(0, "Registration successfully. Logging in...", utils.GetFuncName())
}

func (this *AuthController) AssertionOptions() {
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/assertion/options"), url.Values{}, &response); err != nil {
		this.ResponseError("can't begin login", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) AssertionResult() {
	reqUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/assertion/result")
	Headers := map[string]string{
		"Content-Type": "application/json",
		"Session-Key":  this.Ctx.Request.Header.Get("Session-Key"),
	}

	var response utils.ResponseData
	if err := services.HttpFullPost(reqUrl, this.Ctx.Request.Body, Headers, &response); err != nil {
		this.ResponseError("Login failed", utils.GetFuncName(), err)
		return
	}

	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	data, isOk := response.Data.(map[string]any)
	tokenString := ""
	var authClaim models.AuthClaims
	if isOk {
		tokenString, _ = data["token"].(string)
		err := utils.CatchObject(data["user"], &authClaim)
		if err != nil {
			this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
			return
		}
	} else {
		this.ResponseError("Get login token failed", utils.GetFuncName(), fmt.Errorf("Get login token failed"))
		return
	}
	//set token on session
	this.SetSession(utils.Tokenkey, tokenString)
	this.SetSession(utils.LoginUserKey, authClaim)
	this.ResponseSuccessfully(0, "Login successfully", utils.GetFuncName())
}

func (this *AuthController) BeginUpdatePasskey() {
	//Get login sesssion token string
	loginToken := this.GetSession(utils.Tokenkey).(string)
	if utils.IsEmpty(loginToken) {
		this.ResponseError("Get login token failed", utils.GetFuncName(), fmt.Errorf("Get login token failed"))
		return
	}
	formData := url.Values{
		"authorization": {fmt.Sprintf("%s%s", "Bearer ", loginToken)},
	}
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/updateStart"), formData, &response); err != nil {
		this.ResponseError("can't begin update passkey", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) FinishUpdatePasskey() {
	reqUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/updateFinish")
	Headers := map[string]string{
		"Content-Type":  "application/json",
		"Session-Key":   this.Ctx.Request.Header.Get("Session-Key"),
		"Is-Reset-Key":  this.Ctx.Request.Header.Get("Is-Reset-Key"),
		"Authorization": fmt.Sprintf("%s%s", "Bearer ", this.GetSession(utils.Tokenkey).(string)),
	}

	var response utils.ResponseData
	if err := services.HttpFullPost(reqUrl, this.Ctx.Request.Body, Headers, &response); err != nil {
		this.ResponseError("Update passkey failed", utils.GetFuncName(), err)
		return
	}
	data, isOk := response.Data.(map[string]any)
	var authClaim models.AuthClaims
	tokenString := ""
	if isOk {
		tokenString = data["token"].(string)
		err := utils.CatchObject(data["user"], &authClaim)
		if err != nil {
			this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
			return
		}
	} else {
		this.ResponseError("Get login user data failed", utils.GetFuncName(), fmt.Errorf("Get login user data failed"))
		return
	}
	//set token on session
	this.SetSession(utils.LoginUserKey, authClaim)
	this.SetSession(utils.Tokenkey, tokenString)
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) BeginConfirmPasskey() {
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/confirmStart"), url.Values{}, &response); err != nil {
		this.ResponseError("can't begin confirm passkey", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) FinishConfirmPasskey() {
	reqUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/confirmFinish")
	Headers := map[string]string{
		"Content-Type": "application/json",
		"Session-Key":  this.Ctx.Request.Header.Get("Session-Key"),
	}

	var response utils.ResponseData
	if err := services.HttpFullPost(reqUrl, this.Ctx.Request.Body, Headers, &response); err != nil {
		this.ResponseError("Confirm passkey failed", utils.GetFuncName(), err)
		return
	}

	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) CancelRegister() {
	sessionKey := this.GetString("sessionKey")
	formData := url.Values{
		"sessionKey": {sessionKey},
	}
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/cancelRegister"), formData, &response); err != nil {
		this.ResponseError("can't begin registration", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) Quit() {
	logoutUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/exit")
	var response utils.ResponseData
	services.HttpGet(logoutUrl, map[string]string{}, &response)
	this.Redirect("/login", 302)
	this.StopRun()
}

func (this *AuthController) GenRandomUsername() {
	var response utils.ResponseData
	if err := services.HttpGet(fmt.Sprintf("%s%s", this.AuthSite(), "/gen-random-username"), map[string]string{}, &response); err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) ChangeUsernameFinish() {
	reqUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/changeUsernameFinish")
	oldUserName := this.Ctx.Request.Header.Get("Old-Username")

	Headers := map[string]string{
		"Content-Type": "application/json",
		"Session-Key":  this.Ctx.Request.Header.Get("Session-Key"),
		"Old-Username": oldUserName,
	}

	var response utils.ResponseData
	if err := services.HttpFullPost(reqUrl, this.Ctx.Request.Body, Headers, &response); err != nil {
		this.ResponseError("Change username failed", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	data, isOk := response.Data.(map[string]any)
	var authClaim models.AuthClaims
	tokenString := ""
	if isOk {
		tokenString = data["token"].(string)
		err := utils.CatchObject(data["user"], &authClaim)
		if err != nil {
			this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
			return
		}
	} else {
		this.ResponseError("Get login user data failed", utils.GetFuncName(), fmt.Errorf("Get login user data failed"))
		return
	}
	//set token on session
	this.SetSession(utils.LoginUserKey, authClaim)
	this.SetSession(utils.Tokenkey, tokenString)
	this.SyncUsernameDB(authClaim.Id, oldUserName, authClaim.Username)
	this.ResponseSuccessfully(0, "Registration successfully. Logging in...", utils.GetFuncName())
}

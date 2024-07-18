package controllers

import (
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
	_, err := this.IsLoggingOn()
	if err != nil {
		this.TplName = "login.html"
		return
	}
	this.TplName = "index.html"
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

	this.Data["json"] = response
	this.ServeJSON()
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

	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) BeginUpdatePasskey() {
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/updateStart"), url.Values{}, &response); err != nil {
		this.ResponseError("can't begin update passkey", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AuthController) FinishUpdatePasskey() {
	reqUrl := fmt.Sprintf("%s%s", this.AuthSite(), "/passkey/updateFinish")
	Headers := map[string]string{
		"Content-Type": "application/json",
		"Session-Key":  this.Ctx.Request.Header.Get("Session-Key"),
		"Is-Reset-Key": this.Ctx.Request.Header.Get("Is-Reset-Key"),
	}

	var response utils.ResponseData
	if err := services.HttpFullPost(reqUrl, this.Ctx.Request.Body, Headers, &response); err != nil {
		this.ResponseError("Update passkey failed", utils.GetFuncName(), err)
		return
	}

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

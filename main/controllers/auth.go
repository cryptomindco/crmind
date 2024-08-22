package controllers

import (
	"crmind/models"
	"crmind/pb/authpb"
	"crmind/services"
	"crmind/utils"
	"fmt"
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
	res, err := services.CheckUserHandler(this.Ctx.Request.Context(), &authpb.WithUsernameRequest{
		Username: username,
	})
	if err != nil {
		this.ResponseError("Check user exist failed", utils.GetFuncName(), err)
		return
	}
	var resData map[string]bool
	err = utils.JsonStringToObject(res.Data, &resData)
	if err != nil {
		this.ResponseError("Parse res data failed", utils.GetFuncName(), err)
		return
	}
	exist := resData["exist"]
	this.ResponseSuccessfullyWithAnyData(0, "Check user successfully", utils.GetFuncName(), exist)
}

func (this *AuthController) BeginRegistration() {
	username := this.GetString("username")
	resData, err := services.BeginRegistrationHandler(this.Ctx.Request.Context(), &authpb.WithUsernameRequest{
		Username: username,
	})
	if err != nil {
		this.ResponseError("can't begin registration", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(0, "Begin registration successfully", utils.GetFuncName(), resData.Data)
}

func (this *AuthController) FinishRegistration() {
	sessionKey := this.Ctx.Request.Header.Get("Session-Key")
	res, err := services.FinishRegistrationHandler(this.Ctx.Request.Context(), &authpb.SessionKeyAndHttpRequest{
		SessionKey: sessionKey,
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(this.Ctx.Request.Body),
		},
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		this.ResponseError("Parse res data failed", utils.GetFuncName(), err)
	}
	var authClaim models.AuthClaims
	user, userExist := data["user"]
	token, tokenExist := data["token"]
	if !userExist || !tokenExist {
		this.ResponseError("Get login token failed", utils.GetFuncName(), fmt.Errorf("Get login token failed"))
		return
	}
	tokenString := token.(string)
	err = utils.CatchObject(user, &authClaim)
	if err != nil {
		this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
		return
	}
	//set token on session
	this.SetSession(utils.Tokenkey, tokenString)
	this.SetSession(utils.LoginUserKey, authClaim)
	this.ResponseSuccessfully(0, "Registration successfully. Logging in...", utils.GetFuncName())
}

func (this *AuthController) AssertionOptions() {
	responseData, err := services.AssertionOptionsHandler(this.Ctx.Request.Context())
	if err != nil {
		this.ResponseError("Get Assert options failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(0, "Start assertion options successfully", utils.GetFuncName(), responseData.Data)
}

func (this *AuthController) AssertionResult() {
	sessionKey := this.Ctx.Request.Header.Get("Session-Key")
	res, err := services.AssertionResultHandler(this.Ctx.Request.Context(), &authpb.SessionKeyAndHttpRequest{
		SessionKey: sessionKey,
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(this.Ctx.Request.Body),
		},
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		this.ResponseError("Parse res data failed", utils.GetFuncName(), err)
	}
	tokenString := ""
	var authClaim models.AuthClaims
	tokenString, _ = data["token"].(string)
	err = utils.CatchObject(data["user"], &authClaim)
	if err != nil {
		this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
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
	res, err := services.BeginUpdatePasskeyHandler(this.Ctx.Request.Context(), &authpb.CommonRequest{
		AuthToken: fmt.Sprintf("%s%s", "Bearer ", loginToken),
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = res
	this.ServeJSON()
}

func (this *AuthController) FinishUpdatePasskey() {
	res, err := services.FinishUpdatePasskeyHandler(this.Ctx.Request.Context(), &authpb.FinishUpdatePasskeyRequest{
		Common: &authpb.CommonRequest{
			AuthToken: fmt.Sprintf("%s%s", "Bearer ", this.GetSession(utils.Tokenkey).(string)),
		},
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(this.Ctx.Request.Body),
		},
		SessionKey: this.Ctx.Request.Header.Get("Session-Key"),
		IsReset:    this.Ctx.Request.Header.Get("Is-Reset-Key") == "true",
	})

	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		this.ResponseError("Parse res data failed", utils.GetFuncName(), err)
	}
	var authClaim models.AuthClaims
	tokenString := data["token"].(string)
	err = utils.CatchObject(data["user"], &authClaim)
	if err != nil {
		this.ResponseError("Parse login user failed", utils.GetFuncName(), err)
		return
	}
	//set token on session
	this.SetSession(utils.LoginUserKey, authClaim)
	this.SetSession(utils.Tokenkey, tokenString)
	this.ResponseSuccessfullyWithAnyData(0, "Finish update passkey successfully", utils.GetFuncName(), res.Data)
}

func (this *AuthController) CancelRegister() {
	res, err := services.CancelRegisterHandler(this.Ctx.Request.Context(), &authpb.CancelRegisterRequest{
		SessionKey: this.GetString("sessionKey"),
	})
	if err != nil {
		this.ResponseError("Cancel register failed", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = res
	this.ServeJSON()
}

func (this *AuthController) Quit() {
	this.DestroySession()
	this.Redirect("/login", 302)
	this.StopRun()
}

func (this *AuthController) GenRandomUsername() {
	res, err := services.GenRandomUsernameHandler(this.Ctx.Request.Context())
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = res
	this.ServeJSON()
}

func (this *AuthController) ChangeUsernameFinish() {
	oldUsername := this.Ctx.Request.Header.Get("Old-Username")
	res, err := services.ChangeUsernameFinishHandler(this.Ctx.Request.Context(), &authpb.ChangeUsernameFinishRequest{
		SessionKey:  this.Ctx.Request.Header.Get("Session-Key"),
		OldUsername: oldUsername,
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(this.Ctx.Request.Body),
		},
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}

	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
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
	this.SyncUsernameDB(authClaim.Id, oldUsername, authClaim.Username)
	this.ResponseSuccessfully(0, "Update username successfully", utils.GetFuncName())
}

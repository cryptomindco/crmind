package controllers

import (
	"crmind/logpack"
	"crmind/models"
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

func (this *BaseController) ResponseErrorWithErrName(errorName, msg, funcName string, err error) {
	logpack.Error(msg, funcName, err)
	this.Data["json"] = map[string]string{"error": errorName, "error_msg": msg}
	this.ServeJSON()
}

func (this *BaseController) ResponseError(msg string, funcName string, err error) {
	logpack.Error(msg, funcName, err)
	this.Data["json"] = map[string]string{"error": "true", "error_msg": msg}
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
	authClaim, err := this.GetLoginUser()
	if err != nil {
		this.Redirect("/login", http.StatusFound)
		return nil, err
	}
	this.Data["LoginUser"] = authClaim
	this.Data["IsSuperAdmin"] = this.IsSuperAdmin(*authClaim)
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

// Check user is superadmin
func (this *BaseController) IsSuperAdmin(user models.AuthClaims) bool {
	return user.Role == int(utils.RoleSuperAdmin)
}

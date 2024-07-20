package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/services"
	"crmind/utils"
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
	var response utils.ResponseData
	token := this.GetSession(utils.Tokenkey)
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AuthSite(), "/is-logging"),
		Payload: map[string]string{},
		Header:  map[string]string{"Authorization": fmt.Sprintf("%s%s", "Bearer ", token)},
	}

	if err := services.HttpRequest(req, &response); err != nil {
		return nil, err
	}

	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}

	var authClaim models.AuthClaims
	err := utils.CatchObject(response.Data, &authClaim)
	if err != nil {
		return nil, fmt.Errorf("Get Authentication info failed")
	}
	return &authClaim, nil
}

func (this *BaseController) AuthSite() string {
	return fmt.Sprintf("%s:%s", utils.GetAuthHost(), utils.GetAuthPort())
}

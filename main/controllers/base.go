package controllers

import (
	"crmind/logpack"
	"crmind/services"
	"crmind/utils"
	"fmt"

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

func (this *BaseController) IsLoggingOn() (int64, error) {
	var response utils.ResponseData
	if err := services.HttpGet(fmt.Sprintf("%s%s", this.AuthSite(), "/is-logging"), map[string]string{}, &response); err != nil {
		return 0, err
	}

	if response.IsError {
		return 0, fmt.Errorf(response.Msg)
	}
	return response.Data.(int64), nil
}

func (this *BaseController) AuthSite() string {
	return fmt.Sprintf("%s:%s", utils.GetAuthHost(), utils.GetAuthPort())
}

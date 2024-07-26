package controllers

import (
	"crchat/dohttp"
	"crchat/logpack"
	"crchat/models"
	"crchat/utils"
	"encoding/json"
	"fmt"
	"net/http"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
)

var (
	chatMsgModel     = new(models.ChatMsg)
	chatContentModel = new(models.ChatContent)
)

var (
	leftTreeResultMap = make(map[int][]orm.Params)
)

type BaseController struct {
	beego.Controller
}

func (this *BaseController) GetChatMsgFromId(chatId int64) (*models.ChatMsg, error) {
	o := orm.NewOrm()
	chatMsg := models.ChatMsg{}
	queryErr := o.QueryTable(chatMsgModel).Filter("id", chatId).Limit(1).One(&chatMsg)
	if queryErr != nil {
		return nil, queryErr
	}
	return &chatMsg, nil
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

func (this *BaseController) ResponseLoginRollbackError(loginId int64, tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseLoginError(loginId, msg, funcName, err)
}

func (this *BaseController) ResponseLoginError(loginId int64, msg string, funcName string, err error) {
	if loginId <= 0 {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginId, funcName, err)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: true,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseError(msg string, funcName string, err error) {
	logpack.Error(msg, funcName, err)
	this.Data["json"] = utils.ResponseData{
		IsError: true,
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
		Data:    result,
	}
	this.ServeJSON()
}

func (this *BaseController) AuthTokenCheck(token string) (*models.AuthClaims, error) {
	var response utils.ResponseData
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", utils.AuthSite(), "/is-logging"),
		Payload: map[string]string{},
		Header: map[string]string{
			"Authorization": token,
		},
	}

	err := dohttp.HttpRequest(req, &response)
	if err != nil {
		return nil, err
	}

	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}

	bytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}
	var authRes models.AuthClaims
	err = json.Unmarshal(bytes, &authRes)
	if err != nil {
		return nil, err
	}
	return &authRes, nil
}

func (this *BaseController) AuthCheck() (*models.AuthClaims, error) {
	authen := this.Ctx.Request.Header.Get("Authorization")
	return this.AuthTokenCheck(authen)
}

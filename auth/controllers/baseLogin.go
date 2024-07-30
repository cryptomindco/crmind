package controllers

import (
	"auth/logpack"
	"auth/models"
	"auth/utils"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/golang-jwt/jwt/v4"
)

type BaseLoginController struct {
	BaseController
}

func (this *BaseLoginController) CheckLoggingIn() (*models.AuthClaims, bool) {
	var bearer = this.Ctx.Request.Header.Get("Authorization")
	return this.HanlderCheckLogin(bearer)
}

func (this *BaseLoginController) HanlderCheckLogin(bearer string) (*models.AuthClaims, bool) {
	// Should be a bearer token
	if len(bearer) > 6 && strings.ToUpper(bearer[0:7]) == "BEARER " {
		var tokenStr = bearer[7:]
		var claim models.AuthClaims
		_, err := jwt.ParseWithClaims(tokenStr, &claim, func(token *jwt.Token) (interface{}, error) {
			return []byte(utils.GetConfValue("hmacSecretKey")), nil
		})
		if err != nil || claim.Id <= 0 {
			return nil, false
		}
		return &claim, true
	}
	return nil, false
}

func (this *BaseLoginController) GetUsernameListExcludeId(loginUserId int64) []*models.UserInfo {
	userList := this.GetUserListWithExcludeId(loginUserId)
	result := make([]*models.UserInfo, 0)
	for _, user := range userList {
		result = append(result, &models.UserInfo{
			Id:       user.Id,
			Username: user.Username,
		})
	}
	return result
}

func (this *BaseLoginController) ResponseLoginRollbackError(loginId int64, tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseLoginError(loginId, msg, funcName, err)
}

func (this *BaseLoginController) ResponseRollbackError(tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseError(msg, funcName, err)
}

func (this *BaseLoginController) ResponseLoginError(loginId int64, msg string, funcName string, err error) {
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

func (this *BaseLoginController) ResponseError(msg string, funcName string, err error) {
	logpack.Error(msg, funcName, err)
	this.Data["json"] = utils.ResponseData{
		IsError: true,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseLoginController) ResponseSuccessfully(loginId int64, msg string, funcName string) {
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

func (this *BaseLoginController) ResponseSuccessfullyWithAnyData(loginId int64, msg, funcName string, result any) {
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

func (this *BaseLoginController) GetIntArrayFromString(input string, separator string) []int64 {
	result := make([]int64, 0)
	if utils.IsEmpty(input) {
		return result
	}
	inputArr := strings.Split(input, separator)
	for _, intStr := range inputArr {
		if utils.IsEmpty(strings.TrimSpace(intStr)) {
			continue
		}
		intNum, err := strconv.ParseInt(strings.TrimSpace(intStr), 0, 32)
		if err != nil {
			continue
		}
		result = append(result, intNum)
	}
	return result
}

func (this *BaseLoginController) GetUserListWithExcludeId(excludeId int64) []models.User {
	userList := make([]models.User, 0)
	o := orm.NewOrm()
	_, listErr := o.QueryTable(userModel).Exclude("id", excludeId).OrderBy("createdt").All(&userList)
	if listErr != nil {
		return userList
	}
	return userList
}

func (this *BaseLoginController) CreateAuthClaimSession(loginUser *models.User) (string, *models.AuthClaims, error) {
	aliveSessionHourStr := utils.GetConfValue("aliveSessionHours")
	aliveSessionHours, err := strconv.ParseInt(aliveSessionHourStr, 0, 32)
	if err != nil {
		aliveSessionHours = utils.AliveSessionHours
	}
	authClaims := models.AuthClaims{
		Id:       loginUser.Id,
		Username: loginUser.Username,
		Expire:   time.Now().Add(time.Hour * time.Duration(aliveSessionHours)).Unix(),
		Role:     loginUser.Role,
		Token:    loginUser.Token,
		Contacts: loginUser.Contacts,
		Createdt: loginUser.Createdt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)
	tokenString, err := token.SignedString([]byte(utils.GetConfValue("hmacSecretKey")))
	if err != nil {
		return "", nil, err
	}
	return tokenString, &authClaims, nil
}

package controllers

import (
	"auth/logpack"
	"auth/models"
	"auth/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/golang-jwt/jwt/v4"
)

type BaseLoginController struct {
	BaseController
}

func (this *BaseLoginController) AuthCheck() (*models.SessionUser, bool) {
	userData, check := this.LoginUser()
	if !check {
		this.Redirect("/login", http.StatusFound)
		return nil, false
	}
	this.Data["LoginUser"] = userData
	this.Data["IsSuperAdmin"] = this.IsSuperAdmin(*userData)
	successMsg := this.GetSession("successMessage")
	userListJson := this.GetSession("UserListSessionKey")
	usernameList := make([]string, 0)
	//if userList is empty, get userList
	if utils.IsEmpty(userListJson) {
		usernameList = this.GetUsernameListExcludeId(userData.Id)
	} else {
		usernamesJsonBytes, err := json.Marshal(userListJson)
		if err == nil {
			json.Unmarshal(usernamesJsonBytes, &usernameList)
		}
	}
	this.Data["UsernameList"] = usernameList

	if !utils.IsEmpty(successMsg) {
		this.Data["successFlag"] = true
		this.SetSession("successMessage", "")
		this.Data["successfullyMsg"] = successMsg
	} else {
		this.Data["successFlag"] = false
	}
	return userData, true
}

func (this *BaseLoginController) CheckLoggingIn() (*models.AuthClaims, bool) {
	var bearer = this.Ctx.Request.Header.Get("Authorization")
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

func (this *BaseLoginController) SimpleAuthCheck() (*models.SessionUser, bool) {
	userData, check := this.LoginUser()
	if !check || userData.Id <= 0 {
		return nil, false
	}
	return userData, true
}

func (this *BaseLoginController) SimpleSuperadminAuthCheck() (*models.SessionUser, bool) {
	loginUser, check := this.SimpleAuthCheck()
	if !check {
		return nil, false
	}
	//if not superadmin
	if loginUser.Role != int(utils.RoleSuperAdmin) {
		return nil, false
	}

	return loginUser, true
}

func (this *BaseLoginController) SuperadminAuthCheck() (*models.SessionUser, bool) {
	loginUser, check := this.AuthCheck()
	if !check {
		return nil, false
	}
	//if not superadmin
	if loginUser.Role != int(utils.RoleSuperAdmin) {
		return nil, false
	}

	return loginUser, true
}

func (this *BaseLoginController) GetUsernameListExcludeId(loginUserId int64) []string {
	userList := this.GetUserListWithExcludeId(loginUserId)
	result := make([]string, 0)
	for _, user := range userList {
		result = append(result, user.Username)
	}
	return result
}

func (this *BaseLoginController) ResponseLoginRollbackError(loginUser *models.SessionUser, tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseLoginError(loginUser, msg, funcName, err)
}

func (this *BaseLoginController) ResponseRollbackError(tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseError(msg, funcName, err)
}

func (this *BaseLoginController) ResponseLoginError(loginUser *models.SessionUser, msg string, funcName string, err error) {
	if loginUser == nil {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginUser.User.Id, funcName, err)
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

func (this *BaseLoginController) ResponseSuccessfully(loginUser *models.SessionUser, msg string, funcName string) {
	if loginUser == nil {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginUser.User.Id, funcName)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseLoginController) ResponseSuccessfullyWithAnyData(loginUser *models.SessionUser, msg, funcName string, result any) {
	if loginUser == nil {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginUser.User.Id, funcName)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Data:    result,
	}
	this.ServeJSON()
}

func (this *BaseLoginController) LoginUser() (*models.SessionUser, bool) {
	userData, check := this.CheckLogin()
	if !check {
		return nil, false
	}
	userJson, _ := json.Marshal(userData)
	var loginUser = models.SessionUser{}
	json.Unmarshal(userJson, &loginUser)
	return &loginUser, true
}

func (this *BaseLoginController) IsLogin() bool {
	_, check := this.CheckLogin()
	return check
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
		Id:           loginUser.Id,
		Username:     loginUser.Username,
		Expire:       time.Now().Add(time.Hour * time.Duration(aliveSessionHours)).Unix(),
		Role:         loginUser.Role,
		Token:        loginUser.Token,
		Contacts:     loginUser.Contacts,
		CredsArrJson: loginUser.CredsArrJson,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)
	tokenString, err := token.SignedString([]byte(utils.GetConfValue("hmacSecretKey")))
	if err != nil {
		return "", nil, err
	}
	return tokenString, &authClaims, nil
}

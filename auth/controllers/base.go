package controllers

import (
	"auth/models"
	"auth/utils"
	"reflect"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
)

var (
	userModel = new(models.User)
)

var (
	leftTreeResultMap = make(map[int][]orm.Params)
)

type BaseController struct {
	beego.Controller
}

// Check user login status using session info
func (this *BaseController) CheckLogin() (interface{}, bool) {
	userData := this.GetSession("LoginUser")
	if userData == nil {
		return nil, false
	}
	return userData, true
}

// Check user login status using session info
func (this *BaseController) GetLoginUserName() (string, bool) {
	userData := this.GetSession("LoginUser")
	userReflect := reflect.ValueOf(userData)
	loginUserName := userReflect.FieldByName("Username")
	if loginUserName.Kind() == 0 {
		return "nil", false
	}
	return loginUserName.String(), true
}

// Check user login status using session info
func (this *BaseController) GetLoginId() (int64, bool) {
	userData := this.GetSession("LoginUser")
	userReflect := reflect.ValueOf(userData)
	loginUserId := userReflect.FieldByName("Id")
	if loginUserId.Kind() == 0 {
		return 0, false
	}
	return loginUserId.Int(), true
}

// Check user is superadmin
func (this *BaseController) IsSuperAdmin(loginUser models.SessionUser) bool {
	return loginUser.Role == int(utils.RoleSuperAdmin)
}

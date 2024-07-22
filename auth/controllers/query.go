package controllers

import (
	"auth/models"
	"auth/utils"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
)

type QueryController struct {
	BaseLoginController
}

func (this *QueryController) GetAdminUserList() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		this.ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
		return
	}
	userList := make([]*models.User, 0)
	o := orm.NewOrm()
	_, listErr := o.QueryTable(userModel).Exclude("id", authClaims.Id).OrderBy("createdt").All(&userList)
	if listErr != nil && listErr != orm.ErrNoRows {
		this.ResponseError("Get user list failed", utils.GetFuncName(), fmt.Errorf("Get user list failed"))
		return
	}
	this.ResponseSuccessfullyWithAnyData(nil, "Get user list successfully", utils.GetFuncName(), userList)
}

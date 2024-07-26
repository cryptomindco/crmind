package controllers

import (
	"auth/models"
	"auth/utils"
	"fmt"
	"time"

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
	this.ResponseSuccessfullyWithAnyData(0, "Get user list successfully", utils.GetFuncName(), userList)
}

func (this *QueryController) GetAdminUserInfo() {
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
	userId, err := this.GetInt64("userId")
	if err != nil {
		this.ResponseError("Get user id param failed", utils.GetFuncName(), err)
		return
	}

	o := orm.NewOrm()
	//Get company by id
	user := models.User{}
	err = o.QueryTable(userModel).Filter("id", userId).Limit(1).One(&user)
	if err != nil {
		this.ResponseError("Retrieve user data failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(0, "Get user list successfully", utils.GetFuncName(), user)
}

func (this *QueryController) GetExcludeLoginUserNameList() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	listName := this.GetUsernameListExcludeId(authClaims.Id)
	this.ResponseSuccessfullyWithAnyData(authClaims.Id, "Get user name list successfully", utils.GetFuncName(), listName)
}

func (this *QueryController) ChangeUserStatus() {
	token := this.GetString("authorization")
	authClaims, isLogin := this.HanlderCheckLogin(token)
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		this.ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
		return
	}

	userIdParam, err := this.GetInt("userId")
	activeFlg, activeErr := this.GetInt("active")
	if err != nil || activeErr != nil {
		this.ResponseLoginError(authClaims.Id, "Parameter is corrupted. Please try again!", utils.GetFuncName(), nil)
		return
	}
	user := models.User{}
	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(authClaims.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	queryErr := o.QueryTable(userModel).Filter("id", userIdParam).Limit(1).One(&user)
	if queryErr != nil {
		this.ResponseLoginError(authClaims.Id, "Get user from DB error. Please try again!", utils.GetFuncName(), beginErr)
		return
	}

	user.Status = activeFlg
	user.Updatedt = time.Now().Unix()
	//update user
	_, updateErr := tx.Update(&user)
	if updateErr != nil {
		this.ResponseLoginRollbackError(authClaims.Id, tx, "Update User failed. Please try again!", utils.GetFuncName(), updateErr)
		return
	}
	tx.Commit()
	this.ResponseSuccessfully(authClaims.Id, "Update User successfully!", utils.GetFuncName())
}

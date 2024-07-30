package controllers

import (
	"auth/models"
	"auth/utils"
	"fmt"
	"strings"
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

func (this *QueryController) UpdateContacts() {
	authToken := this.GetString("authorization")
	authClaims, isLogin := this.HanlderCheckLogin(authToken)
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	o := orm.NewOrm()
	user := models.User{}
	err := o.QueryTable(userModel).Filter("id", authClaims.Id).Limit(1).One(&user)
	if err != nil {
		this.ResponseError("Retrieve user data failed", utils.GetFuncName(), err)
		return
	}
	contacts := this.GetString("contacts")
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	user.Contacts = contacts
	user.Updatedt = time.Now().Unix()
	_, err = tx.Update(&user)
	if err != nil {
		this.ResponseRollbackError(tx, "Update contacts failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(authClaims.Id, "Update user contacts successfully", utils.GetFuncName())
}

func (this *QueryController) GetContactList() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	contactList, err := utils.GetContactListFromUser(authClaims.Id)
	if err != nil {
		this.ResponseError("Get contact list failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(authClaims.Id, "Get user info successfully", utils.GetFuncName(), contactList)
}

func (this *QueryController) GetUserInfoByUsername() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	username := this.Ctx.Input.Query("username")
	if utils.IsEmpty(username) {
		this.ResponseError("Get username parame failed", utils.GetFuncName(), fmt.Errorf("Get username parame failed"))
		return
	}

	user, err := utils.GetUserByUsername(username, orm.NewOrm())
	if err != nil {
		this.ResponseError("Get user by username failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(authClaims.Id, "Get user info successfully", utils.GetFuncName(), models.UserInfo{
		Id:       user.Id,
		Username: user.Username,
		Token:    user.Token,
		Contacts: user.Contacts,
	})
}

func (this *QueryController) CheckAndCreateToken() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}

	token, err := utils.CheckAndCreateUserToken(*authClaims)
	if err != nil {
		this.ResponseError("Check and create user token failed", utils.GetFuncName(), err)
	}

	this.ResponseSuccessfullyWithAnyData(authClaims.Id, "Check and create user token successfully", utils.GetFuncName(), token)
}

func (this *QueryController) GetToken() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}

	o := orm.NewOrm()
	user := models.User{}
	err := o.QueryTable(userModel).Filter("id", authClaims.Id).Limit(1).One(&user)
	if err != nil {
		this.ResponseError("Retrieve user data failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(authClaims.Id, "Get user token successfully", utils.GetFuncName(), user.Token)
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

func (this *QueryController) CheckContactUser() {
	authClaims, isLogin := this.CheckLoggingIn()
	if !isLogin {
		this.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
		return
	}
	username := strings.TrimSpace(this.GetString("username"))
	if authClaims.Username == username {
		this.ResponseLoginError(authClaims.Id, "The recipient cannot be you", utils.GetFuncName(), nil)
		return
	}
	o := orm.NewOrm()
	//Check username exist
	userCount, err := o.QueryTable(userModel).Filter("username", username).Count()
	if err == nil && userCount > 0 {
		//check if user is setted on loginUser contacts
		contactList, contactErr := utils.GetContactListFromUser(authClaims.Id)
		if contactErr != nil {
			this.ResponseLoginError(authClaims.Id, "Parse Contact list failed", utils.GetFuncName(), nil)
			return
		}
		exist := utils.CheckUsernameExistOnContactList(username, contactList)
		result := struct {
			ContactExist bool `json:"contactExist"`
		}{
			ContactExist: exist,
		}
		resultStr, err := utils.ConvertToJsonString(result)
		if err != nil {
			this.ResponseLoginError(authClaims.Id, "Convert result json failed", utils.GetFuncName(), nil)
			return
		}
		this.ResponseSuccessfullyWithAnyData(authClaims.Id, "Username exist", utils.GetFuncName(), resultStr)
		return
	}
	//user not exist
	this.ResponseLoginError(authClaims.Id, "Username does not exist", utils.GetFuncName(), nil)
}

package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/pb/assetspb"
	"crmind/pb/authpb"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/client/orm"
)

type AdminController struct {
	BaseController
}

func (this *AdminController) Get() {
	authClaim, err := this.AuthCheck()
	if err != nil || authClaim.Role != int(utils.RoleSuperAdmin) {
		this.TplName = "err_403.html"
		return
	}
	var userList = make([]models.User, 0)
	res, err := services.GetAdminUserListHandler(this.Ctx.Request.Context(), &authpb.CommonRequest{
		AuthToken: this.GetLoginToken(),
	})

	if err == nil && !res.Error {
		utils.JsonStringToObject(res.Data, &userList)
	}
	this.Data["UserList"] = userList
	this.TplName = "admin/users.html"
}

func (this *AdminController) UserDetail() {
	authClaim, err := this.AuthCheck()
	if err != nil || authClaim.Role != int(utils.RoleSuperAdmin) {
		this.TplName = "err_403.html"
		return
	}
	var userId int64
	if err := this.Ctx.Input.Bind(&userId, "id"); err != nil {
		logpack.FError(err.Error(), userId, utils.GetFuncName(), nil)
		this.Redirect("/", http.StatusFound)
		return
	}

	res, err := services.GetAdminUserInfoHandler(this.Ctx.Request.Context(), &authpb.WithUserIdRequest{
		UserId: userId,
		Common: &authpb.CommonRequest{
			AuthToken: this.GetLoginToken(),
		},
	})
	var targetUser models.User
	if err == nil && !res.Error {
		utils.JsonStringToObject(res.Data, &targetUser)
	}
	this.Data["User"] = targetUser
	logpack.Info(fmt.Sprintf("User Detail, Useid: %d", userId), utils.GetFuncName())
	this.TplName = "admin/user_detail.html"
}

func (this *AdminController) ChangeUserStatus() {
	_, err := this.SimpleAdminAuthCheck()
	if err != nil {
		this.ResponseError("Check login session failed", utils.GetFuncName(), err)
		return
	}
	userId, userIdErr := this.GetInt64("userId")
	activeFlg, activeErr := this.GetInt64("active")
	if userIdErr != nil || activeErr != nil {
		this.ResponseError("Get user info param failed", utils.GetFuncName(), fmt.Errorf("Get user info param failed"))
		return
	}

	res, err := services.ChangeUserStatusHandler(this.Ctx.Request.Context(), &authpb.ChangeUserStatusRequest{
		Common: &authpb.CommonRequest{
			AuthToken: this.GetLoginToken(),
		},
		UserId: userId,
		Active: activeFlg,
	})
	if err != nil {
		this.ResponseError("Request change user status failed", utils.GetFuncName(), err)
	}
	this.Data["json"] = res
	this.ServeJSON()
}

func (this *AdminController) GetSettings() {
	_, err := this.AdminAuthCheck()
	if err != nil {
		this.TplName = "err_403.html"
		logpack.Error(err.Error(), utils.GetFuncName(), err)
		return
	}

	allows, err := utils.GetAllowAssetFromSettings()
	if err != nil {
		logpack.Error(err.Error(), utils.GetFuncName(), nil)
		this.TplName = "err_403.html"
		return
	}
	this.Data["AllowAssets"] = allows
	this.TplName = "admin/settings.html"
}

func (this *AdminController) UpdateSettings() {
	loginUser, check := this.SimpleAdminAuthCheck()
	if check != nil {
		this.ResponseError("There is no permission to access this feature", utils.GetFuncName(), nil)
		return
	}

	selectedAssetStr := this.GetString("selectedAsset")
	//Get Settings
	settings, settingErr := this.CheckSettingsExist()
	//if get settings has DB error
	if settingErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get Settings from DB error. Please try again!", utils.GetFuncName(), settingErr)
		return
	}
	isCreate := settings == nil
	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "Start DB transaction failed. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	if utils.IsEmpty(selectedAssetStr) {
		selectedAssetStr = "usd"
	}
	//if creating new settings
	if isCreate {
		newSettings := models.Settings{
			ActiveAssets: selectedAssetStr,
		}
		//insert to DB
		_, insertErr := tx.Insert(&newSettings)
		if insertErr != nil {
			this.ResponseLoginRollbackError(loginUser.Id, tx, "Insert new settings failed. Please try again!", utils.GetFuncName(), insertErr)
			return
		}
	} else {
		settings.ActiveAssets = selectedAssetStr
		//update DB
		_, updateErr := tx.Update(settings)
		if updateErr != nil {
			this.ResponseLoginRollbackError(loginUser.Id, tx, "Update settings failed. Please try again!", utils.GetFuncName(), updateErr)
			return
		}
	}
	//comit change
	tx.Commit()
	//set to global allow assets
	utils.AllowAssets = selectedAssetStr
	this.ResponseSuccessfully(loginUser.Id, "Update settings successfully!", utils.GetFuncName())
}

func (this *AdminController) SyncTransactions() {
	loginUser, err := this.SimpleAdminAuthCheck()
	if err != nil {
		this.ResponseError("There is no permission to access this feature", utils.GetFuncName(), nil)
		return
	}

	_, err = services.SyncTransactionsHandler(this.Ctx.Request.Context(), &assetspb.CommonRequest{
		Role:      int64(loginUser.Role),
		LoginName: loginUser.Username,
	})

	if err != nil {
		this.ResponseError("can't sync transactions", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Synchronized transaction successfully", utils.GetFuncName())
}

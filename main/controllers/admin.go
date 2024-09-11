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
		parseErr := utils.JsonStringToObject(res.Data, &targetUser)
		if parseErr != nil {
			this.TplName = "err_403.html"
			return
		}
	}
	//get asset list
	assetRes, err := services.GetAssetDBListHandler(this.Ctx.Request.Context(), &assetspb.GetAssetDBListRequest{
		Common: &assetspb.CommonRequest{
			LoginName: authClaim.Username,
		},
		Allowassets: utils.GetAllowAssets(),
		Username:    targetUser.Username,
	})
	if err != nil {
		this.TplName = "err_403.html"
		return
	}
	var assets []models.Asset
	parseAssetErr := utils.JsonStringToObject(assetRes.Data, &assets)
	if parseAssetErr != nil {
		assets = make([]models.Asset, 0)
	}
	this.Data["Assets"] = assets
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
		return
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
	settings, err := utils.GetSettings()
	if err != nil && err != orm.ErrNoRows {
		this.TplName = "err_403.html"
		logpack.Error(err.Error(), utils.GetFuncName(), err)
		return
	}
	allowAssetStr := ""
	activeServicesStr := ""
	exchange := ""
	priceSpread := float64(0)
	if err == nil {
		allowAssetStr = settings.ActiveAssets
		activeServicesStr = settings.ActiveServices
		exchange = settings.RateServer
		priceSpread = settings.PriceSpread
	}
	allows := utils.HanlderAllowAssetStr(allowAssetStr)
	activeServices := utils.HandlerActiveServiceStr(activeServicesStr)

	//exchange
	this.Data["AllowAssets"] = allows
	this.Data["ActiveServices"] = activeServices
	this.Data["Exchange"] = exchange
	this.Data["PriceSpread"] = priceSpread
	this.TplName = "admin/settings.html"
}

func (this *AdminController) UpdateSettings() {
	loginUser, check := this.SimpleAdminAuthCheck()
	if check != nil {
		this.ResponseError("There is no permission to access this feature", utils.GetFuncName(), nil)
		return
	}

	selectedAssetStr := this.GetString("selectedAsset")
	selectedServicesStr := this.GetString("selectedServices")
	exchange := this.GetString("exchange")
	priceSpread, pSpreadErr := this.GetFloat("priceSpread", 0)
	if pSpreadErr != nil {
		this.ResponseError("Price Spread param failed", utils.GetFuncName(), nil)
		return
	}
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
	if utils.IsEmpty(selectedServicesStr) {
		selectedServicesStr = "auth"
	}
	//if creating new settings
	if isCreate {
		newSettings := models.Settings{
			ActiveAssets:   selectedAssetStr,
			ActiveServices: selectedServicesStr,
			RateServer:     exchange,
			PriceSpread:    priceSpread,
		}
		//insert to DB
		_, insertErr := tx.Insert(&newSettings)
		if insertErr != nil {
			this.ResponseLoginRollbackError(loginUser.Id, tx, "Insert new settings failed. Please try again!", utils.GetFuncName(), insertErr)
			return
		}
	} else {
		settings.ActiveAssets = selectedAssetStr
		settings.ActiveServices = selectedServicesStr
		settings.RateServer = exchange
		settings.PriceSpread = priceSpread
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
	utils.ActiveServices = selectedServicesStr
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

func (this *AdminController) AdminUpdateBalance() {
	loginUser, err := this.GetLoginUser()
	if err != nil || loginUser.Role != int(utils.RoleSuperAdmin) {
		this.ResponseError("There is no permission to access this feature", utils.GetFuncName(), nil)
		return
	}

	inputValue, inputErr := this.GetFloat("input")
	username := this.GetString("username")
	if inputErr != nil || utils.IsEmpty(username) {
		this.ResponseError("Get params failed", utils.GetFuncName(), nil)
		return
	}
	typeStr := this.GetString("type")
	action := this.GetString("action")
	//get user from username
	userInfo, err := this.GetUserByUsername(username)
	if err != nil {
		this.ResponseError("Get user info failed", utils.GetFuncName(), err)
		return
	}
	res, err := services.AdminUpdateBalanceHandler(this.Ctx.Request.Context(), &assetspb.AdminBalanceUpdateRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
			Role:      int64(loginUser.Role),
		},
		Input:    inputValue,
		Username: username,
		UserRole: int64(userInfo.Role),
		Type:     typeStr,
		Action:   action,
	})
	if err != nil {
		this.ResponseError("Update user balance failed", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Update user balance successfully", utils.GetFuncName(), res.Data)
}

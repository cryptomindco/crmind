package controllers

import (
	"crassets/handler"
	"crassets/utils"

	"github.com/beego/beego/v2/client/orm"
)

type AdminController struct {
	BaseController
}

func (this *AdminController) SyncTransactions() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	if loginUser.Role != int(utils.RoleSuperAdmin) {
		this.ResponseError("Check admin login failed", utils.GetFuncName(), nil)
		return
	}
	//create task to sync transactions
	go func() {
		o := orm.NewOrm()
		handler.SystemSyncHandler(o)
	}()
	this.ResponseSuccessfully(loginUser.Id, "Synchronized transaction successfully", utils.GetFuncName())
}

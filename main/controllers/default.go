package controllers

import (
	"crmind/logpack"
	"crmind/utils"
)

type MainController struct {
	BaseController
}

func (this *MainController) Get() {
	authClaims, err := this.AuthCheck()
	if err != nil {
		logpack.FWarn("User is not logged in", 0, utils.GetFuncName())
		this.Redirect("/login", 302)
		this.StopRun()
	}
	this.Data["LoginUser"] = authClaims
	this.TplName = "index.html"
}

package controllers

import (
	"crmind/models"
	"fmt"
)

var (
	settingsModel = new(models.Settings)
)

type MainController struct {
	BaseController
}

func (this *MainController) Get() {
	fmt.Println("kdjfksadjfksdf")
	_, err := this.AuthCheck()
	if err != nil {
		this.TplName = "login.html"
		return
	}
	this.TplName = "index.html"
}

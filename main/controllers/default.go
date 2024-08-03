package controllers

import "fmt"

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

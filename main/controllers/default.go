package controllers

type MainController struct {
	BaseController
}

func (this *MainController) Get() {
	_, err := this.AuthCheck()
	if err != nil {
		return
	}
	this.TplName = "index.html"
}

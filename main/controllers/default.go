package controllers

type MainController struct {
	BaseController
}

func (this *MainController) Get() {
	_, err := this.IsLoggingOn()
	if err != nil {
		this.TplName = "login.html"
		return
	}

	this.TplName = "index.html"
}

package controllers

type ProfileController struct {
	BaseController
}

func (this *ProfileController) Get() {
	_, err := this.AuthCheck()
	if err != nil {
		return
	}
	this.TplName = "profile/profile.html"
}

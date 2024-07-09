package controllers

import (
	"auth/utils"
	"fmt"
)

type LoginController struct {
	BaseLoginController
}

// Logout handler
func (this *LoginController) Quit() {
	this.DestroySession()
	this.Redirect("/login", 302)
	this.StopRun()
}

func (this *LoginController) GenRandomUsername() {
	username, err := utils.GetNewRandomUsername()
	if err != nil {
		this.ResponseError("Create new username failed", utils.GetFuncName(), err)
		return
	}
	result := make(map[string]string)
	result["username"] = username
	this.ResponseSuccessfullyWithData(nil, fmt.Sprintf("Get random username: %s", username), utils.GetFuncName(), utils.ObjectToJsonString(result))
}

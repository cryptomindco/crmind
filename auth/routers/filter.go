package routers

import (
	"auth/models"

	"github.com/beego/beego/v2/adapter/context"
)

var (
	loginUrl                   = "/login"
	url404                     = "/404"
	filterExcludeURLMap        = make(map[string]int)
	filterOnlyLoginCheckURLMap = make(map[string]int)
)

var FilterCryptomind = func(ctx *context.Context) {
	//Determine whether the URL is excluded
	if _, ok := filterExcludeURLMap[ctx.Request.URL.Path]; ok {
		return
	}
	_, okLogin := ctx.Input.Session("LoginUser").(models.SessionUser)
	//Determine whether to only verify the login URL
	if _, ok := filterOnlyLoginCheckURLMap[ctx.Request.URL.Path]; okLogin && ok {
		return
	}
	if !okLogin {
		ctx.Redirect(302, loginUrl)
	}
}

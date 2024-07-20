package routers

import (
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/http"
	"strings"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
)

var (
	loginUrl                   = "/login"
	url404                     = "/404"
	filterExcludeURLMap        = make(map[string]int)
	filterOnlyLoginCheckURLMap = make(map[string]int)
)

var InitSetFilterUrl = func() {
	for _, excludeUrl := range utils.LoginExcludeUrl {
		filterExcludeURLMap[excludeUrl] = 1
	}
	checkLoginUrl := beego.AppConfig.String("filterOnlyLoginCheckURL")
	if len(checkLoginUrl) > 0 {
		checkLoginUrlSlice := strings.Split(checkLoginUrl, ",")
		if len(checkLoginUrlSlice) > 0 {
			for _, v := range checkLoginUrlSlice {
				filterOnlyLoginCheckURLMap[v] = 1
			}
		}
	}
}

var FilterCryptomind = func(ctx *context.Context) {
	//Determine whether the URL is excluded
	if _, ok := filterExcludeURLMap[ctx.Request.URL.Path]; ok {
		return
	}
	//check login token session
	token, ok := ctx.Input.Session(utils.Tokenkey).(string)
	if !ok {
		ctx.Redirect(302, loginUrl)
		return
	}
	var response utils.ResponseData
	okLogin := false
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", utils.AuthSite(), "/is-logging"),
		Payload: map[string]string{},
		Header: map[string]string{
			"Authorization": fmt.Sprintf("%s%s", "Bearer ", token),
		},
	}

	err := services.HttpRequest(req, &response)
	if err != nil {
		ctx.Redirect(302, loginUrl)
		return
	}

	okLogin = !response.IsError
	//Determine whether to only verify the login URL
	if _, ok := filterOnlyLoginCheckURLMap[ctx.Request.URL.Path]; okLogin && ok {
		return
	}
	if !okLogin {
		ctx.Redirect(302, loginUrl)
	}
}

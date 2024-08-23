package routers

import (
	"crmind/pb/authpb"
	"crmind/services"
	"crmind/utils"
	"fmt"
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
	okService := utils.CheckServiceValidUrl(ctx.Request.URL.Path)
	if !okService {
		ctx.Redirect(403, "/403")
		return
	}
	//check service and
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

	//check login with rpc client
	okLogin, err := services.CheckMiddlewareLogin(ctx.Request.Context(), &authpb.CommonRequest{
		AuthToken: fmt.Sprintf("%s%s", "Bearer ", token),
	})
	if err != nil || !okLogin {
		ctx.Redirect(302, loginUrl)
		return
	}
	//Determine whether to only verify the login URL
	if _, ok := filterOnlyLoginCheckURLMap[ctx.Request.URL.Path]; okLogin && ok {
		return
	}
}

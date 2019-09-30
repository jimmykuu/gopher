package middlewares

import (
	"fmt"

	"gitea.com/lunny/tango"
)

// ApiAuthHandler /api 中间件，用于判断用户是否登录
func ApiAuthHandler() tango.HandlerFunc {
	return func(ctx *tango.Context) {
		fmt.Println("this is my api auth tango handler")
		var path = ctx.Req().URL.Path
		if path == "/api/signin" || path == "/api/signup" {
			// 登录和注册状态下不用检查登录状态
			ctx.Next()
			return
		} else {
			// 检查登录情况
		}
		fmt.Println(ctx.Req().URL.Path)
		ctx.Next()
	}
}

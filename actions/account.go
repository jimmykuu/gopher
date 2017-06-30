package actions

import (
	"github.com/tango-contrib/renders"
)

// Signin 登录
type Signin struct {
	RenderBase
}

// Get /signup 登录页面
func (c *Signin) Get() error {
	return c.Render("account/signin.html", renders.T{
		"title": "登录",
	})
}

// Signup 注册
type Signup struct {
	RenderBase
}

// Get /signup 注册页面
func (c *Signup) Get() error {
	return c.Render("account/signup.html", renders.T{
		"title": "注册",
	})
}

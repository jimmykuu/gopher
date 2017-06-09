package actions

import (
	"github.com/tango-contrib/renders"
)

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

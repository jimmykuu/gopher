package actions

import (
	"github.com/tango-contrib/renders"
)

type Signin struct {
	RenderBase
}

func (c *Signin) Get() error {
	return c.Render("account/signin.html", renders.T{
		"title": "登录",
	})
}

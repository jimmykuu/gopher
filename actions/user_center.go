package actions

import (
	"github.com/tango-contrib/renders"
)

// UserCenter 显示和修改用户信息
type UserCenter struct {
	RenderBase
}

// Get /user_center
func (a *UserCenter) Get() error {
	return a.Render("user_center/index.html", renders.T{})
}
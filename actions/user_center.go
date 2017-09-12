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
	return a.Render("user_center/index.html", renders.T{
		"active": "profile",
	})
}

// UserChangePassword 用户修改密码
type UserChangePassword struct {
	UserCenter
}

// Get /user_center/change_password
func (a *UserChangePassword) Get() error {
	return a.Render("user_center/index.html", renders.T{
		"active": "changePassword",
	})
}

// UserFavorite 用户修改密码
type UserFavorite struct {
	UserCenter
}

// Get /user_center/favorites
func (a *UserFavorite) Get() error {
	return a.Render("user_center/index.html", renders.T{
		"active": "favorites",
	})
}

package actions

import (
	"github.com/tango-contrib/renders"
)

// UserCenter 显示和修改用户信息
type UserCenter struct {
	AuthRenderBase
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
	page := a.FormInt("p", 1)
	if page <= 0 {
		page = 1
	}

	pagination := NewPagination(a.User.TopicsCollected, PerPage)
	collects, err := pagination.Page(page)
	if err != nil {
		return err
	}

	return a.Render("user_center/index.html", renders.T{
		"active":   "favorites",
		"collects": collects,
	})
}

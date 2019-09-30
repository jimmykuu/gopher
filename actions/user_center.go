package actions

import (
	"gitea.com/tango/renders"
)

// UserCenter 显示和修改用户信息
type UserCenter struct {
	AuthRenderBase
}

// Get /user_center
func (a *UserCenter) Get() error {
	return a.Render("user_center/index.html", renders.T{
		"title":      "基本资料",
		"sub_active": "user_info",
	})
}

type UserProfile struct {
	AuthRenderBase
}

// Get /user_center/profile
func (a *UserProfile) Get() error {
	return a.Render("user_center/profile.html", renders.T{
		"title":      "详细资料",
		"sub_active": "user_profile",
	})
}

// ChangeAvatar 修改头像
type ChangeAvatar struct {
	AuthRenderBase
}

// Get /user_center/avatar
func (a *ChangeAvatar) Get() error {
	return a.Render("user_center/avatar.html", renders.T{
		"title":      "修改头像",
		"sub_active": "change_avatar",
	})
}

// ChangePassword 用户修改密码
type ChangePassword struct {
	AuthRenderBase
}

// Get /user_center/change_password
func (a *ChangePassword) Get() error {
	return a.Render("user_center/change_password.html", renders.T{
		"title":      "修改密码",
		"sub_active": "change_password",
	})
}

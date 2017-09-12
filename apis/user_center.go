package apis

import (
	"strings"

	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/gopher/utils"
	"github.com/pborman/uuid"
	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"
)

// UserCenter 用户中心
type UserCenter struct {
	Base
	binding.Binder
}

// ProfileForm 个人信息表单
type ProfileForm struct {
	Email          string `json:"email"`
	Website        string `json:"website"`
	Location       string `json:"location"`
	Tagline        string `json:"tagline"`
	Bio            string `json:"bio"`
	GithubUsername string `json:"github_username"`
	Weibo          string `json:"weibo"`
}

// UserProfile 用户信息
type UserProfile struct {
	UserCenter
}

// Put /api/user_center/profile 编辑个人信息
func (a *UserProfile) Put() interface{} {
	var form ProfileForm
	a.ReadJSON(&form)

	c := a.DB.C(models.USERS)
	err := c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{
		"email":          form.Email,
		"website":        form.Website,
		"location":       form.Location,
		"tagline":        form.Tagline,
		"bio":            form.Bio,
		"githubusername": form.GithubUsername,
		"weibo":          form.Weibo,
	}})

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "保存个人信息出错",
		}
	}

	return map[string]interface{}{
		"status": 1,
	}
}

// UserChangePassword 修改用户密码
type UserChangePassword struct {
	UserCenter
}

// ChangePasswordForm 用户密码表单
type ChangePasswordForm struct {
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

// Put /user_center/change_password
func (a *UserChangePassword) Put() interface{} {
	var form ChangePasswordForm
	a.ReadJSON(&form)

	if len(form.OldPassword) == 0 {
		return map[string]interface{}{
			"status":  0,
			"message": "原密码不能为空",
		}
	}

	if len(form.NewPassword) == 0 {
		return map[string]interface{}{
			"status":  0,
			"message": "新密码不能为空",
		}
	}

	if form.NewPassword != form.ConfirmPassword {
		return map[string]interface{}{
			"status":  0,
			"message": "新密码不一致",
		}
	}

	if !a.User.CheckPassword(form.OldPassword) {
		return map[string]interface{}{
			"status":  0,
			"message": "原密码不正确",
		}
	}

	c := a.DB.C(models.USERS)

	salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
	password := utils.EncryptPassword(form.NewPassword, salt, models.PublicSalt)
	err := c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{
		"password": password,
		"salt":     salt,
	}})

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "密码修改失败",
		}
	}

	return map[string]interface{}{
		"status":  1,
		"message": "密码修改成功",
	}
}

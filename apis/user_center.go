package apis

import (
	"fmt"
	"strings"

	"github.com/pborman/uuid"
	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/gopher/utils"
)

// UserInfo 用户信息
type UserInfo struct {
	Base
	binding.Binder
}

// Get /api/user_center/user_info
func (a *UserInfo) Get() interface{} {
	if !a.IsLogin {
		return map[string]interface{}{
			"status":  0,
			"message": "未登录，不能获取用户信息",
		}
	}

	return map[string]interface{}{
		"status":   1,
		"username": a.User.Username,
		"email":    a.User.Email,
		"avatar":   a.User.Avatar,
		"avatars": []string{
			a.User.AvatarImgSrc(128),
			a.User.AvatarImgSrc(64),
			a.User.AvatarImgSrc(32),
		},
		"website":  a.User.Website,
		"location": a.User.Location,
		"tagline":  a.User.Tagline,
		"bio":      a.User.Bio,
		"github":   a.User.GitHubUsername,
		"weibo":    a.User.Weibo,
	}
}

// UserProfile 用户中心
type UserProfile struct {
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
	Base
	binding.Binder
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

// UploadAvatarImage 上传头像图片
type UploadAvatarImage struct {
	Base
	binding.Binder
}

// Post /api/user_center/upload_avatar
func (a *UploadAvatarImage) Post() interface{} {
	filename, err := saveImage(a.Req(), []string{"avatar"}, 500*1024)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败（%s）", err.Error()),
		}
	}

	c := a.DB.C(models.USERS)
	c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{
		"avatar": filename,
	}})

	a.User.Avatar = filename

	return map[string]interface{}{
		"status": 1,
		"avatars": []string{
			a.User.AvatarImgSrc(128),
			a.User.AvatarImgSrc(64),
			a.User.AvatarImgSrc(32),
		},
	}
}

// DefaultAvatars 默认头像
type DefaultAvatars struct {
	Base
	binding.Binder
}

// Get /api/user_center/default_avatars
func (a *DefaultAvatars) Get() interface{} {
	var defaultAvatars = []string{
		"gopher_aqua.jpg",
		"gopher_boy.jpg",
		"gopher_brown.jpg",
		"gopher_gentlemen.jpg",
		"gopher_girl.jpg",
		"gopher_strawberry_bg.jpg",
		"gopher_strawberry.jpg",
		"gopher_teal.jpg",
	}

	return map[string]interface{}{
		"status":          1,
		"default_avatars": defaultAvatars,
	}
}

// SetAvatar 设置头像
type SetAvatar struct {
	Base
	binding.Binder
}

// Put /api/user_center/set_avatar
func (a *SetAvatar) Put() interface{} {
	var form struct {
		Avatar string `json:"avatar"`
	}

	a.ReadJSON(&form)

	c := a.DB.C(models.USERS)
	c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{"avatar": form.Avatar}})

	a.User.Avatar = form.Avatar

	return map[string]interface{}{
		"status": 1,
		"avatars": []string{
			a.User.AvatarImgSrc(128),
			a.User.AvatarImgSrc(64),
			a.User.AvatarImgSrc(32),
		},
	}
}

package apis

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
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

// BasicInfo 用户基本信息
type BasicInfo struct {
	Base
	binding.Binder
}

// Put /api/user_center/basic_info
func (a *BasicInfo) Put() interface{} {
	var form struct {
		Email string `json:"email" valid:"required,email"`
		Bio   string `json:"bio"`
	}

	a.ReadJSON(&form)

	result, err := govalidator.ValidateStruct(form)
	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	c := a.DB.C(models.USERS)
	err = c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{
		"email": form.Email,
		"bio":   form.Bio,
	}})

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"status":  1,
		"message": "修改成功",
	}
}

// UserProfile 用户详细信息
type UserProfile struct {
	Base
	binding.Binder
}

// Put /api/user_center/profile 编辑个人信息
func (a *UserProfile) Put() interface{} {
	var form struct {
		Location       string `json:"location"`
		Website        string `json:"website" valid:"url,optional"`
		Tagline        string `json:"tagline"`
		GithubUsername string `json:"github"`
		Weibo          string `json:"weibo"`
	}
	a.ReadJSON(&form)

	result, err := govalidator.ValidateStruct(form)
	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	c := a.DB.C(models.USERS)
	err = c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{
		"location":       form.Location,
		"website":        form.Website,
		"tagline":        form.Tagline,
		"githubusername": form.GithubUsername,
		"weibo":          form.Weibo,
	}})

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"status":  1,
		"message": "修改成功",
	}
}

// ChangePassword 修改用户密码
type ChangePassword struct {
	Base
	binding.Binder
}

// Put /user_center/change_password
func (a *ChangePassword) Put() interface{} {
	var form struct {
		OldPassword     string `json:"oldPassword" valid:"required,ascii"`
		NewPassword     string `json:"newPassword" valid:"required,ascii"`
		ConfirmPassword string `json:"confirmPassword" valid:"required,ascii"`
	}
	a.ReadJSON(&form)

	result, err := govalidator.ValidateStruct(form)
	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
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
	err = c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{
		"password": password,
		"salt":     salt,
	}})

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("密码修改失败（%s）", err.Error()),
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
	file, header, err := a.Req().FormFile("image")
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败（%s）", err.Error()),
		}
	}
	defer file.Close()

	fileType := header.Header["Content-Type"][0]

	filename, err := saveImage(file, fileType, "", []string{"avatar"}, 500*1024)
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

// FromGravatar 从 Gravatar 获取头像
type FromGravatar struct {
	Base
	binding.Binder
}

// Get /api/user_center/from_gravatar
func (a *FromGravatar) Get() interface{} {
	h := md5.New()
	io.WriteString(h, a.User.Email)
	url := fmt.Sprintf("http://www.gravatar.com/avatar/%x?s=%d", h.Sum(nil), 256)

	resp, err := http.Get(url)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("获取头像失败（%s）", err.Error()),
		}
	}
	defer resp.Body.Close()

	fileType := resp.Header["Content-Type"][0]

	filename, err := saveImage(resp.Body, fileType, "", []string{"avatar"}, -1)
	if err != nil {
		panic(err)
	}

	c := a.DB.C(models.USERS)
	c.Update(bson.M{"_id": a.User.Id_}, bson.M{"$set": bson.M{"avatar": filename}})

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

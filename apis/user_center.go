package apis

import (
	"github.com/jimmykuu/gopher/models"
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

type UserProfile struct {
	UserCenter
}

// Put /api/user_center/profile 编辑个人信息
func (a *UserProfile) Put() interface{} {
	user := a.User
	c := a.DB.C(models.USERS)

	var form ProfileForm
	a.ReadJSON(&form)

	err := c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{
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

type UserChangePassword struct {
	UserCenter
}

func (a *UserChangePassword) Put() error {
	return nil
}

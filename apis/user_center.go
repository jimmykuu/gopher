package apis

import (
	"github.com/jimmykuu/gopher/models"
	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"
)

// UserCenter 评论
type UserCenter struct {
	Base
	binding.Binder
}

// TopicForm 主题表单，新建和编辑共用
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
func (a *UserCenter) Put() interface{} {
	user := a.User
	c := a.DB.C(models.USERS)

	var form ProfileForm
	a.ReadJSON(&form)

	c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{
		"email":          form.Email,
		"website":        form.Website,
		"location":       form.Location,
		"tagline":        form.Tagline,
		"bio":            form.Bio,
		"githubusername": form.GithubUsername,
		"weibo":          form.Weibo,
	}})

	return map[string]interface{}{
		"status": 1,
	}
}
package actions

import (
	"github.com/lunny/tango"
	"github.com/tango-contrib/renders"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/gopher/utils"
)

// RenderBase 渲染基类
type RenderBase struct {
	renders.Renderer
	tango.Ctx
}

// Render 渲染模板
func (r *RenderBase) Render(tmpl string, t ...renders.T) error {

	var ts = renders.T{}

	if len(t) > 0 {
		ts = t[0].Merge(renders.T{})
	} else {
		ts = renders.T{}
	}

	user, has := r.currentUser()

	if has {
		ts["user"] = user
		ts["ussername"] = user.Username
	}

	return r.Renderer.Render(tmpl, ts)
}

// currentUser 当前用户
func (r *RenderBase) currentUser() (*models.User, bool) {
	cookies := r.Cookies()
	userID, err := cookies.String("user")
	if err != nil {
		return nil, false
	}

	userID2, err := utils.Base64Decode([]byte(userID))
	if err != nil {
		panic(err)
	}

	session, DB := models.GetSessionAndDB()
	defer session.Close()

	user := models.User{}

	c := DB.C(models.USERS)

	// 检查用户名
	err = c.Find(bson.M{"_id": bson.ObjectIdHex(string(userID2))}).One(&user)

	if err != nil {
		return nil, false
	}

	return &user, true
}

package actions

import (
	"fmt"

	"github.com/lunny/tango"
	"github.com/tango-contrib/renders"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/gopher/utils"
)

type RenderBase struct {
	renders.Renderer
	tango.Ctx
}

func (r *RenderBase) Render(tmpl string, t ...renders.T) error {

	var ts = renders.T{}

	if len(t) > 0 {
		ts = t[0].Merge(renders.T{})
	} else {
		ts = renders.T{}
	}

	user, has := r.currentUser()

	fmt.Println(user, has)

	if has {
		ts["user"] = user
	}

	return r.Renderer.Render(tmpl, ts)
}

func (r *RenderBase) currentUser() (*models.User, bool) {
	cookies := r.Cookies()
	userId, err := cookies.String("user")
	if err != nil {
		return nil, false
	}

	userId2, err := utils.Base64Decode([]byte(userId))
	if err != nil {
		panic(err)
	}

	session, DB := models.GetSessionAndDB()
	defer session.Close()

	user := models.User{}

	c := DB.C(models.USERS)

	// 检查用户名
	err = c.Find(bson.M{"_id": bson.ObjectIdHex(string(userId2))}).One(&user)

	if err != nil {
		return nil, false
	}

	return &user, true
}

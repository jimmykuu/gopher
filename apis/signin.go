package apis

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/gopher/utils"
)

// Signin 登录
type Signin struct {
	Base
	binding.Binder
}

// Post /api/signin 提交登录
func (a *Signin) Post() interface{} {
	var form struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	a.ReadJSON(&form)
	session, DB := models.GetSessionAndDB()
	defer session.Close()

	c := DB.C(models.USERS)
	user := models.User{}

	err := c.Find(bson.M{"username": form.Username}).One(&user)

	if err != nil {
		return map[string]interface{}{
			"status":   0,
			"messages": []string{"该用户不存在"},
		}
	}

	if !user.CheckPassword(form.Password) {
		return map[string]interface{}{
			"status":   0,
			"messages": []string{"密码和用户名不匹配"},
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id_.Hex(),
		"exp":     time.Now().AddDate(1, 0, 0).Unix(),
	})

	tokenString, err := token.SignedString([]byte("fwZ1owO330suuhtfb0zvjlrXSYREnyhG"))
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{
		"status": 1,
		"token":  tokenString,
		"cookie": map[string]interface{}{
			"user": string(utils.Base64Encode([]byte(user.Id_.Hex()))),
		},
	}
}

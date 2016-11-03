package apis

import (
	"encoding/base64"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

const (
	base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var coder = base64.NewEncoding(base64Table)

func base64Encode(src []byte) []byte {
	return []byte(coder.EncodeToString(src))
}

func base64Decode(src []byte) ([]byte, error) {
	return coder.DecodeString(string(src))
}

type Signin struct {
	Base
	binding.Binder
}

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
			"status":  0,
			"message": "该用户不存在",
		}
	}

	if !user.CheckPassword(form.Password) {
		return map[string]interface{}{
			"status":  0,
			"message": "密码和用户名不匹配",
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

	println(">>> src:", user.Id_.Hex())
	encode := string(base64Encode([]byte(user.Id_.Hex())))
	println(">>> encode", encode)
	decode, err := base64Decode([]byte(encode))
	if err != nil {
		panic(err)
	}
	println(">>> decode", decode)

	return map[string]interface{}{
		"status": 1,
		"token":  tokenString,
		"cookie": map[string]interface{}{
			"user": string(base64Encode([]byte(user.Id_.Hex()))),
		},
	}
}

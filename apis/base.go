package apis

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitea.com/lunny/tango"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2"

	"github.com/jimmykuu/gopher/models"
	"gopkg.in/mgo.v2/bson"
)

const JWT_KEY = "fwZ1owO330suuhtfb0zvjlrXSYREnyhG"

type Base struct {
	tango.Json
	tango.Ctx

	session *mgo.Session
	DB      *mgo.Database
	User    models.User
	IsLogin bool
}

func (b *Base) Before() {
	b.session, b.DB = models.GetSessionAndDB()

	var token = b.Req().Header.Get("Authorization")
	if strings.Index(token, "Bearer ") != 0 {
		fmt.Printf("Token 解析错误")
		return
	}
	var tokenString = token[7:]

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_KEY), nil
	})

	if err == nil && parsedToken.Valid {
		claims_, _ := parsedToken.Claims.(jwt.MapClaims)
		userID, _ := claims_["user_id"].(string)

		if !bson.IsObjectIdHex(userID) {
			fmt.Println("非法的用户 ID：", userID)
			return
		}

		var user models.User
		var c = b.DB.C(models.USERS)
		err := c.Find(bson.M{"_id": bson.ObjectIdHex(userID)}).One(&user)

		if err != nil {
			fmt.Println("没有找到用户：", userID)
			return
		}

		b.User = user
		b.IsLogin = true
	}
}

func (b *Base) After() {
	b.session.Close()
}

func (b *Base) ReadJSON(v interface{}) error {
	decoder := json.NewDecoder(b.Req().Body)
	return decoder.Decode(&v)
}

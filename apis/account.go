package apis

import (
	"fmt"
	"time"
	"strings"

	"github.com/pborman/uuid"
	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"
	"github.com/dgrijalva/jwt-go"

	"github.com/jimmykuu/gopher/utils"
	"github.com/jimmykuu/gopher/models"
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
	c := a.DB.C(models.USERS)
	user := models.User{}

	err := c.Find(bson.M{"username": form.Username}).One(&user)

	if err != nil {
		fmt.Println(err.Error())
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

	tokenString, err := token.SignedString([]byte(JWT_KEY))
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

// Signup 注册
type Signup struct {
	Base
	binding.Binder
}

// Post /api/signup 提交注册
func (a *Signup) Post() interface{} {
	var form struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	a.ReadJSON(&form)

	session, DB := models.GetSessionAndDB()
	defer session.Close()

	c := DB.C(models.USERS)

	result := models.User{}

	// TODO: 检查邮箱合法性，使用极验验证

	// 检查用户名
	err := c.Find(bson.M{"username": form.Username}).One(&result)
	if err == nil {
		return map[string]interface{}{
			"status":   0,
			"messages": []string{"该用户名已经被注册"},
		}
	}

	// 检查邮箱
	err = c.Find(bson.M{"email": form.Email}).One(&result)

	if err == nil {
		return map[string]interface{}{
			"status":   0,
			"messages": []string{"该电子邮件地址已经被注册"},
		}
	}

	c2 := DB.C(models.STATUS)
	var status models.Status
	c2.Find(nil).One(&status)

	id := bson.NewObjectId()
	validateCode := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
	salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
	index := status.UserIndex + 1
	user := &models.User{
		Id_:          id,
		Username:     form.Username,
		Password:     utils.EncryptPassword(form.Password, salt, models.PublicSalt),
		Avatar:       "", // defaultAvatars[rand.Intn(len(defaultAvatars))],
		Salt:         salt,
		Email:        form.Email,
		ValidateCode: validateCode,
		IsActive:     true,
		JoinedAt:     time.Now(),
		Index:        index,
	}

	err = c.Insert(user)
	if err != nil {
		return map[string]interface{}{
			"status":   0,
			"messages": []string{err.Error()},
		}
	}

	c2.Update(nil, bson.M{"$inc": bson.M{"userindex": 1, "usercount": 1}})

	return map[string]interface{}{
		"status": 1,
		"cookie": map[string]interface{}{
			"user": string(utils.Base64Encode([]byte(user.Id_.Hex()))),
		},
	}
}


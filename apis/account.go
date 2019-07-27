package apis

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitea.com/tango/binding"
	"github.com/Youngyezi/geetest"
	"github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"
	"github.com/jimmykuu/webhelpers"
	"github.com/pborman/uuid"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/conf"
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
		Username         string `json:"username" valid:"required,ascii"`
		Password         string `json:"password" valid:"required,ascii"`
		GeetestChallenge string `json:"geetest_challenge" valid:"required,ascii"`
		GeetestValidate  string `json:"geetest_validate" valid:"required,ascii"`
		GeetestCeccode   string `json:"geetest_seccode" valid:"required,ascii"`
	}

	a.ReadJSON(&form)

	result, err := govalidator.ValidateStruct(form)

	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	p := url.Values{
		"client": {"web"},
	}
	g := geetest.New(conf.Config.GtCaptchaId, conf.Config.GtPrivateKey)
	success := g.SuccessValidate(form.GeetestChallenge, form.GeetestValidate, form.GeetestCeccode, p)

	if success == 0 {
		return map[string]interface{}{
			"status":  0,
			"message": "验证码错误",
		}
	}

	c := a.DB.C(models.USERS)
	user := models.User{}

	if strings.Contains(form.Username, "@") {
		err = c.Find(bson.M{"email": bson.M{"$regex": form.Username, "$options": "i"}}).One(&user)
	} else {
		err = c.Find(bson.M{"username": bson.M{"$regex": form.Username, "$options": "i"}}).One(&user)
	}

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "该用户不存在",
		}
	}

	if !user.CheckPassword(form.Password) {
		return map[string]interface{}{
			"status":  0,
			"message": "密码和用户名/邮箱不匹配",
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
		Username         string `json:"username" valid:"required,ascii"`
		Password         string `json:"password" valid:"required,ascii"`
		Email            string `json:"email" valid:"required,email"`
		GeetestChallenge string `json:"geetest_challenge" valid:"required,ascii"`
		GeetestValidate  string `json:"geetest_validate" valid:"required,ascii"`
		GeetestCeccode   string `json:"geetest_seccode" valid:"required,ascii"`
	}

	a.ReadJSON(&form)

	result, err := govalidator.ValidateStruct(form)

	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	if strings.Contains(form.Username, " ") {
		return map[string]interface{}{
			"status":  0,
			"message": "用户名中不能有空格",
		}
	}

	c := a.DB.C(models.USERS)

	var user models.User

	// 检查用户名
	err = c.Find(bson.M{"username": bson.M{"$regex": form.Username, "$options": "i"}}).One(&user)
	if err == nil {
		return map[string]interface{}{
			"status":  0,
			"message": "该用户名已经被注册",
		}
	}

	// 检查邮箱
	err = c.Find(bson.M{"email": bson.M{"$regex": form.Email, "$options": "i"}}).One(&user)

	if err == nil {
		return map[string]interface{}{
			"status":  0,
			"message": "该电子邮件地址已经被注册",
		}
	}

	c2 := a.DB.C(models.STATUS)
	var status models.Status
	c2.Find(nil).One(&status)

	id := bson.NewObjectId()
	validateCode := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
	salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
	index := status.UserIndex + 1
	newUser := &models.User{
		Id_:          id,
		Username:     form.Username,
		Password:     utils.EncryptPassword(form.Password, salt, models.PublicSalt),
		Avatar:       "",
		Salt:         salt,
		Email:        form.Email,
		ValidateCode: validateCode,
		IsActive:     true,
		JoinedAt:     time.Now(),
		Index:        index,
	}

	err = c.Insert(newUser)
	if err != nil {
		return map[string]interface{}{
			"status":   0,
			"messages": err.Error(),
		}
	}

	// 从 http://identicon.relucks.org 下载头像作为用户的默认头像
	go func() {
		res, err := http.Get(fmt.Sprintf("http://identicon.relucks.org/%s?size=400", form.Username))
		if err == nil {
			defer res.Body.Close()
			saveImage(res.Body, "image/png", form.Username+".png", []string{"avatar"}, -1)
		}

	}()

	c2.Update(nil, bson.M{"$inc": bson.M{"userindex": 1, "usercount": 1}})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": newUser.Id_.Hex(),
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
			"user": string(utils.Base64Encode([]byte(newUser.Id_.Hex()))),
		},
	}
}

// GetRandomString 生成随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// ForgotPassword 忘记密码
type ForgotPassword struct {
	Base
	binding.Binder
}

// Post /api/forgot_password
func (a *ForgotPassword) Post() interface{} {
	var form struct {
		UsernameOrEmail  string `json:"username_or_email" valid:"required,ascii"`
		GeetestChallenge string `json:"geetest_challenge" valid:"required,ascii"`
		GeetestValidate  string `json:"geetest_validate" valid:"required,ascii"`
		GeetestCeccode   string `json:"geetest_seccode" valid:"required,ascii"`
	}

	err := a.ReadJSON(&form)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	result, err := govalidator.ValidateStruct(form)

	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	p := url.Values{
		"client": {"web"},
	}
	g := geetest.New(conf.Config.GtCaptchaId, conf.Config.GtPrivateKey)
	success := g.SuccessValidate(form.GeetestChallenge, form.GeetestValidate, form.GeetestCeccode, p)

	if success == 0 {
		return map[string]interface{}{
			"status":  0,
			"message": "验证码错误",
		}
	}

	var user models.User
	c := a.DB.C(models.USERS)
	err = c.Find(bson.M{"$or": []bson.M{
		bson.M{"username": bson.M{"$regex": form.UsernameOrEmail, "$options": "i"}},
		bson.M{"email": bson.M{"$regex": form.UsernameOrEmail, "$options": "i"}}}}).One(&user)

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "没有该用户",
		}
	}

	subject := `Hi %s,<br>
我们的系统收到一个请求，说你希望通过电子邮件重新设置你在 Golang中国 的密码。你可以点击下面的链接开始重设密码：

<a href="%s/reset/%s">%s/reset/%s</a><br>

如果这个请求不是由你发起的，那没问题，你不用担心，你可以安全地忽略这封邮件。

如果你有任何疑问，可以回复 <a href="mailto:support@golangtc.com">support@golangtc.com</a> 向我提问。`
	var code string
	for {
		code = GetRandomString(32)
		// 不过有重复
		if n, _ := c.Find(bson.M{"resetcode": code}).Count(); n == 0 {
			break
		}
	}

	c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"resetcode": code}})
	subject = fmt.Sprintf(subject, user.Username, conf.Config.Host, code, conf.Config.Host, code)
	if conf.Config.SendMailPath == "" {
		webhelpers.SendMail(
			"[Golang 中国]重设密码",
			subject,
			conf.Config.FromEmail,
			[]string{user.Email},
			webhelpers.SmtpConfig{
				Username: conf.Config.SmtpUsername,
				Password: conf.Config.SmtpPassword,
				Host:     conf.Config.SmtpHost,
				Addr:     conf.Config.SmtpAddr,
			},
			true,
		)
	} else {
		webhelpers.SendMailExec(
			"[Golang 中国]重设密码",
			subject,
			conf.Config.FromEmail,
			[]string{user.Email},
			conf.Config.SendMailPath,
			true,
		)
	}

	return map[string]interface{}{
		"status":  1,
		"message": "一封包含了重设密码指令的邮件已经发送到你的注册邮箱，按照邮件中的提示，即可重设你的密码。",
	}
}

// ResetPassword 重设密码
type ResetPassword struct {
	Base
	binding.Binder
}

// Post /api/reset_password
func (a *ResetPassword) Post() interface{} {
	var form struct {
		Code            string `json:"code"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	err := a.ReadJSON(&form)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	result, err := govalidator.ValidateStruct(form)

	if !result {
		return map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		}
	}

	var user models.User
	c := a.DB.C(models.USERS)
	err = c.Find(bson.M{"resetcode": form.Code}).One(&user)

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "无效的重设密码标记，可能你已经重新设置过了或者链接已经失效。",
		}
	}

	if form.NewPassword != form.ConfirmPassword {
		return map[string]interface{}{
			"status":  0,
			"message": "新密码和确认密码不一致",
		}
	}

	salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
	c.Update(
		bson.M{"_id": user.Id_},
		bson.M{
			"$set": bson.M{
				"password":  utils.EncryptPassword(form.NewPassword, salt, models.PublicSalt),
				"salt":      salt,
				"resetcode": "",
			},
		},
	)

	return map[string]interface{}{
		"status":  1,
		"message": "重设密码成功",
	}
}

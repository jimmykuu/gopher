/*
处理用户相关的操作,注册,登录,验证,等等
*/
package gopher

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/bradrydzewski/go.auth"
	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jimmykuu/webhelpers"
	"github.com/jimmykuu/wtforms"
	. "github.com/qiniu/api/conf"
	qiniuIo "github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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

const (
	GITHUB_PICTURE  = "github_picture"
	GITHUB_ID       = "github_id"
	GITHUB_LINK     = "github_link"
	GITHUB_EMAIL    = "github_email"
	GITHUB_NAME     = "github_name"
	GITHUB_ORG      = "github_org"
	GITHUB_PROVIDER = "github_provider"
)

//init 在config.go下,因为这个文件的init调用比config慢
var githubHandler *auth.AuthHandler

// 生成users.json字符串
func generateUsersJson(db *mgo.Database) {
	var users []User
	c := db.C(USERS)
	err := c.Find(nil).All(&users)
	if err != nil {
		panic(err)
	}
	var usernames []string
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}

	b, err := json.Marshal(usernames)
	if err != nil {
		panic(err)
	}
	usersJson = b
}

// 加密密码，md5(md5(password + salt) + public_salt)
func encryptPassword(password, salt string) string {
	saltedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password))) + salt
	md5SaltedPassword := fmt.Sprintf("%x", md5.Sum([]byte(saltedPassword)))
	return fmt.Sprintf("%x", md5.Sum([]byte(md5SaltedPassword+Config.PublicSalt)))
}

// 返回当前用户
func currentUser(handler *Handler) (*User, bool) {
	r := handler.Request
	session, _ := store.Get(r, "user")
	username, ok := session.Values["username"]

	if !ok {
		return nil, false
	}

	username = username.(string)

	user := User{}

	c := handler.DB.C(USERS)

	// 检查用户名
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		return nil, false
	}

	return &user, true
}

// URL: /auth/login
func authLoginHandler(handler *Handler) {
	githubHandler.ServeHTTP(handler.ResponseWriter, handler.Request)
}

// wrapAuthHandler返回符合 go.auth包要求签名的函数.
func wrapAuthHandler(handler *Handler) func(w http.ResponseWriter, r *http.Request, u auth.User) {
	return func(w http.ResponseWriter, r *http.Request, u auth.User) {
		c := handler.DB.C(USERS)
		user := User{}
		session, _ := store.Get(r, "user")
		c.Find(bson.M{"username": u.Id()}).One(&user)
		//关联github帐号,直接登录
		if user.Provider == GITHUB_COM {
			session.Values["username"] = user.Username
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		form := wtforms.NewForm(wtforms.NewTextField("username", "用户名", "", wtforms.Required{}),
			wtforms.NewPasswordField("password", "密码", wtforms.Required{}))
		session.Values[GITHUB_EMAIL] = u.Email()
		session.Values[GITHUB_ID] = u.Id()
		session.Values[GITHUB_LINK] = u.Link()
		session.Values[GITHUB_NAME] = u.Name()
		session.Values[GITHUB_ORG] = u.Org()
		session.Values[GITHUB_PICTURE] = u.Picture()
		session.Values[GITHUB_PROVIDER] = u.Provider()
		session.Save(r, w)

		//关联已有帐号
		if handler.Request.Method == "POST" {
			if form.Validate(handler.Request) {
				user := User{}
				err := c.Find(bson.M{"username": form.Value("username")}).One(&user)
				if err != nil {
					form.AddError("username", "该用户不存在")
					handler.renderTemplate("accoun/auth_login.html", BASE, map[string]interface{}{"form": form})
					return
				}
				if user.Password != encryptPassword(form.Value("password"), user.Salt) {
					form.AddError("password", "密码和用户名不匹配")

					handler.renderTemplate("account/auth_login.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
					return
				}
				c.UpdateId(user.Id_, bson.M{"$set": bson.M{
					"emal":       session.Values[GITHUB_EMAIL],
					"accountref": session.Values[GITHUB_NAME],
					"username":   session.Values[GITHUB_ID],
					"idref":      session.Values[GITHUB_ID],
					"linkref":    session.Values[GITHUB_LINK],
					"orgref":     session.Values[GITHUB_ORG],
					"pictureref": session.Values[GITHUB_PICTURE],
					"provider":   session.Values[GITHUB_PROVIDER],
				}})
				deleteGithubValues(session)
				session.Values["username"] = u.Name()
				session.Save(r, w)
				http.Redirect(handler.ResponseWriter, handler.Request, "/", http.StatusFound)
			}
		}
		handler.renderTemplate("account/auth_login.html", BASE, map[string]interface{}{"form": form})
	}
}

// URL: /auth/signup
func authSignupHandler(handler *Handler) {
	fn := auth.SecureUser(wrapAuthHandler(handler))
	fn.ServeHTTP(handler.ResponseWriter, handler.Request)
}

// URL: /signup
// 处理用户注册,要求输入用户名,密码和邮箱
func signupHandler(handler *Handler) {
	// 如果已经登录了，跳转到首页
	_, has := currentUser(handler)
	if has {
		handler.Redirect("/")
	}

	var username string
	var email string
	session, _ := store.Get(handler.Request, "user")
	if handler.Request.Method == "GET" {
		//如果是从新建关联过来的就自动填充字段
		if session.Values[GITHUB_PROVIDER] == GITHUB_COM {
			username = session.Values[GITHUB_ID].(string)
			email = session.Values[GITHUB_EMAIL].(string)
		}
	}
	form := wtforms.NewForm(
		wtforms.NewTextField("username", "用户名", username, wtforms.Required{}, wtforms.Regexp{Expr: `^[a-zA-Z0-9_]{3,16}$`, Message: "请使用a-z, A-Z, 0-9以及下划线, 长度3-16之间"}),
		wtforms.NewPasswordField("password", "密码", wtforms.Required{}),
		wtforms.NewTextField("email", "电子邮件", email, wtforms.Required{}, wtforms.Email{}),
		wtforms.NewTextField("captcha", "验证码", "", wtforms.Required{}),
		wtforms.NewHiddenField("captchaId", ""),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			// 检查验证码
			if !captcha.VerifyString(form.Value("captchaId"), form.Value("captcha")) {
				form.AddError("captcha", "验证码错误")
				fmt.Println("captcha")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signup.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			c := handler.DB.C(USERS)

			result := User{}

			// 检查用户名
			err := c.Find(bson.M{"username": form.Value("username")}).One(&result)
			if err == nil {
				form.AddError("username", "该用户名已经被注册")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signup.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			// 检查邮箱
			err = c.Find(bson.M{"email": form.Value("email")}).One(&result)

			if err == nil {
				form.AddError("email", "电子邮件地址已经被注册")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signup.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			c2 := handler.DB.C(STATUS)
			var status Status
			c2.Find(nil).One(&status)

			id := bson.NewObjectId()
			username := form.Value("username")
			validateCode := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
			salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
			index := status.UserIndex + 1
			u := &User{
				Id_:          id,
				Username:     username,
				Password:     encryptPassword(form.Value("password"), salt),
				Avatar:       "", // defaultAvatars[rand.Intn(len(defaultAvatars))],
				Salt:         salt,
				Email:        form.Value("email"),
				ValidateCode: validateCode,
				IsActive:     true,
				JoinedAt:     time.Now(),
				Index:        index,
			}
			if session.Values[GITHUB_PROVIDER] == GITHUB_COM {
				u.GetGithubValues(session)
				defer deleteGithubValues(session)
			}
			err = c.Insert(u)
			if err != nil {
				logger.Println(err)
				return
			}

			c2.Update(nil, bson.M{"$inc": bson.M{"userindex": 1, "usercount": 1}})

			// 重新生成users.json字符串
			generateUsersJson(handler.DB)

			// 发送邮件
			/*
							subject := "欢迎加入Golang 中国"
							message2 := `欢迎加入Golang 中国。请访问下面地址激活你的帐户。

				<a href="%s/activate/%s">%s/activate/%s</a>

				如果你没有注册，请忽略这封邮件。

				©2012 Golang 中国`
							message2 = fmt.Sprintf(message2, config["host"], validateCode, config["host"], validateCode)
							sendMail(subject, message2, []string{form.Value("email")})

							message(w, r, "注册成功", "请查看你的邮箱进行验证，如果收件箱没有，请查看垃圾邮件，如果还没有，请给jimmykuu@126.com发邮件，告知你的用户名。", "success")
			*/
			// 注册成功后设成登录状态
			session, _ := store.Get(handler.Request, "user")
			session.Values["username"] = username
			session.Save(handler.Request, handler.ResponseWriter)

			// 跳到修改用户信息页面
			http.Redirect(handler.ResponseWriter, handler.Request, "/profile", http.StatusFound)
			return
		}
	}
	form.SetValue("captcha", "")
	handler.renderTemplate("account/signup.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
}

// URL: /activate/{code}
// 用户根据邮件中的链接进行验证,根据code找到是否有对应的用户,如果有,修改User.IsActive为true
func activateHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	code := vars["code"]

	var user User

	c := handler.DB.C(USERS)

	err := c.Find(bson.M{"validatecode": code}).One(&user)

	if err != nil {
		message(handler, "没有该验证码", "请检查连接是否正确", "error")
		return
	}

	c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"isactive": true, "validatecode": ""}})

	c = handler.DB.C(STATUS)
	var status Status
	c.Find(nil).One(&status)
	c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"usercount": 1}})

	message(handler, "通过验证", `恭喜你通过验证,请 <a href="/signin">登录</a>.`, "success")
}

// URL: /signin
// 处理用户登录,如果登录成功,设置Cookie
func signinHandler(handler *Handler) {
	// 如果已经登录了，跳转到首页
	_, has := currentUser(handler)
	if has {
		handler.Redirect("/")
	}

	next := handler.Request.FormValue("next")

	form := wtforms.NewForm(
		wtforms.NewHiddenField("next", next),
		wtforms.NewTextField("username", "用户名", "", &wtforms.Required{}),
		wtforms.NewPasswordField("password", "密码", &wtforms.Required{}),
		wtforms.NewTextField("captcha", "验证码", "", wtforms.Required{}),
		wtforms.NewHiddenField("captchaId", ""),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			// 检查验证码
			if !captcha.VerifyString(form.Value("captchaId"), form.Value("captcha")) {
				form.AddError("captcha", "验证码错误")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signin.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			c := handler.DB.C(USERS)
			user := User{}

			err := c.Find(bson.M{"username": form.Value("username")}).One(&user)

			if err != nil {
				form.AddError("username", "该用户不存在")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signin.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			if !user.IsActive {
				form.AddError("username", "邮箱没有经过验证,如果没有收到邮件,请联系管理员")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signin.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			if user.Password != encryptPassword(form.Value("password"), user.Salt) {
				form.AddError("password", "密码和用户名不匹配")
				form.SetValue("captcha", "")

				handler.renderTemplate("account/signin.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
				return
			}

			session, _ := store.Get(handler.Request, "user")
			session.Values["username"] = user.Username
			session.Save(handler.Request, handler.ResponseWriter)

			if form.Value("next") == "" {
				http.Redirect(handler.ResponseWriter, handler.Request, "/", http.StatusFound)
			} else {
				http.Redirect(handler.ResponseWriter, handler.Request, next, http.StatusFound)
			}

			return
		}
	}

	form.SetValue("captcha", "")
	handler.renderTemplate("account/signin.html", BASE, map[string]interface{}{"form": form, "captchaId": captcha.New()})
}

// URL: /signout
// 用户登出,清除Cookie
func signoutHandler(handler *Handler) {
	session, _ := store.Get(handler.Request, "user")
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(handler.Request, handler.ResponseWriter)
	handler.renderTemplate("account/signout.html", BASE, map[string]interface{}{"signout": true})
}

func followHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	username := vars["username"]

	currUser, _ := currentUser(handler)

	//不能关注自己
	if currUser.Username == username {
		message(handler, "提示", "不能关注自己", "error")
		return
	}

	user := User{}
	c := handler.DB.C(USERS)
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(handler, "关注的会员未找到", "关注的会员未找到", "error")
		return
	}

	if user.IsFollowedBy(currUser.Username) {
		message(handler, "你已经关注该会员", "你已经关注该会员", "error")
		return
	}
	c.Update(bson.M{"_id": user.Id_}, bson.M{"$push": bson.M{"fans": currUser.Username}})
	c.Update(bson.M{"_id": currUser.Id_}, bson.M{"$push": bson.M{"follow": user.Username}})

	http.Redirect(handler.ResponseWriter, handler.Request, "/member/"+user.Username, http.StatusFound)
}

func unfollowHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	username := vars["username"]

	currUser, _ := currentUser(handler)

	//不能取消关注自己
	if currUser.Username == username {
		message(handler, "提示", "不能对自己进行操作", "error")
		return
	}

	user := User{}
	c := handler.DB.C(USERS)
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(handler, "没有该会员", "没有该会员", "error")
		return
	}

	if !user.IsFollowedBy(currUser.Username) {
		message(handler, "不能取消关注", "该会员不是你的粉丝,不能取消关注", "error")
		return
	}

	c.Update(bson.M{"_id": user.Id_}, bson.M{"$pull": bson.M{"fans": currUser.Username}})
	c.Update(bson.M{"_id": currUser.Id_}, bson.M{"$pull": bson.M{"follow": user.Username}})

	http.Redirect(handler.ResponseWriter, handler.Request, "/member/"+user.Username, http.StatusFound)
}

// URL /profile
// 用户设置页面,显示用户设置,用户头像,密码修改
func profileHandler(handler *Handler) {
	user, _ := currentUser(handler)

	profileForm := wtforms.NewForm(
		wtforms.NewTextField("email", "电子邮件", user.Email, wtforms.Email{}),
		wtforms.NewTextField("website", "个人网站", user.Website),
		wtforms.NewTextField("location", "所在地", user.Location),
		wtforms.NewTextField("tagline", "签名", user.Tagline),
		wtforms.NewTextArea("bio", "个人简介", user.Bio),
		wtforms.NewTextField("github_username", "GitHub用户名", user.GitHubUsername),
		wtforms.NewTextField("weibo", "新浪微博", user.Weibo),
	)

	if handler.Request.Method == "POST" {
		if profileForm.Validate(handler.Request) {
			c := handler.DB.C(USERS)

			// 检查邮箱
			result := new(User)
			err := c.Find(bson.M{"email": profileForm.Value("email")}).One(result)
			if err == nil && result.Id_ != user.Id_ {
				profileForm.AddError("email", "电子邮件地址已经被使用")

				handler.renderTemplate("account/profile.html", BASE, map[string]interface{}{
					"user":           user,
					"profileForm":    profileForm,
					"defaultAvatars": defaultAvatars,
				})
				return
			}

			c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{
				"email":          profileForm.Value("email"),
				"website":        profileForm.Value("website"),
				"location":       profileForm.Value("location"),
				"tagline":        profileForm.Value("tagline"),
				"bio":            profileForm.Value("bio"),
				"githubusername": profileForm.Value("github_username"),
				"weibo":          profileForm.Value("weibo"),
			}})
			http.Redirect(handler.ResponseWriter, handler.Request, "/profile", http.StatusFound)
			return
		}
	}

	handler.renderTemplate("account/profile.html", BASE, map[string]interface{}{
		"user":           user,
		"profileForm":    profileForm,
		"defaultAvatars": defaultAvatars,
	})
}

// URL: /forgot_password
// 忘记密码,输入用户名和邮箱,如果匹配,发出邮件
func forgotPasswordHandler(handler *Handler) {
	form := wtforms.NewForm(
		wtforms.NewTextField("username", "用户名", "", wtforms.Required{}),
		wtforms.NewTextField("email", "电子邮件", "", wtforms.Email{}),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			var user User
			c := handler.DB.C(USERS)
			err := c.Find(bson.M{"username": form.Value("username")}).One(&user)
			if err != nil {
				form.AddError("username", "没有该用户")
			} else if user.Email != form.Value("email") {
				form.AddError("username", "用户名和邮件不匹配")
			} else {
				message2 := `Hi %s,<br>
我们的系统收到一个请求，说你希望通过电子邮件重新设置你在 Golang中国 的密码。你可以点击下面的链接开始重设密码：

<a href="%s/reset/%s">%s/reset/%s</a><br>

如果这个请求不是由你发起的，那没问题，你不用担心，你可以安全地忽略这封邮件。

如果你有任何疑问，可以回复<a href="mailto:support@golangtc.com">support@golangtc.com</a>向我提问。`
				code := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
				c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"resetcode": code}})
				message2 = fmt.Sprintf(message2, user.Username, Config.Host, code, Config.Host, code)
				webhelpers.SendMail(
					"[Golang中国]重设密码",
					message2,
					Config.FromEmail,
					[]string{user.Email},
					webhelpers.SmtpConfig{
						Username: Config.SmtpUsername,
						Password: Config.SmtpPassword,
						Host:     Config.SmtpHost,
						Addr:     Config.SmtpAddr,
					},
					true,
				)
				message(handler, "通过电子邮件重设密码", "一封包含了重设密码指令的邮件已经发送到你的注册邮箱，按照邮件中的提示，即可重设你的密码。", "success")
				return
			}
		}
	}

	handler.renderTemplate("account/forgot_password.html", BASE, map[string]interface{}{"form": form})
}

// URL: /reset/{code}
// 用户点击邮件中的链接,根据code找到对应的用户,设置新密码,修改完成后清除code
func resetPasswordHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	code := vars["code"]

	var user User
	c := handler.DB.C(USERS)
	err := c.Find(bson.M{"resetcode": code}).One(&user)

	if err != nil {
		message(handler, "重设密码", `无效的重设密码标记,可能你已经重新设置过了或者链接已经失效,请通过<a href="/forgot_password">忘记密码</a>进行重设密码`, "error")
		return
	}

	form := wtforms.NewForm(
		wtforms.NewPasswordField("new_password", "新密码", wtforms.Required{}),
		wtforms.NewPasswordField("confirm_password", "确认新密码", wtforms.Required{}),
	)

	if handler.Request.Method == "POST" && form.Validate(handler.Request) {
		if form.Value("new_password") == form.Value("confirm_password") {
			salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
			c.Update(
				bson.M{"_id": user.Id_},
				bson.M{
					"$set": bson.M{
						"password":  encryptPassword(form.Value("new_password"), salt),
						"salt":      salt,
						"resetcode": "",
					},
				},
			)
			message(handler, "重设密码成功", `密码重设成功,你现在可以 <a href="/signin" class="btn btn-primary">登录</a> 了`, "success")
			return
		} else {
			form.AddError("confirm_password", "密码不匹配")
		}
	}

	handler.renderTemplate("account/reset_password.html", BASE, map[string]interface{}{"form": form, "code": code, "account": user.Username})
}

type Sizer interface {
	Size() int64
}

// 上传到七牛，并返回文件名
func uploadAvatarToQiniu(file io.ReadCloser, contentType string) (filename string, err error) {
	isValidateType := false
	for _, imgType := range []string{"image/png", "image/jpeg"} {
		if imgType == contentType {
			isValidateType = true
			break
		}
	}

	if !isValidateType {
		return "", errors.New("文件类型错误")
	}

	filenameExtension := ".jpg"
	if contentType == "image/png" {
		filenameExtension = ".png"
	}

	// 文件名：32位uuid，不带减号和后缀组成
	filename = strings.Replace(uuid.NewUUID().String(), "-", "", -1) + filenameExtension

	ACCESS_KEY = Config.QiniuAccessKey
	SECRET_KEY = Config.QiniuSecretKey

	key := "avatar/" + filename

	ret := new(qiniuIo.PutRet)

	var policy = rs.PutPolicy{
		Scope: "gopher",
	}

	err = qiniuIo.Put(
		nil,
		ret,
		policy.Token(nil),
		key,
		file,
		nil,
	)

	if err != nil {
		return "", err
	}

	return filename, nil
}

// URL: /profile/avatar
// 修改头像,提交到七牛云存储
func changeAvatarHandler(handler *Handler) {
	user, _ := currentUser(handler)

	if handler.Request.Method == "POST" {
		formFile, formHeader, err := handler.Request.FormFile("file")
		if err != nil {
			fmt.Println("changeAvatarHandler:", err.Error())
			handler.renderTemplate("account/avatar.html", BASE, map[string]interface{}{
				"user":  user,
				"error": "请选择图片上传",
			})
			return
		}
		// 检查文件尺寸是否在500K以内
		fileSize := formFile.(Sizer).Size()

		if fileSize > 500*1024 {
			// > 500K
			fmt.Printf("upload image size > 500K: %dK\n", fileSize/1024)
			handler.renderTemplate("account/avatar.html", BASE, map[string]interface{}{
				"user":  user,
				"error": "图片大小大于500K，请选择500K以内图片上传。",
			})
			return
		}

		defer formFile.Close()

		// 检查是否是jpg或png文件
		uploadFileType := formHeader.Header["Content-Type"][0]

		filename, err := uploadAvatarToQiniu(formFile, uploadFileType)

		// 存储远程文件名
		c := handler.DB.C(USERS)
		c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": filename}})

		http.Redirect(handler.ResponseWriter, handler.Request, "/profile#avatar", http.StatusFound)
		return
	}

	handler.renderTemplate("account/avatar.html", BASE, map[string]interface{}{"user": user})
}

// URL: /profile/choose_default_avatar
// 选择默认头像
func chooseDefaultAvatar(handler *Handler) {
	user, _ := currentUser(handler)

	if handler.Request.Method == "POST" {
		avatar := handler.Request.FormValue("defaultAvatars")

		if avatar != "" {
			c := handler.DB.C(USERS)
			c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": avatar}})
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/profile#avatar", http.StatusFound)
	}
}

// URL: /change_password
// 修改密码
func changePasswordHandler(handler *Handler) {
	user, _ := currentUser(handler)

	form := wtforms.NewForm(
		wtforms.NewPasswordField("current_password", "当前密码", wtforms.Required{}),
		wtforms.NewPasswordField("new_password", "新密码", wtforms.Required{}),
		wtforms.NewPasswordField("confirm_password", "新密码确认", wtforms.Required{}),
	)

	if handler.Request.Method == "POST" && form.Validate(handler.Request) {
		if form.Value("new_password") == form.Value("confirm_password") {
			currentPassword := encryptPassword(form.Value("current_password"), user.Password)
			if currentPassword == user.Password {
				c := handler.DB.C(USERS)
				salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)
				c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{
					"password": encryptPassword(form.Value("new_password"), salt),
					"salt":     salt,
				}})
				message(handler, "密码修改成功", `密码修改成功`, "success")
				return
			} else {
				form.AddError("current_password", "当前密码错误")
			}
		} else {
			form.AddError("confirm_password", "密码不匹配")
		}
	}

	handler.renderTemplate("account/change_password.html", BASE, map[string]interface{}{"form": form})
}

//  URL: /users.json
// 获取所有用户的json列表
func usersJsonHandler(handler *Handler) {
	handler.ResponseWriter.Write(usersJson)
}

// URl: /profile/avatar/gravatar
// 从Gravatar获取头像
func setAvatarFromGravatar(handler *Handler) {
	user, _ := currentUser(handler)
	url := webhelpers.Gravatar(user.Email, 256)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	filename, err := uploadAvatarToQiniu(resp.Body, resp.Header["Content-Type"][0])
	if err != nil {
		panic(err)
	}

	c := handler.DB.C(USERS)
	c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": filename}})

	http.Redirect(handler.ResponseWriter, handler.Request, "/profile#avatar", http.StatusFound)
}

/*
处理用户相关的操作,注册,登录,验证,等等
*/
package main

import (
	"./wtforms"
	"crypto/md5"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"time"
)

// 加密密码,转成md5
func encryptPassword(password string) string {
	h := md5.New()
	io.WriteString(h, password)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 返回当前用户
func currentUser(r *http.Request) (*User, bool) {
	session, _ := store.Get(r, "user")
	username, ok := session.Values["username"]

	if !ok {
		return nil, false
	}

	username = username.(string)

	user := User{}

	c := db.C("users")

	// 检查用户名
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		return nil, false
	}

	return &user, true
}

// URL: /signup
// 处理用户注册,要求输入用户名,密码和邮箱
func signupHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("username", "用户名", "", wtforms.Required{}, wtforms.Regexp{Expr: `^[a-zA-Z0-9_]{3,16}$`, Message: "请使用a-z, A-Z, 0-9以及下划线, 长度3-16之间"}),
		wtforms.NewPasswordField("password", "密码", wtforms.Required{}),
		wtforms.NewTextField("email", "电子邮件", "", wtforms.Required{}, wtforms.Email{}),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			c := db.C("users")

			result := User{}

			// 检查用户名
			err := c.Find(bson.M{"username": form.Value("username")}).One(&result)
			if err == nil {
				form.AddError("username", "该用户名已经被注册")

				renderTemplate(w, r, "account/signup.html", map[string]interface{}{"form": form})
				return
			}

			// 检查邮箱
			err = c.Find(bson.M{"email": form.Value("email")}).One(&result)

			if err == nil {
				form.AddError("email", "电子邮件地址已经被注册")

				renderTemplate(w, r, "account/signup.html", map[string]interface{}{"form": form})
				return
			}

			c2 := db.C("status")
			var status Status
			c2.Find(nil).One(&status)

			id := bson.NewObjectId()

			validateCode := uuid()
			index := status.UserIndex + 1
			err = c.Insert(&User{
				Id_:          id,
				Username:     form.Value("username"),
				Password:     encryptPassword(form.Value("password")),
				Email:        form.Value("email"),
				ValidateCode: validateCode,
				IsActive:     true,
				JoinedAt:     time.Now(),
				Index:        index,
			})

			if err != nil {
				panic(err)
			}

			c2.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"userindex": 1}})

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
			message(w, r, "注册成功", fmt.Sprintf(`感谢您的注册，您已经成为Golang中国第 <strong>%d</strong>位用户，<a href="/signin">登录</a>。`, index), "success")
			return
		}
	}

	renderTemplate(w, r, "account/signup.html", map[string]interface{}{"form": form})
}

// URL: "/activate/{code}"
// 用户根据邮件中的链接进行验证,根据code找到是否有对应的用户,如果有,修改User.IsActive为true
func activateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	var user User

	c := db.C("users")

	err := c.Find(bson.M{"validatecode": code}).One(&user)

	if err != nil {
		message(w, r, "没有该验证码", "请检查连接是否正确", "error")
		return
	}

	c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"isactive": true, "validatecode": ""}})

	c = db.C("status")
	var status Status
	c.Find(nil).One(&status)
	c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"usercount": 1}})

	message(w, r, "通过验证", `恭喜你通过验证,请 <a href="/signin">登录</a>.`, "success")
}

// URL: /signin
// 处理用户登录,如果登录成功,设置Cookie
func signinHandler(w http.ResponseWriter, r *http.Request) {
	next := r.FormValue("next")

	form := wtforms.NewForm(
		wtforms.NewHiddenField("next", next),
		wtforms.NewTextField("username", "用户名", "", &wtforms.Required{}),
		wtforms.NewPasswordField("password", "密码", &wtforms.Required{}),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			c := db.C("users")
			user := User{}

			err := c.Find(bson.M{"username": form.Value("username")}).One(&user)

			if err != nil {
				form.AddError("username", "该用户不存在")

				renderTemplate(w, r, "account/signin.html", map[string]interface{}{"form": form})
				return
			}

			if !user.IsActive {
				form.AddError("username", "邮箱没有经过验证,如果没有收到邮件,请联系管理员")
				renderTemplate(w, r, "account/signin.html", map[string]interface{}{"form": form})
				return
			}

			if user.Password != encryptPassword(form.Value("password")) {
				form.AddError("password", "密码和用户名不匹配")

				renderTemplate(w, r, "account/signin.html", map[string]interface{}{"form": form})
				return
			}

			session, _ := store.Get(r, "user")
			session.Values["username"] = user.Username
			session.Save(r, w)

			if form.Value("next") == "" {
				http.Redirect(w, r, "/", http.StatusFound)
			} else {
				http.Redirect(w, r, next, http.StatusFound)
			}

			return
		}
	}

	renderTemplate(w, r, "account/signin.html", map[string]interface{}{"form": form})
}

// URL: /signout
// 用户登出,清除Cookie
func signoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "user")
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)
	renderTemplate(w, r, "account/signout.html", map[string]interface{}{"signout": true})
}

// URL: /member/{username}
// 显示用户信息
func memberInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	c := db.C("users")

	user := User{}

	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "会员未找到", "会员未找到", "error")
		return
	}

	renderTemplate(w, r, "account/info.html", map[string]interface{}{"user": user})
}

// URL: /member/{username}/topics
// 用户发表的所有主题
func memberTopicsHandler(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("p")
	page := 1

	if p != "" {
		var err error
		page, err = strconv.Atoi(p)

		if err != nil {
			message(w, r, "没有找到页面", "没有找到页面", "error")
			return
		}
	}

	vars := mux.Vars(r)
	username := vars["username"]
	c := db.C("users")

	user := User{}
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "会员未找到", "会员未找到", "error")
		return
	}

	c = db.C("topics")

	pagination := NewPagination(c.Find(bson.M{"userid": user.Id_}).Sort("-latestrepliedat"), "/member/"+username+"/topics", PerPage)

	var topics []Topic

	query, err := pagination.Page(page)

	if err != nil {
		message(w, r, "没有找到页面", "没有找到页面", "error")
		return
	}

	query.All(&topics)

	renderTemplate(w, r, "account/topics.html", map[string]interface{}{"user": user, "topics": topics, "pagination": pagination, "page": page})
}

// /member/{username}/replies
// 用户的所有回复
func memberRepliesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	c := db.C("users")

	user := User{}
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "会员未找到", "会员未找到", "error")
		return
	}

	var replies []Reply
	c = db.C("replies")
	c.Find(bson.M{"userid": user.Id_}).Sort("-createdat").All(&replies)

	renderTemplate(w, r, "account/replies.html", map[string]interface{}{"user": user, "replies": replies})
}

func followHandler(w http.ResponseWriter, r *http.Request) {
	println("follow")
	vars := mux.Vars(r)
	username := vars["username"]

	// 检查当前用户是否存在
	currUser, ok := currentUser(r)

	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	//不能关注自己
	if currUser.Username == username {
		message(w, r, "提示", "不能关注自己", "error")
		return
	}

	user := User{}
	c := db.C("users")
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "关注的会员未找到", "关注的会员未找到", "error")
		return
	}

	if user.IsFollowedBy(currUser.Username) {
		message(w, r, "你已经关注该会员", "你已经关注该会员", "error")
		return
	}
	c.Update(bson.M{"_id": user.Id_}, bson.M{"$push": bson.M{"fans": currUser.Username}})
	c.Update(bson.M{"_id": currUser.Id_}, bson.M{"$push": bson.M{"follow": user.Username}})

	http.Redirect(w, r, "/member/"+user.Username, http.StatusFound)
}

func unfollowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	// 检查当前用户是否存在
	currUser, ok := currentUser(r)

	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	//不能取消关注自己
	if currUser.Username == username {
		message(w, r, "提示", "不能对自己进行操作", "error")
		return
	}

	user := User{}
	c := db.C("users")
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "没有该会员", "没有该会员", "error")
		return
	}

	if !user.IsFollowedBy(currUser.Username) {
		message(w, r, "不能取消关注", "该会员不是你的粉丝,不能取消关注", "error")
		return
	}

	c.Update(bson.M{"_id": user.Id_}, bson.M{"$pull": bson.M{"fans": currUser.Username}})
	c.Update(bson.M{"_id": currUser.Id_}, bson.M{"$pull": bson.M{"follow": user.Username}})

	http.Redirect(w, r, "/member/"+user.Username, http.StatusFound)
}

// URL /profile
// 用户设置页面,显示用户设置,用户头像,密码修改
func profileHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)

	if !ok {
		http.Redirect(w, r, "/signin?next=/profile", http.StatusFound)
		return
	}

	profileForm := wtforms.NewForm(
		wtforms.NewTextField("email", "电子邮件", user.Email, wtforms.Email{}),
		wtforms.NewTextField("website", "个人网站", user.Website),
		wtforms.NewTextField("location", "所在地", user.Location),
		wtforms.NewTextField("tagline", "签名", user.Tagline),
		wtforms.NewTextArea("bio", "个人简介", user.Bio),
	)

	if r.Method == "POST" {
		if profileForm.Validate(r) {
			c := db.C("users")
			c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"website": profileForm.Value("website"),
				"location": profileForm.Value("location"),
				"tagline":  profileForm.Value("tagline"),
				"bio":      profileForm.Value("bio"),
			}})
			http.Redirect(w, r, "/profile", http.StatusFound)
			return
		}
	}

	changePasswordForm := wtforms.NewForm(
		wtforms.NewPasswordField("current_password", "当前密码"),
		wtforms.NewPasswordField("new_password", "新密码"),
		wtforms.NewPasswordField("confirm_password", "新密码确认"),
	)

	renderTemplate(w, r, "account/profile.html", map[string]interface{}{"user": user, "profileForm": profileForm, "changePasswordForm": changePasswordForm})
}

// URL: /forgot_password
// 忘记密码,输入用户名和邮箱,如果匹配,发出邮件
func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("username", "用户名", "", wtforms.Required{}),
		wtforms.NewTextField("email", "电子邮件", "", wtforms.Email{}),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			var user User
			c := db.C("users")
			err := c.Find(bson.M{"username": form.Value("username")}).One(&user)
			if err != nil {
				form.AddError("username", "没有该用户")
			} else if user.Email != form.Value("email") {
				form.AddError("username", "用户名和邮件不匹配")
			} else {
				message2 := `Hi %s,
我们的系统收到一个请求，说你希望通过电子邮件重新设置你在 Golang中国 的密码。你可以点击下面的链接开始重设密码：

%s/reset/%s

如果这个请求不是由你发起的，那没问题，你不用担心，你可以安全地忽略这封邮件。

如果你有任何疑问，可以回复这封邮件向我提问。`
				code := uuid()
				c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"resetcode": code}})
				message2 = fmt.Sprintf(message2, user.Username, config["host"], code)
				sendMail("[Golang中国]重设密码", message2, []string{user.Email})
				message(w, r, "通过电子邮件重设密码", "一封包含了重设密码指令的邮件已经发送到你的注册邮箱，按照邮件中的提示，即可重设你的密码。", "success")
				return
			}
		}
	}

	renderTemplate(w, r, "account/forgot_password.html", map[string]interface{}{"form": form})
}

// URL: /reset/{code}
// 用户点击邮件中的链接,根据code找到对应的用户,设置新密码,修改完成后清除code
func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	var user User
	c := db.C("users")
	err := c.Find(bson.M{"resetcode": code}).One(&user)

	if err != nil {
		message(w, r, "重设密码", `无效的重设密码标记,可能你已经重新设置过了或者链接已经失效,请通过<a href="/forgot_password">忘记密码</a>进行重设密码`, "error")
		return
	}

	form := wtforms.NewForm(
		wtforms.NewPasswordField("new_password", "新密码", wtforms.Required{}),
		wtforms.NewPasswordField("confirm_password", "确认新密码", wtforms.Required{}),
	)

	if r.Method == "POST" && form.Validate(r) {
		if form.Value("new_password") == form.Value("confirm_password") {
			c.Update(
				bson.M{"_id": user.Id_},
				bson.M{
					"$set": bson.M{
						"password":  encryptPassword(form.Value("new_password")),
						"resetcode": "",
					},
				},
			)
			message(w, r, "重设密码成功", `密码重设成功,你现在可以 <a href="/signin" class="btn btn-primary">登录</a> 了`, "success")
			return
		} else {
			form.AddError("confirm_password", "密码不匹配")
		}
	}

	renderTemplate(w, r, "account/reset_password.html", map[string]interface{}{"form": form, "code": code, "account": user.Username})
}

// URL: /profile/avatar
// 修改头像,头像使用又拍云存储,直接使用从页面提交到又拍云,然后回调
//func changeAvatarHandler(w http.ResponseWriter, r *http.Request) {
//	user, ok := getCurrentUser(r)

//	if !ok {
//		http.Redirect(w, r, "/signin?next=/profile/avatar", http.StatusFound)
//		return
//	}

//	policyStr := policy()
//	signStr := sign(policyStr)
//	renderTemplate(w, r, "account/avatar.html", map[string]interface{}{"user": user, "policy": policyStr, "sign": signStr})
//}

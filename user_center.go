package gopher

import (
	"fmt"
	"net/http"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/jimmykuu/webhelpers"
	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2/bson"
)

// URL: /user_center
// 用户中心首页
func userCenterHandler(handler *Handler) {
	handler.renderTemplate("user_center/dashboard.html", BASE, map[string]interface{}{
		"active": "dashboard",
	})
}

// URL: /user_center/change_avatar
// 修改头像,提交到七牛云存储
func changeAvatarHandler(handler *Handler) {
	handler.renderTemplate("user_center/change_avatar.html", BASE, map[string]interface{}{
		"active":         "change_avatar",
		"defaultAvatars": defaultAvatars,
	})
}

// URL: /user_center/upload_avatar
// 上传头像
func uploadAvatarHandler(handler *Handler) {
	if handler.Request.Method != "POST" {
		return
	}

	formFile, formHeader, err := handler.Request.FormFile("file")
	if err != nil {
		fmt.Println("changeAvatarHandler:", err.Error())

		handler.renderTemplate("user_center/change_avatar.html", BASE, map[string]interface{}{
			"active":         "change_avatar",
			"defaultAvatars": defaultAvatars,
			"error":          "请选择图片上传",
		})

		return
	}
	// 检查文件尺寸是否在500K以内
	fileSize := formFile.(Sizer).Size()

	if fileSize > 500*1024 {
		// > 500K
		fmt.Printf("upload image size > 500K: %dK\n", fileSize/1024)
		handler.renderTemplate("user_center/change_avatar.html", BASE, map[string]interface{}{
			"active":         "change_avatar",
			"defaultAvatars": defaultAvatars,
			"error":          "图片大小大于500K，请选择500K以内图片上传。",
		})
		return
	}

	defer formFile.Close()

	// 检查是否是jpg或png文件
	uploadFileType := formHeader.Header["Content-Type"][0]

	filename, err := uploadAvatarToQiniu(formFile, uploadFileType)

	if err != nil {
		fmt.Println(err)
		handler.renderTemplate("user_center/change_avatar.html", BASE, map[string]interface{}{
			"active":         "change_avatar",
			"defaultAvatars": defaultAvatars,
			"error":          "图片上传失败",
		})
		return
	}

	user, _ := currentUser(handler)
	// 存储远程文件名
	c := handler.DB.C(USERS)
	c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": filename}})

	handler.redirect("/user_center", http.StatusFound)
}

// URL: /user_center/choose_avatar
// 选择默认头像
func chooseAvatarHandler(handler *Handler) {
	if handler.Request.Method != "POST" {
		return
	}

	user, _ := currentUser(handler)

	avatar := handler.Request.FormValue("defaultAvatars")

	if avatar != "" {
		c := handler.DB.C(USERS)
		c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": avatar}})
	}

	handler.redirect("/user_center", http.StatusFound)
}

// URl: /user_center/get_gravatar
// 从 Gravatar 获取头像
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

// URL /user_center/edit_info
// 修改用户资料
func editUserInfoHandler(handler *Handler) {
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

				handler.renderTemplate("user_center/info_form.html", BASE, map[string]interface{}{
					"user":        user,
					"profileForm": profileForm,
					"active":      "edit_info",
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
			handler.redirect("/user_center/edit_info", http.StatusFound)
			return
		}
	}

	handler.renderTemplate("user_center/info_form.html", BASE, map[string]interface{}{
		"user":        user,
		"profileForm": profileForm,
		"active":      "edit_info",
	})
}

// URL: /user_center/change_password
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
			if user.CheckPassword(form.Value("current_password")) {
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

	handler.renderTemplate("user_center/change_password.html", BASE, map[string]interface{}{
		"form":   form,
		"active": "change_password",
	})
}

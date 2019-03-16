package actions

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/Youngyezi/geetest"
	"github.com/tango-contrib/renders"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/conf"
	"github.com/jimmykuu/gopher/models"
)

// Signin 登录
type Signin struct {
	RenderBase
}

// Get /signin 登录页面
func (a *Signin) Get() error {
	g := geetest.New(conf.Config.GtCaptchaId, conf.Config.GtPrivateKey)
	p := url.Values{
		"client": {"web"},
	}

	resp := g.PreProcess(p)

	var next = a.Form("next", "/")
	return a.Render("account/signin.html", renders.T{
		"title":       "登录",
		"next":        next,
		"gt":          resp["gt"],
		"challenge":   resp["challenge"],
		"success":     resp["success"],
		"new_captcha": resp["new_captcha"],
	})
}

// Signup 注册
type Signup struct {
	RenderBase
}

// Get /signup 注册页面
func (a *Signup) Get() error {
	g := geetest.New(conf.Config.GtCaptchaId, conf.Config.GtPrivateKey)
	p := url.Values{
		"client": {"web"},
	}

	resp := g.PreProcess(p)

	return a.Render("account/signup.html", renders.T{
		"title":       "注册",
		"gt":          resp["gt"],
		"challenge":   resp["challenge"],
		"success":     resp["success"],
		"new_captcha": resp["new_captcha"],
	})
}

// AccountIndex 账户首页
type AccountIndex struct {
	RenderBase
}

// Get /member/:username
func (a *AccountIndex) Get() error {
	username := a.Param("username")
	c := a.DB.C(models.USERS)
	user := models.User{}
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		return errors.New("会员未找到")
	}

	topics, pagination, err := GetTopics(a, a.DB, bson.M{"content.type": models.TypeTopic, "content.createdby": user.Id_})

	if err != nil {
		return err
	}

	return a.Render("account/index.html", renders.T{
		"title":      username,
		"member":     user,
		"topics":     topics,
		"pagination": pagination,
		"url":        fmt.Sprintf("/member/%s", username),
	})
}

// AccountComments 会员的所有回复
type AccountComments struct {
	RenderBase
}

// Get /member/:username/comments
func (a *AccountComments) Get() error {
	page := a.FormInt("p", 1)
	if page <= 0 {
		page = 1
	}

	username := a.Param("username")
	c := a.DB.C(models.USERS)
	user := models.User{}

	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		return errors.New("会员未找到")
	}

	var comments []models.Comment

	c = a.DB.C(models.COMMENTS)

	pagination := NewPagination(c.Find(bson.M{"createdby": user.Id_, "type": models.TypeTopic}).Sort("-createdat"), PerPage)

	query, err := pagination.Page(page)

	query.All(&comments)

	return a.Render("account/comments.html", renders.T{
		"title":      fmt.Sprintf("%s 的评论", username),
		"member":     user,
		"pagination": pagination,
		"comments":   comments,
		"url":        fmt.Sprintf("/member/%s/comments", username),
	})
}

// AccountCollections 用户收藏
type AccountCollections struct {
	RenderBase
}

// Get /member/:username/collections
func (a *AccountCollections) Get() error {
	username := a.Param("username")
	c := a.DB.C(models.USERS)
	user := models.User{}
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		return errors.New("会员未找到")
	}

	var topicIDs = []bson.ObjectId{}

	for _, topicCollected := range user.TopicsCollected {
		topicIDs = append(topicIDs, bson.ObjectIdHex(topicCollected.TopicId))
	}

	var conditions = bson.M{
		"content.type": models.TypeTopic,
		"_id":          bson.M{"$in": topicIDs},
	}

	topics, pagination, err := GetTopics(a, a.DB, conditions)

	if err != nil {
		return err
	}

	return a.Render("account/collections.html", renders.T{
		"title":      fmt.Sprintf("%s 的收藏", username),
		"member":     user,
		"pagination": pagination,
		"topics":     topics,
		"url":        fmt.Sprintf("/member/%s/collections", username),
	})
}

// AccountBlock 用户禁言
type AccountBlock struct {
	RenderBase
}

// Get /member/:username/block
func (a *AccountBlock) Get() error {
	if !a.User.IsSuperuser {
		return errors.New("没有该权限")
	}

	username := a.Param("username")
	c := a.DB.C(models.USERS)
	user := models.User{}
	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		return errors.New("会员未找到")
	}

	c.Update(bson.M{"username": username}, bson.M{"$set": bson.M{"is_blocked": true}})

	var nexts = a.FormStrings("next")
	var next = fmt.Sprintf("/member/%s", username)
	if len(nexts) > 0 {
		next = nexts[0]
	}

	a.Redirect(next)

	return nil
}

// ListUsers 会员列表
type ListUsers struct {
	RenderBase
}

// LatestUsers 最新会员
type LatestUsers struct {
	ListUsers
}

// Get /members
func (a *LatestUsers) Get() error {
	var members []models.User
	c := a.DB.C(models.USERS)
	c.Find(nil).Sort("-joinedat").Limit(40).All(&members)

	return a.Render("account/members.html", renders.T{
		"title":   "最新会员",
		"members": members,
	})
}

// AllUsers 所有会员带分页
type AllUsers struct {
	ListUsers
}

// Get /members/all?p=1
func (a *ListUsers) Get() error {
	page := a.FormInt("p", 1)
	if page <= 0 {
		page = 1
	}

	var members []models.User
	c := a.DB.C(models.USERS)

	pagination := NewPagination(c.Find(nil).Sort("joinedat"), 40)

	query, err := pagination.Page(page)
	if err != nil {
		return err
	}

	query.All(&members)

	return a.Render("account/members_all.html", renders.T{
		"title":      "所有会员",
		"members":    members,
		"pagination": pagination,
		"page":       page,
		"url":        "/members/all",
	})
}

// ForgotPassword 忘记密码
type ForgotPassword struct {
	RenderBase
}

// Get /forgot_password
func (a *ForgotPassword) Get() error {
	g := geetest.New(conf.Config.GtCaptchaId, conf.Config.GtPrivateKey)
	p := url.Values{
		"client": {"web"},
	}

	resp := g.PreProcess(p)

	return a.Render("account/forgot_password.html", renders.T{
		"title":       "忘记密码",
		"gt":          resp["gt"],
		"challenge":   resp["challenge"],
		"success":     resp["success"],
		"new_captcha": resp["new_captcha"],
	})
}

// ResetPassword 重设密码
type ResetPassword struct {
	RenderBase
}

// Get /reset/:code
func (a *ResetPassword) Get() error {
	code := a.Param("code")
	println(">>>>", code)
	var user models.User
	c := a.DB.C(models.USERS)
	err := c.Find(bson.M{"resetcode": code}).One(&user)
	if err != nil {
		a.NotFound("错误的重设代码")
		return nil
	}

	return a.Render("account/reset_password.html", renders.T{
		"title":          "重设密码",
		"code":           code,
		"reset_username": user.Username,
	})
}

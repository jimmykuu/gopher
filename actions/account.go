package actions

import (
	"github.com/jimmykuu/gopher/models"
	"github.com/tango-contrib/renders"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Signin 登录
type Signin struct {
	RenderBase
}

// Get /signin 登录页面
func (a *Signin) Get() error {
	var next = a.Form("next", "/")
	return a.Render("account/signin.html", renders.T{
		"title": "登录",
		"next":  next,
	})
}

// Signup 注册
type Signup struct {
	RenderBase
}

// Get /signup 注册页面
func (a *Signup) Get() error {
	return a.Render("account/signup.html", renders.T{
		"title": "注册",
	})
}

// AccountIndex 账户首页
type AccountIndex struct {
	RenderBase
}

// Get /menuber/:username
func (a *AccountIndex) Get() error {
	username := a.Param("username")
	session, DB := models.GetSessionAndDB()
	defer session.Close()
	c := DB.C(models.USERS)

	user := models.User{}

	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		a.NotFound("会员未找到")
		return nil
	}

	return a.Render("account/index.html", renders.T{
		"title":  username,
		"member": user,
	})
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

	query.(*mgo.Query).All(&members)

	return a.Render("account/members_all.html", renders.T{
		"title":      "所有会员",
		"members":    members,
		"pagination": pagination,
		"page":       page,
		"url":        "/members/all",
	})
}

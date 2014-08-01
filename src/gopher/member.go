/*
会员
*/

package gopher

import (
	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

// 显示最新加入的会员
// URL: /members
func membersHandler(handler Handler) {
	c := handler.DB.C(USERS)
	var newestMembers []User
	c.Find(nil).Sort("-joinedat").Limit(40).All(&newestMembers)

	membersCount, _ := c.Find(nil).Count()

	renderTemplate(handler, "member/index.html", BASE, map[string]interface{}{
		"newestMembers": newestMembers,
		"membersCount":  membersCount,
		"active":        "members",
	})
}

// 显示所有会员
// URL: /members/all
func allMembersHandler(handler Handler) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	c := handler.DB.C(USERS)

	pagination := NewPagination(c.Find(nil).Sort("joinedat"), "/members/all", 40)

	var members []User

	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	query.All(&members)

	renderTemplate(handler, "member/list.html", BASE, map[string]interface{}{
		"members":    members,
		"active":     "members",
		"pagination": pagination,
		"page":       page,
	})
}

// URL: /member/{username}
// 显示用户信息
func memberInfoHandler(handler Handler) {
	vars := mux.Vars(handler.Request)
	username := vars["username"]
	c := handler.DB.C(USERS)

	user := User{}

	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(handler, "会员未找到", "会员未找到", "error")
		return
	}

	renderTemplate(handler, "account/info.html", BASE, map[string]interface{}{
		"user":   user,
		"active": "members",
	})
}

// URL: /member/{username}/collect/
// 用户收集的topic
func memberTopicsCollectedHandler(handler Handler) {
	page, err := getPage(handler.Request)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
	}
	vars := mux.Vars(handler.Request)
	username := vars["username"]
	c := handler.DB.C(USERS)
	user := User{}
	err = c.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		message(handler, "会员未找到", "会员未找到", "error")
	}
	renderTemplate(handler, "account/topics.html", BASE, map[string]interface{}{
		"user":       user,
		"collects":   user.TopicsCollected,
		"pagination": pagination,
		"page":       page,
		"active":     "members",
	})
}

// URL: /member/{username}/topics
// 用户发表的所有主题
func memberTopicsHandler(handler Handler) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	vars := mux.Vars(handler.Request)
	username := vars["username"]
	c := handler.DB.C(USERS)

	user := User{}
	err = c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(handler, "会员未找到", "会员未找到", "error")
		return
	}

	c = handler.DB.C("contents")

	pagination := NewPagination(c.Find(bson.M{"content.createdby": user.Id_, "content.type": TypeTopic}).Sort("-latestrepliedat"), "/member/"+username+"/topics", PerPage)

	var topics []Topic

	query, err := pagination.Page(page)

	if err != nil {
		message(handler, "没有找到页面", "没有找到页面", "error")
		return
	}

	query.All(&topics)

	renderTemplate(handler, "account/topics.html", BASE, map[string]interface{}{
		"user":       user,
		"topics":     topics,
		"pagination": pagination,
		"page":       page,
		"active":     "members",
	})
}

// /member/{username}/replies
// 用户的所有回复
func memberRepliesHandler(handler Handler) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	vars := mux.Vars(handler.Request)
	username := vars["username"]
	c := handler.DB.C(USERS)

	user := User{}
	err = c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(handler, "会员未找到", "会员未找到", "error")
		return
	}

	if err != nil {
		message(handler, "没有找到页面", "没有找到页面", "error")
		return
	}

	var replies []Comment

	c = handler.DB.C(COMMENTS)

	pagination := NewPagination(c.Find(bson.M{"createdby": user.Id_, "type": TypeTopic}).Sort("-createdat"), "/member/"+username+"/replies", PerPage)

	query, err := pagination.Page(page)

	query.All(&replies)

	renderTemplate(handler, "account/replies.html", BASE, map[string]interface{}{
		"user":       user,
		"pagination": pagination,
		"page":       page,
		"replies":    replies,
		"active":     "members",
	})
}

// URL: /member/{username}/comments
func memmberNewsHandler(handler Handler) {
	page, err := getPage(handler.Request)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	vars := mux.Vars(handler.Request)
	username := vars["username"]
	c := handler.DB.C(USERS)
	user := User{}
	err = c.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		message(handler, "会员未找到", "会员未找到", "error")
		return
	}

	renderTemplate(handler, "account/news.html", BASE, map[string]interface{}{
		"user":     user,
		"page":     page,
		"comments": user.RecentReplies,
		"ats":      user.RecentAts,

		"active": "members",
	})
}

// URL: /member/{username}/comments
func memberAtsHandler(handler Handler) {
	return
}

// URL: /members/city/{cityName}
// 同城会员
func membersInTheSameCityHandler(handler Handler) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	cityName := mux.Vars(handler.Request)["cityName"]

	c := handler.DB.C(USERS)

	pagination := NewPagination(c.Find(bson.M{"location": cityName}).Sort("joinedat"), "/members/city/"+cityName, 40)

	var members []User

	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	query.All(&members)

	renderTemplate(handler, "member/list.html", BASE, map[string]interface{}{
		"members":    members,
		"active":     "members",
		"pagination": pagination,
		"page":       page,
		"city":       cityName,
	})
}

/*
会员
*/

package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

// 显示最新加入的会员
// URL: /members
func membersHandler(w http.ResponseWriter, r *http.Request) {
	c := DB.C("users")
	var newestMembers []User
	c.Find(nil).Sort("-joinedat").Limit(40).All(&newestMembers)

	membersCount, _ := c.Find(nil).Count()

	renderTemplate(w, r, "member/index.html", BASE, map[string]interface{}{
		"newestMembers": newestMembers,
		"membersCount":  membersCount,
		"active":        "members",
	})
}

// 显示所有会员
// URL: /members/all
func allMembersHandler(w http.ResponseWriter, r *http.Request) {
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	c := DB.C("users")

	pagination := NewPagination(c.Find(nil).Sort("joinedat"), "/members/all", 40)

	var members []User

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&members)

	renderTemplate(w, r, "member/list.html", BASE, map[string]interface{}{
		"members":    members,
		"active":     "members",
		"pagination": pagination,
		"page":       page,
	})
}

// URL: /member/{username}
// 显示用户信息
func memberInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	c := DB.C("users")

	user := User{}

	err := c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "会员未找到", "会员未找到", "error")
		return
	}

	renderTemplate(w, r, "account/info.html", BASE, map[string]interface{}{
		"user":   user,
		"active": "members",
	})
}

// URL: /member/{username}/topics
// 用户发表的所有主题
func memberTopicsHandler(w http.ResponseWriter, r *http.Request) {
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	c := DB.C("users")

	user := User{}
	err = c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "会员未找到", "会员未找到", "error")
		return
	}

	c = DB.C("contents")

	pagination := NewPagination(c.Find(bson.M{"content.createdby": user.Id_, "content.type": TypeTopic}).Sort("-latestrepliedat"), "/member/"+username+"/topics", PerPage)

	var topics []Topic

	query, err := pagination.Page(page)

	if err != nil {
		message(w, r, "没有找到页面", "没有找到页面", "error")
		return
	}

	query.All(&topics)

	renderTemplate(w, r, "account/topics.html", BASE, map[string]interface{}{
		"user":       user,
		"topics":     topics,
		"pagination": pagination,
		"page":       page,
		"active":     "members",
	})
}

// /member/{username}/replies
// 用户的所有回复
func memberRepliesHandler(w http.ResponseWriter, r *http.Request) {
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	c := DB.C("users")

	user := User{}
	err = c.Find(bson.M{"username": username}).One(&user)

	if err != nil {
		message(w, r, "会员未找到", "会员未找到", "error")
		return
	}

	if err != nil {
		message(w, r, "没有找到页面", "没有找到页面", "error")
		return
	}

	var replies []Comment

	c = DB.C("comments")

	pagination := NewPagination(c.Find(bson.M{"createdby": user.Id_, "type": TypeTopic}).Sort("-createdat"), "/member/"+username+"/replies", PerPage)

	query, err := pagination.Page(page)

	query.All(&replies)

	renderTemplate(w, r, "account/replies.html", BASE, map[string]interface{}{
		"user":       user,
		"pagination": pagination,
		"page":       page,
		"replies":    replies,
		"active":     "members",
	})
}

// URL: /members/city/{cityName}
// 同城会员
func membersInTheSameCityHandler(w http.ResponseWriter, r *http.Request) {
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	cityName := mux.Vars(r)["cityName"]

	c := DB.C("users")

	pagination := NewPagination(c.Find(bson.M{"location": cityName}).Sort("joinedat"), "/members/city/"+cityName, 40)

	var members []User

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&members)

	renderTemplate(w, r, "member/list.html", BASE, map[string]interface{}{
		"members":    members,
		"active":     "members",
		"pagination": pagination,
		"page":       page,
		"city":       cityName,
	})
}

/*
会员
*/

package gopher

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	at    = "at"
	reply = "reply"
)

func returnJson(w http.ResponseWriter, input interface{}) {
	js, err := json.Marshal(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

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

	query.(*mgo.Query).All(&members)

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
	pagination := NewPagination(user.TopicsCollected, "/member/"+username+"/collect", 3)
	collects, err := pagination.Page(page)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
	}
	renderTemplate(handler, "account/collects.html", BASE, map[string]interface{}{
		"user":       user,
		"collects":   collects,
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

	query.(*mgo.Query).All(&topics)

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

	query.(*mgo.Query).All(&replies)

	renderTemplate(handler, "account/replies.html", BASE, map[string]interface{}{
		"user":       user,
		"pagination": pagination,
		"page":       page,
		"replies":    replies,
		"active":     "members",
	})
}

// URL: /member/{username}/clear/{t}
func memmberNewsClear(handler Handler) {
	vars := mux.Vars(handler.Request)
	username := vars["username"]
	t := vars["t"]
	res := map[string]interface{}{}
	user, ok := currentUser(handler)
	if ok {
		if user.Username == username {
			var user User
			c := handler.DB.C(USERS)
			c.Find(bson.M{"username": username}).One(&user)
			if t == at {
				user.RecentAts = user.RecentAts[:0]
				c.Update(bson.M{"username": username}, bson.M{"$set": bson.M{"recentats": user.RecentAts}})
				res["status"] = true
			} else if t == reply {
				user.RecentReplies = user.RecentReplies[:0]
				c.Update(bson.M{"username": username}, bson.M{"$set": bson.M{"recentreplies": user.RecentReplies}})
				res["status"] = true
			} else {
				res["status"] = false
				res["error"] = "Wrong Type"
			}

		} else {
			res["status"] = false
			res["error"] = "Need authentication"
		}
	} else {
		res["status"] = false
		res["error"] = "No such User"
	}
	returnJson(handler.ResponseWriter, res)
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

	query.(*mgo.Query).All(&members)

	renderTemplate(handler, "member/list.html", BASE, map[string]interface{}{
		"members":    members,
		"active":     "members",
		"pagination": pagination,
		"page":       page,
		"city":       cityName,
	})
}

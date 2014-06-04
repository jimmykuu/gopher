/*
后台管理
*/

package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

// URL: /admin
// 后台管理首页
func adminHandler(handler Handler) {
	renderTemplate(handler, "admin/index.html", ADMIN, map[string]interface{}{})
}

// URL: /admin/users
// 列出所有用户
func adminListUsersHandler(handler Handler) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	var users []User
	c := DB.C(USERS)

	pagination := NewPagination(c.Find(nil).Sort("-joinedat"), "/admin/users", PerPage)

	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	query.All(&users)

	renderTemplate(handler, "admin/users.html", ADMIN, map[string]interface{}{"users": users, "pagination": pagination, "total": pagination.Count(), "page": page})
}

// URL: /admin/user/{userId}/activate
// 激活用户
func adminActivateUserHandler(handler Handler) {
	userId := mux.Vars(handler.Request)["userId"]

	c := DB.C(USERS)
	c.Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$set": bson.M{"isactive": true}})
	http.Redirect(handler.ResponseWriter, handler.Request, "/admin/users", http.StatusFound)
}

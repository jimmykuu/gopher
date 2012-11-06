/*
后台管理
*/

package main

import (
	"./wtforms"
	"code.google.com/p/gorilla/mux"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
)

// 管理页面的子菜单
const ADMIN_NAV = template.HTML(`<div class="span3">
	<ul class="nav nav-list" id="admin-sidebar">
		<li><a href="/admin/nodes"><i class="icon-chevron-right"></i>节点管理</a></li>
		<li><a href="/admin/site_categories"><i class="icon-chevron-right"></i> 站点内容管理</a></li>
		<li><a href="/admin/users"><i class="icon-chevron-right"></i> 用户管理</a></li>
	</ul>
</div>`)

// URL: /admin
// 后台管理首页
func adminHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/admin", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有后台管理权限", "error")
		return
	}

	renderTemplate(w, r, "admin/index.html", map[string]interface{}{"adminNav": ADMIN_NAV})
}

// URL: /admin/nodes
// 列出所有的节点
func adminListNodesHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/admin/nodes", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有列出查看节点的权限", "error")
		return
	}

	var nodes []Node
	c := db.C("nodes")
	c.Find(nil).All(&nodes)
	renderTemplate(w, r, "admin/nodes.html", map[string]interface{}{"adminNav": ADMIN_NAV, "nodes": nodes})
}

// URL: /admin/site_categories
// 列出所有的站点分类
func adminListSiteCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/admin/site_categories", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有查看所有站点分类的权限", "error")
		return
	}

	var categories []SiteCategory
	c := db.C("sitecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/site_categories.html", map[string]interface{}{"adminNav": ADMIN_NAV, "categories": categories})
}

// URL: /admin/site_category/new
// 新建站点分类
func adminNewSiteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/admin/site_category/new", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有新建站点分类的权限", "error")
		return
	}

	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/new_site_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		c := db.C("sitecategories")
		var category SiteCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(w, r, "admin/new_site_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		err = c.Insert(&SiteCategory{
			Id_:  bson.NewObjectId(),
			Name: form.Value("name"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/admin/site_category/new", http.StatusFound)
	}

	renderTemplate(w, r, "admin/new_site_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
}

// URL: /admin/node/new
// 新建节点
func adminNewNodeHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/node/new", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有新建节点的权限", "error")
		return
	}

	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", "", &wtforms.Required{}),
		wtforms.NewTextField("name", "名称", "", &wtforms.Required{}),
		wtforms.NewTextArea("description", "描述", "", &wtforms.Required{}),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			c := db.C("nodes")
			node := Node{}

			err := c.Find(bson.M{"id": form.Value("id")}).One(&node)

			if err == nil {
				form.AddError("id", "该ID已经存在")

				renderTemplate(w, r, "node/new.html", map[string]interface{}{"form": form, "adminNav": ADMIN_NAV})
				return
			}

			err = c.Find(bson.M{"name": form.Value("name")}).One(&node)

			if err == nil {
				form.AddError("name", "该名称已经存在")

				renderTemplate(w, r, "node/new.html", map[string]interface{}{"form": form, "adminNav": ADMIN_NAV})
				return
			}

			Id_ := bson.NewObjectId()
			err = c.Insert(&Node{
				Id_:         Id_,
				Id:          form.Value("id"),
				Name:        form.Value("name"),
				Description: form.Value("description")})

			if err != nil {
				panic(err)
			}

			http.Redirect(w, r, "/admin/node/new", http.StatusFound)
		}
	}

	renderTemplate(w, r, "node/new.html", map[string]interface{}{"form": form, "adminNav": ADMIN_NAV})
}

// URL: /admin/users
// 列出所有用户
func adminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/admin/users", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有查看所有用户的权限", "error")
		return
	}

	var users []User
	c := db.C("users")
	c.Find(nil).Sort("-joinedat").All(&users)

	renderTemplate(w, r, "admin/users.html", map[string]interface{}{"users": users, "adminNav": ADMIN_NAV})
}

// URL: /admin/user/{userId}/activate
// 激活用户
func adminActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin?next=/admin/users", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有权限", "你没有激活用户的权限", "error")
		return
	}

	userId := mux.Vars(r)["userId"]

	c := db.C("users")
	c.Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$set": bson.M{"isactive": true}})
	http.Redirect(w, r, "/admin/users", http.StatusFound)
}

/*
后台管理
*/

package gopher

import (
	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
)

// 管理页面的子菜单
const ADMIN_NAV = template.HTML(`<div class="span3">
	<ul class="nav nav-list" id="admin-sidebar">
		<li><a href="/admin/nodes"><i class="icon-chevron-right"></i>节点管理</a></li>
		<li><a href="/admin/site_categories"><i class="icon-chevron-right"></i> 站点分类管理</a></li>
		<li><a href="/admin/article_categories"><i class="icon-chevron-right"></i> 文章分类管理</a></li>
		<li><a href="/admin/package_categories"><i class="icon-chevron-right"></i> 包分类管理</a></li>
		<li><a href="/admin/users"><i class="icon-chevron-right"></i> 用户管理</a></li>
	</ul>
</div>`)

// URL: /admin
// 后台管理首页
func adminHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "admin/index.html", map[string]interface{}{"adminNav": ADMIN_NAV})
}

// URL: /admin/nodes
// 列出所有的节点
func adminListNodesHandler(w http.ResponseWriter, r *http.Request) {
	var nodes []Node
	c := DB.C("nodes")
	c.Find(nil).All(&nodes)
	renderTemplate(w, r, "admin/nodes.html", map[string]interface{}{"adminNav": ADMIN_NAV, "nodes": nodes})
}

// URL: /admin/site_categories
// 列出所有的站点分类
func adminListSiteCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []SiteCategory
	c := DB.C("sitecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/site_categories.html", map[string]interface{}{"adminNav": ADMIN_NAV, "categories": categories})
}

// URL: /admin/site_category/new
// 新建站点分类
func adminNewSiteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/new_site_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		c := DB.C("sitecategories")
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
	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", "", &wtforms.Required{}),
		wtforms.NewTextField("name", "名称", "", &wtforms.Required{}),
		wtforms.NewTextArea("description", "描述", "", &wtforms.Required{}),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			c := DB.C("nodes")
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
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	var users []User
	c := DB.C("users")

	pagination := NewPagination(c.Find(nil).Sort("-joinedat"), "/admin/users", PerPage)

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&users)

	renderTemplate(w, r, "admin/users.html", map[string]interface{}{"users": users, "adminNav": ADMIN_NAV, "pagination": pagination, "total": pagination.Count(), "page": page})
}

// URL: /admin/user/{userId}/activate
// 激活用户
func adminActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["userId"]

	c := DB.C("users")
	c.Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$set": bson.M{"isactive": true}})
	http.Redirect(w, r, "/admin/users", http.StatusFound)
}

// URL: /admin/article_categories
// 列出所有的文章分类
func adminListArticleCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []SiteCategory
	c := DB.C("articlecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/article_categories.html", map[string]interface{}{"adminNav": ADMIN_NAV, "categories": categories})
}

// URL: /admin/article_category/new
// 新建文章分类
func adminNewArticleCategoryHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/new_article_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		c := DB.C("articlecategories")
		var category ArticleCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(w, r, "admin/new_article_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		err = c.Insert(&ArticleCategory{
			Id_:  bson.NewObjectId(),
			Name: form.Value("name"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/admin/article_category/new", http.StatusFound)
	}

	renderTemplate(w, r, "admin/new_article_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
}

// URL: /admin/package_categories
// 列出所有的包分类
func adminListPackageCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []PackageCategory
	c := DB.C("packagecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/package_categories.html", map[string]interface{}{"adminNav": ADMIN_NAV, "categories": categories})
}

// URL: /admin/package_category/new
// 新建包分类
func adminNewPackageCategoryHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", "", wtforms.Required{}),
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/new_package_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		c := DB.C("packagecategories")
		var category PackageCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(w, r, "admin/new_package_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		err = c.Insert(&PackageCategory{
			Id_:  bson.NewObjectId(),
			Id:   form.Value("id"),
			Name: form.Value("name"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/admin/package_category/new", http.StatusFound)
	}

	renderTemplate(w, r, "admin/new_package_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
}

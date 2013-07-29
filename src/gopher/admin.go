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
		<li><a href="/admin/link_exchanges"><i class="icon-chevron-right"></i> 友情链接</a></li>
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

	renderTemplate(w, r, "admin/package_category_form.html", map[string]interface{}{
		"adminNav": ADMIN_NAV,
		"form":     form,
		"isNew":    true,
	})
}

// URL: /admin/package_category/{id}/edit
// 修改包分类
func adminEditPackageCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c := DB.C("packagecategories")
	var category PackageCategory
	c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&category)

	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", category.Id, wtforms.Required{}),
		wtforms.NewTextField("name", "名称", category.Name, wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/new_package_category.html", map[string]interface{}{"adminNav": ADMIN_NAV, "form": form})
			return
		}

		c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{
			"id":   form.Value("id"),
			"name": form.Value("name"),
		}})

		http.Redirect(w, r, "/admin/package_categories", http.StatusFound)
	}

	renderTemplate(w, r, "admin/package_category_form.html", map[string]interface{}{
		"adminNav": ADMIN_NAV,
		"form":     form,
		"isNew":    false,
	})
}

// URL: /admin/link_exchanges
// 友情链接列表
func adminListLinkExchangesHandler(w http.ResponseWriter, r *http.Request) {
	c := DB.C("link_exchanges")
	var linkExchanges []LinkExchange
	c.Find(nil).All(&linkExchanges)

	renderTemplate(w, r, "admin/link_exchanges.html", map[string]interface{}{
		"adminNav":      ADMIN_NAV,
		"linkExchanges": linkExchanges,
	})
}

// ULR: /admin/link_exchange/new
// 增加友链
func adminNewLinkExchangeHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
		wtforms.NewTextField("url", "URL", "", wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextField("description", "描述", "", wtforms.Required{}),
		wtforms.NewTextField("logo", "Logo", ""),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/link_exchange_form.html", map[string]interface{}{
				"adminNav": ADMIN_NAV,
				"form":     form,
				"isNew":    true,
			})
			return
		}

		c := DB.C("link_exchanges")
		var linkExchange LinkExchange
		err := c.Find(bson.M{"url": form.Value("url")}).One(&linkExchange)

		if err == nil {
			form.AddError("url", "该URL已经有了")
			renderTemplate(w, r, "admin/link_exchange_category.html", map[string]interface{}{
				"adminNav": ADMIN_NAV,
				"form":     form,
				"isNew":    true,
			})
			return
		}

		err = c.Insert(&LinkExchange{
			Id_:         bson.NewObjectId(),
			Name:        form.Value("name"),
			URL:         form.Value("url"),
			Description: form.Value("description"),
			Logo:        form.Value("logo"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/admin/link_exchanges", http.StatusFound)
		return
	}

	renderTemplate(w, r, "admin/link_exchange_form.html", map[string]interface{}{
		"adminNav": ADMIN_NAV,
		"form":     form,
		"isNew":    true,
	})
}

// URL: /admin/link_exchange/{linkExchangeId}/edit
// 编辑友情链接
func adminEditLinkExchangeHandler(w http.ResponseWriter, r *http.Request) {
	linkExchangeId := mux.Vars(r)["linkExchangeId"]

	c := DB.C("link_exchanges")
	var linkExchange LinkExchange
	c.Find(bson.M{"_id": bson.ObjectIdHex(linkExchangeId)}).One(&linkExchange)

	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", linkExchange.Name, wtforms.Required{}),
		wtforms.NewTextField("url", "URL", linkExchange.URL, wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextField("description", "描述", linkExchange.Description, wtforms.Required{}),
		wtforms.NewTextField("logo", "Logo", linkExchange.Logo),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "admin/link_exchange_form.html", map[string]interface{}{
				"adminNav": ADMIN_NAV,
				"form":     form,
				"isNew":    false,
			})
			return
		}

		err := c.Update(bson.M{"_id": linkExchange.Id_}, bson.M{"$set": bson.M{
			"name":        form.Value("name"),
			"url":         form.Value("url"),
			"description": form.Value("description"),
			"logo":        form.Value("logo"),
		}})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/admin/link_exchanges", http.StatusFound)
		return
	}

	renderTemplate(w, r, "admin/link_exchange_form.html", map[string]interface{}{
		"adminNav": ADMIN_NAV,
		"form":     form,
		"isNew":    false,
	})
}

// URL: /admin/link_exchange/{linkExchangeId}/delete
// 删除友情链接
func adminDeleteLinkExchangeHandler(w http.ResponseWriter, r *http.Request) {
	linkExchangeId := mux.Vars(r)["linkExchangeId"]

	c := DB.C("link_exchanges")
	c.RemoveId(bson.ObjectIdHex(linkExchangeId))

	w.Write([]byte("true"))
}

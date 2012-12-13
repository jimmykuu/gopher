/*
酷站
*/

package main

import (
	"./wtforms"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
	"net/http"
)

// URL: /sites
// 酷站首页,列出所有分类及站点
func sitesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []SiteCategory
	c := db.C("sitecategories")
	c.Find(nil).All(&categories)
	renderTemplate(w, r, "site/index.html", map[string]interface{}{"categories": categories})
}

// URL: /site/new
// 提交站点
func newSiteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	var categories []SiteCategory
	c := db.C("sitecategories")
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewTextField("name", "网站名称", "", wtforms.Required{}),
		wtforms.NewTextField("url", "地址", "", wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", ""),
		wtforms.NewSelectField("category", "分类", choices, "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "site/form.html", map[string]interface{}{"form": form, "action": "/site/new", "title": "新建"})
			return
		}

		var site Site
		c = db.C("sites")
		err := c.Find(bson.M{"url": form.Value("url")}).One(&site)
		if err == nil {
			form.AddError("url", "该站点已经有了")
			renderTemplate(w, r, "site/form.html", map[string]interface{}{"form": form, "action": "/site/new", "title": "新建"})
			return
		}

		Id_ := bson.NewObjectId()

		c = db.C("sites")

		c.Insert(&Site{
			Id_:         Id_,
			Name:        form.Value("name"),
			Url:         form.Value("url"),
			Description: form.Value("description"),
			CategoryId:  bson.ObjectIdHex(form.Value("category")),
			UserId:      user.Id_,
		})

		http.Redirect(w, r, "/sites#site-"+Id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(w, r, "site/form.html", map[string]interface{}{"form": form, "action": "/site/new", "title": "新建"})
}

// URL: /site/{siteId}/edit
// 修改提交过的站点信息,提交者自己或者管理员可以修改
func editSiteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	siteId := mux.Vars(r)["siteId"]

	var site Site
	c := db.C("sites")

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(siteId)}).One(&site)

	if err != nil {
		message(w, r, "错误的连接", "错误的连接", "error")
		return
	}

	if !site.CanEdit(user.Username) {
		message(w, r, "没有权限", "你没有权限可以修改站点", "error")
		return
	}

	var categories []SiteCategory
	c = db.C("sitecategories")
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewTextField("name", "网站名称", site.Name, wtforms.Required{}),
		wtforms.NewTextField("url", "地址", site.Url, wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", site.Description),
		wtforms.NewSelectField("category", "分类", choices, site.CategoryId.Hex(), wtforms.Required{}),
	)

	if r.Method == "POST" && form.Validate(r) {
		// 检查是否用重复
		var site2 Site
		c = db.C("sites")
		err := c.Find(bson.M{"url": form.Value("url"), "_id": bson.M{"$ne": site.Id_}}).One(&site2)
		if err == nil {
			form.AddError("url", "该站点已经有了")
			renderTemplate(w, r, "site/form.html", map[string]interface{}{"form": form, "action": "/site/" + siteId + "/edit", "title": "编辑"})
			return
		}

		c.Update(bson.M{"_id": site.Id_},
			bson.M{"$set": bson.M{
				"name":        form.Value("name"),
				"url":         form.Value("url"),
				"description": form.Value("description"),
				"categoryid":  bson.ObjectIdHex(form.Value("category")),
			},
			})

		http.Redirect(w, r, "/sites#site-"+site.Id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(w, r, "site/form.html", map[string]interface{}{"form": form, "action": "/site/" + siteId + "/edit", "title": "编辑"})
}

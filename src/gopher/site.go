/*
酷站
*/

package gopher

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /sites
// 酷站首页,列出所有分类及站点
func sitesHandler(handler Handler) {
	var categories []SiteCategory
	c := handler.DB.C(SITE_CATEGORIES)
	c.Find(nil).All(&categories)
	renderTemplate(handler, "site/index.html", BASE, map[string]interface{}{
		"categories": categories,
		"active":     "site",
	})
}

// URL: /site/new
// 提交站点
func newSiteHandler(handler Handler) {
	user, _ := currentUser(handler)

	var categories []SiteCategory
	c := handler.DB.C(SITE_CATEGORIES)
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

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			renderTemplate(handler, "site/form.html", BASE, map[string]interface{}{"form": form, "action": "/site/new", "title": "新建"})
			return
		}

		var site Site
		c = handler.DB.C(CONTENTS)
		err := c.Find(bson.M{"url": form.Value("url")}).One(&site)
		if err == nil {
			form.AddError("url", "该站点已经有了")
			renderTemplate(handler, "site/form.html", BASE, map[string]interface{}{"form": form, "action": "/site/new", "title": "新建"})
			return
		}

		id_ := bson.NewObjectId()

		c.Insert(&Site{
			Id_: id_,
			Content: Content{
				Id_:       id_,
				Type:      TypeSite,
				Title:     form.Value("name"),
				Markdown:  form.Value("description"),
				CreatedBy: user.Id_,
				CreatedAt: time.Now(),
			},
			Url:        form.Value("url"),
			CategoryId: bson.ObjectIdHex(form.Value("category")),
		})

		http.Redirect(handler.ResponseWriter, handler.Request, "/sites#site-"+id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(handler, "site/form.html", BASE, map[string]interface{}{
		"form":   form,
		"action": "/site/new",
		"title":  "新建",
		"active": "site",
	})
}

// URL: /site/{siteId}/edit
// 修改提交过的站点信息,提交者自己或者管理员可以修改
func editSiteHandler(handler Handler) {
	user, _ := currentUser(handler)

	siteId := mux.Vars(handler.Request)["siteId"]
	if !bson.IsObjectIdHex(siteId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	var site Site
	c := handler.DB.C(CONTENTS)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(siteId), "content.type": TypeSite}).One(&site)

	if err != nil {
		message(handler, "错误的连接", "错误的连接", "error")
		return
	}

	if !site.CanEdit(user.Username, handler.DB) {
		message(handler, "没有权限", "你没有权限可以修改站点", "error")
		return
	}

	var categories []SiteCategory
	c = handler.DB.C(SITE_CATEGORIES)
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewTextField("name", "网站名称", site.Title, wtforms.Required{}),
		wtforms.NewTextField("url", "地址", site.Url, wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", site.Markdown),
		wtforms.NewSelectField("category", "分类", choices, site.CategoryId.Hex(), wtforms.Required{}),
	)

	if handler.Request.Method == "POST" && form.Validate(handler.Request) {
		// 检查是否用重复
		var site2 Site
		c = handler.DB.C(CONTENTS)
		err := c.Find(bson.M{"url": form.Value("url"), "_id": bson.M{"$ne": site.Id_}}).One(&site2)
		if err == nil {
			form.AddError("url", "该站点已经有了")
			renderTemplate(handler, "site/form.html", BASE, map[string]interface{}{"form": form, "action": "/site/" + siteId + "/edit", "title": "编辑"})
			return
		}

		c.Update(bson.M{"_id": site.Id_},
			bson.M{"$set": bson.M{
				"content.title":     form.Value("name"),
				"content.markdown":  form.Value("description"),
				"content.updatedby": user.Id_.Hex(),
				"content.updatedat": time.Now(),
				"url":               form.Value("url"),
				"categoryid":        bson.ObjectIdHex(form.Value("category")),
			},
			})

		http.Redirect(handler.ResponseWriter, handler.Request, "/sites#site-"+site.Id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(handler, "site/form.html", BASE, map[string]interface{}{
		"form":   form,
		"action": "/site/" + siteId + "/edit",
		"title":  "编辑",
		"active": "site",
	})
}

// URL: /site/{siteId}/delete
// 删除站点,提交者自己或者管理员可以删除
func deleteSiteHandler(handler Handler) {
	user, _ := currentUser(handler)

	siteId := mux.Vars(handler.Request)["siteId"]

	if !bson.IsObjectIdHex(siteId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	var site Site
	c := handler.DB.C(CONTENTS)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(siteId)}).One(&site)

	if err != nil {
		message(handler, "错误的连接", "错误的连接", "error")
		return
	}

	if !site.CanEdit(user.Username, handler.DB) {
		message(handler, "没有权限", "你没有权限可以删除站点", "error")
		return
	}

	c.Remove(bson.M{"_id": site.Id_})

	http.Redirect(handler.ResponseWriter, handler.Request, "/sites", http.StatusFound)
}

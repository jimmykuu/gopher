package gopher

import (
	"net/http"

	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2/bson"
)

// URL: /admin/site_categories
// 列出所有的站点分类
func adminListSiteCategoriesHandler(handler *Handler) {
	var categories []SiteCategory
	c := handler.DB.C(SITE_CATEGORIES)
	c.Find(nil).All(&categories)

	handler.renderTemplate("admin/site_categories.html", ADMIN, map[string]interface{}{"categories": categories})
}

// URL: /admin/site_category/new
// 新建站点分类
func adminNewSiteCategoryHandler(handler *Handler) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			handler.renderTemplate("site_category/new.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c := handler.DB.C(SITE_CATEGORIES)
		var category SiteCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			handler.renderTemplate("site_category/new.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		err = c.Insert(&SiteCategory{
			Id_:  bson.NewObjectId(),
			Name: form.Value("name"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/site_category/new", http.StatusFound)
	}

	handler.renderTemplate("site_category/new.html", ADMIN, map[string]interface{}{"form": form})
}

package gopher

import (
	"net/http"

	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /admin/site_categories
// 列出所有的站点分类
func adminListSiteCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []SiteCategory
	c := DB.C("sitecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/site_categories.html", ADMIN, map[string]interface{}{"categories": categories})
}

// URL: /admin/site_category/new
// 新建站点分类
func adminNewSiteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "site_category/new.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c := DB.C("sitecategories")
		var category SiteCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(w, r, "site_category/new.html", ADMIN, map[string]interface{}{"form": form})
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

	renderTemplate(w, r, "site_category/new.html", ADMIN, map[string]interface{}{"form": form})
}

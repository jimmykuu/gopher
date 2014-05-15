package gopher

import (
	"net/http"

	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /admin/article_categories
// 列出所有的文章分类
func adminListArticleCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []SiteCategory
	c := DB.C(ARTICLE_CATEGORIES)
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/article_categories.html", ADMIN, map[string]interface{}{"categories": categories})
}

// URL: /admin/article_category/new
// 新建文章分类
func adminNewArticleCategoryHandler(w http.ResponseWriter, r *http.Request) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "article_category/new.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c := DB.C(ARTICLE_CATEGORIES)
		var category ArticleCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(w, r, "article_category/new.html", ADMIN, map[string]interface{}{"form": form})
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

	renderTemplate(w, r, "article_category/new.html", ADMIN, map[string]interface{}{"form": form})
}

package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /admin/package_categories
// 列出所有的包分类
func adminListPackageCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []PackageCategory
	c := DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "admin/package_categories.html", ADMIN, map[string]interface{}{"categories": categories})
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
			renderTemplate(w, r, "package_category/form.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c := DB.C(PACKAGE_CATEGORIES)
		var category PackageCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(w, r, "package_category/form.html", ADMIN, map[string]interface{}{"form": form})
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

	renderTemplate(w, r, "package_category/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": true,
	})
}

// URL: /admin/package_category/{id}/edit
// 修改包分类
func adminEditPackageCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c := DB.C(PACKAGE_CATEGORIES)
	var category PackageCategory
	c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&category)

	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", category.Id, wtforms.Required{}),
		wtforms.NewTextField("name", "名称", category.Name, wtforms.Required{}),
	)

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "package_category/form.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{
			"id":   form.Value("id"),
			"name": form.Value("name"),
		}})

		http.Redirect(w, r, "/admin/package_categories", http.StatusFound)
	}

	renderTemplate(w, r, "package_category/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": false,
	})
}

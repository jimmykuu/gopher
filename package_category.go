package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2/bson"
)

// URL: /admin/package_categories
// 列出所有的包分类
func adminListPackageCategoriesHandler(handler Handler) {
	var categories []PackageCategory
	c := handler.DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	renderTemplate(handler, "admin/package_categories.html", ADMIN, map[string]interface{}{"categories": categories})
}

// URL: /admin/package_category/new
// 新建包分类
func adminNewPackageCategoryHandler(handler Handler) {
	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", "", wtforms.Required{}),
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			renderTemplate(handler, "package_category/form.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c := handler.DB.C(PACKAGE_CATEGORIES)
		var category PackageCategory
		err := c.Find(bson.M{"name": form.Value("name")}).One(&category)

		if err == nil {
			form.AddError("name", "该名称已经有了")
			renderTemplate(handler, "package_category/form.html", ADMIN, map[string]interface{}{"form": form})
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

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/package_category/new", http.StatusFound)
	}

	renderTemplate(handler, "package_category/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": true,
	})
}

// URL: /admin/package_category/{id}/edit
// 修改包分类
func adminEditPackageCategoryHandler(handler Handler) {
	id := mux.Vars(handler.Request)["id"]
	c := handler.DB.C(PACKAGE_CATEGORIES)
	var category PackageCategory
	c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&category)

	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", category.Id, wtforms.Required{}),
		wtforms.NewTextField("name", "名称", category.Name, wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			renderTemplate(handler, "package_category/form.html", ADMIN, map[string]interface{}{"form": form})
			return
		}

		c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{
			"id":   form.Value("id"),
			"name": form.Value("name"),
		}})

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/package_categories", http.StatusFound)
	}

	renderTemplate(handler, "package_category/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": false,
	})
}

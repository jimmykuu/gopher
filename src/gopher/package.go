/*
第三方包
*/

package gopher

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
	"time"
	"wtforms"
)

// URL: /packages
// 列出最新的一些第上方包
func packagesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []PackageCategory

	c := db.C("packagecategories")
	c.Find(nil).All(&categories)

	var latestPackages []Package
	c = db.C("packages")
	c.Find(nil).Sort("-createdat").Limit(10).All(&latestPackages)

	renderTemplate(w, r, "package/index.html", map[string]interface{}{"categories": categories, "latestPackages": latestPackages})
}

// URL: /package/new
// 新建第三方包
func newPackageHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	var categories []PackageCategory

	c := db.C("packagecategories")
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
		wtforms.NewSelectField("category_id", "分类", choices, ""),
		wtforms.NewTextField("url", "网址", "", wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", "", wtforms.Required{}),
	)

	if r.Method == "POST" && form.Validate(r) {
		c = db.C("packages")
		id := bson.NewObjectId()
		categoryId := bson.ObjectIdHex(form.Value("category_id"))
		c.Insert(&Package{
			Id_:        id,
			UserId:     user.Id_,
			CategoryId: categoryId,
			Name:       form.Value("name"),
			Url:        form.Value("url"),
			Markdown:   form.Value("description"),
			Html:       template.HTML(form.Value("html")),
			CreatedAt:  time.Now(),
		})

		c = db.C("packagecategories")
		// 增加数量
		c.Update(bson.M{"_id": categoryId}, bson.M{"$inc": bson.M{"packagecount": 1}})
	}
	renderTemplate(w, r, "package/form.html", map[string]interface{}{"form": form, "title": "提交第三方包", "action": "/package/new"})
}

// URL: /packages/{categoryId}
// 根据类别列出包
func listPackagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryId := vars["categoryId"]
	c := db.C("packagecategories")

	category := PackageCategory{}
	err := c.Find(bson.M{"id": categoryId}).One(&category)

	if err != nil {
		message(w, r, "没有该类别", "没有该类别", "error")
		return
	}

	var packages []Package

	c = db.C("packages")
	c.Find(bson.M{"categoryid": category.Id_}).Sort("name").All(&packages)

	var categories []PackageCategory

	c = db.C("packagecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "package/list.html", map[string]interface{}{"categories": categories, "packages": packages, "category": category})
}

// URL: /p/{packageId}
// 显示第三方包详情
func showPackageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packageId := vars["packageId"]

	c := db.C("packages")

	package_ := Package{}
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(packageId)}).One(&package_)

	if err != nil {
		message(w, r, "没找到该包", "请检查链接是否正确", "error")
		fmt.Println("showPackageHandler:", err.Error())
		return
	}

	var categories []PackageCategory

	c = db.C("packagecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "package/show.html", map[string]interface{}{"package": package_, "categories": categories})
}

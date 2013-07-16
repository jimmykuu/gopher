/*
第三方包
*/

package gopher

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

// URL: /packages
// 列出最新的一些第三方包
func packagesHandler(w http.ResponseWriter, r *http.Request) {
	var categories []PackageCategory

	c := DB.C("packagecategories")
	c.Find(nil).All(&categories)

	var latestPackages []Package
	c = DB.C("contents")
	c.Find(bson.M{"content.type": TypePackage}).Sort("-content.createdat").Limit(10).All(&latestPackages)

	renderTemplate(w, r, "package/index.html", map[string]interface{}{
		"categories":     categories,
		"latestPackages": latestPackages,
		"active":         "package",
	})
}

// URL: /package/new
// 新建第三方包
func newPackageHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)

	var categories []PackageCategory

	c := DB.C("packagecategories")
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
		c = DB.C("contents")
		id := bson.NewObjectId()
		categoryId := bson.ObjectIdHex(form.Value("category_id"))
		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)
		c.Insert(&Package{
			Content: Content{
				Id_:       id,
				Type:      TypePackage,
				Title:     form.Value("name"),
				Markdown:  form.Value("description"),
				Html:      template.HTML(html),
				CreatedBy: user.Id_,
				CreatedAt: time.Now(),
			},
			Id_:        id,
			CategoryId: categoryId,
			Url:        form.Value("url"),
		})

		c = DB.C("packagecategories")
		// 增加数量
		c.Update(bson.M{"_id": categoryId}, bson.M{"$inc": bson.M{"packagecount": 1}})

		http.Redirect(w, r, "/p/"+id.Hex(), http.StatusFound)
		return
	}
	renderTemplate(w, r, "package/form.html", map[string]interface{}{
		"form":   form,
		"title":  "提交第三方包",
		"action": "/package/new",
		"active": "package",
	})
}

// URL: /package/{packageId}/edit
// 编辑第三方包
func editPackageHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)

	vars := mux.Vars(r)
	packageId := vars["packageId"]

	package_ := Package{}
	c := DB.C("contents")
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(packageId), "content.type": TypePackage}).One(&package_)

	if err != nil {
		message(w, r, "没有该包", "没有该包", "error")
		return
	}

	if !package_.CanEdit(user.Username) {
		message(w, r, "没有权限", "你没有权限编辑该包", "error")
		return
	}

	var categories []PackageCategory

	c = DB.C("packagecategories")
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("name", "名称", package_.Title, wtforms.Required{}),
		wtforms.NewSelectField("category_id", "分类", choices, package_.CategoryId.Hex()),
		wtforms.NewTextField("url", "网址", package_.Url, wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", package_.Markdown, wtforms.Required{}),
	)

	if r.Method == "POST" && form.Validate(r) {
		c = DB.C("contents")
		categoryId := bson.ObjectIdHex(form.Value("category_id"))
		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)
		c.Update(bson.M{"_id": package_.Id_}, bson.M{"$set": bson.M{
			"categoryid":        categoryId,
			"url":               form.Value("url"),
			"content.title":     form.Value("name"),
			"content.markdown":  form.Value("description"),
			"content.html":      template.HTML(html),
			"content.updateDBy": user.Id_.Hex(),
			"content.updatedat": time.Now(),
		}})

		c = DB.C("packagecategories")
		if categoryId != package_.CategoryId {
			// 减少原来类别的包数量
			c.Update(bson.M{"_id": package_.CategoryId}, bson.M{"$inc": bson.M{"packagecount": -1}})
			// 增加新类别的包数量
			c.Update(bson.M{"_id": categoryId}, bson.M{"$inc": bson.M{"packagecount": 1}})
		}

		http.Redirect(w, r, "/p/"+package_.Id_.Hex(), http.StatusFound)
		return
	}

	form.SetValue("html", "")
	renderTemplate(w, r, "package/form.html", map[string]interface{}{
		"form":   form,
		"title":  "编辑第三方包",
		"action": "/p/" + packageId + "/edit",
		"active": "package",
	})
}

// URL: /packages/{categoryId}
// 根据类别列出包
func listPackagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryId := vars["categoryId"]
	c := DB.C("packagecategories")

	category := PackageCategory{}
	err := c.Find(bson.M{"id": categoryId}).One(&category)

	if err != nil {
		message(w, r, "没有该类别", "没有该类别", "error")
		return
	}

	var packages []Package

	c = DB.C("contents")
	c.Find(bson.M{"categoryid": category.Id_, "content.type": TypePackage}).Sort("name").All(&packages)

	var categories []PackageCategory

	c = DB.C("packagecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "package/list.html", map[string]interface{}{
		"categories": categories,
		"packages":   packages,
		"category":   category,
		"active":     "package",
	})
}

// URL: /p/{packageId}
// 显示第三方包详情
func showPackageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packageId := vars["packageId"]

	c := DB.C("contents")

	package_ := Package{}
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(packageId), "content.type": TypePackage}).One(&package_)

	if err != nil {
		message(w, r, "没找到该包", "请检查链接是否正确", "error")
		fmt.Println("showPackageHandler:", err.Error())
		return
	}

	var categories []PackageCategory

	c = DB.C("packagecategories")
	c.Find(nil).All(&categories)

	renderTemplate(w, r, "package/show.html", map[string]interface{}{
		"package":    package_,
		"categories": categories,
		"active":     "package",
	})
}

// URL: /p/{packageId}/delete
// 删除第三方包
func deletePackageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packageId := vars["packageId"]

	c := DB.C("contents")

	package_ := Package{}
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(packageId), "content.type": TypePackage}).One(&package_)

	if err != nil {
		return
	}

	c.Remove(bson.M{"_id": bson.ObjectIdHex(packageId)})

	// 修改分类下的数量
	c = DB.C("packagecategories")
	c.Update(bson.M{"_id": package_.CategoryId}, bson.M{"$inc": bson.M{"packagecount": -1}})

	http.Redirect(w, r, "/packages", http.StatusFound)
}

/*
酷站
*/

package main

import (
	"./wtforms"
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
	if _, ok := currentUser(r); !ok {
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
			renderTemplate(w, r, "site/new.html", map[string]interface{}{"form": form})
			return
		}

		var site Site
		c = db.C("sites")
		err := c.Find(bson.M{"url": form.Value("url")}).One(&site)
		if err == nil {
			form.AddError("url", "该站点已经有了")
			renderTemplate(w, r, "site/new.html", map[string]interface{}{"form": form})
			return
		}

		Id_ := bson.NewObjectId()

		session, _ := store.Get(r, "user")
		username, _ := session.Values["username"]
		username = username.(string)

		user := User{}
		c = db.C("users")
		c.Find(bson.M{"username": username}).One(&user)

		c = db.C("sites")

		c.Insert(&Site{
			Id_:         Id_,
			Name:        form.Value("name"),
			Url:         form.Value("url"),
			Description: form.Value("description"),
			CategoryId:  bson.ObjectIdHex(form.Value("category")),
			UserId:      user.Id_,
		})

		message(w, r, "站点提交成功", "感谢您提交站点", "success")
		return
	}

	renderTemplate(w, r, "site/new.html", map[string]interface{}{"form": form})
}

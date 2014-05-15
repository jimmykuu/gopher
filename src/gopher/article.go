/*
文章板块
*/

package gopher

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /article/new
// 新建文章
func newArticleHandler(w http.ResponseWriter, r *http.Request) {
	var categories []ArticleCategory
	c := DB.C(ARTICLE_CATEGORIES)
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("title", "标题", "", wtforms.Required{}),
		wtforms.NewTextField("original_source", "原始出处", "", wtforms.Required{}),
		wtforms.NewTextField("original_url", "原始链接", "", wtforms.URL{}),
		wtforms.NewSelectField("category", "分类", choices, ""),
	)

	if r.Method == "POST" && form.Validate(r) {
		user, _ := currentUser(r)

		c = DB.C(CONTENTS)

		id_ := bson.NewObjectId()

		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

		categoryId := bson.ObjectIdHex(form.Value("category"))
		err := c.Insert(&Article{
			Content: Content{
				Id_:       id_,
				Type:      TypeArticle,
				Title:     form.Value("title"),
				CreatedBy: user.Id_,
				CreatedAt: time.Now(),
			},
			Id_:            id_,
			CategoryId:     categoryId,
			OriginalSource: form.Value("original_source"),
			OriginalUrl:    form.Value("original_url"),
		})

		if err != nil {
			fmt.Println("newArticleHandler:", err.Error())
			return
		}

		http.Redirect(w, r, "/a/"+id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(w, r, "article/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "新建",
		"action": "/article/new",
		"active": "article",
	})
}

// URL: /articles
// 列出所有文章
func listArticlesHandler(w http.ResponseWriter, r *http.Request) {
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	//	var hotNodes []Node
	//	c := db.C("nodes")
	//	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").Limit(10).All(&hotNodes)

	//	var status Status
	//	c = db.C("status")
	//	c.Find(nil).One(&status)

	c := DB.C(CONTENTS)

	pagination := NewPagination(c.Find(bson.M{"content.type": TypeArticle}).Sort("-content.createdat"), "/articles", PerPage)

	var articles []Article

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&articles)

	renderTemplate(w, r, "article/index.html", BASE, map[string]interface{}{
		"articles":   articles,
		"pagination": pagination,
		"page":       page,
		"active":     "article",
	})
}

// URL: /a{articleId}/redirect
// 转到原文链接
func redirectArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleId := vars["articleId"]
	c := DB.C(CONTENTS)

	article := new(Article)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("redirectArticleHandler:", err.Error())
		return
	}

	c.UpdateId(bson.ObjectIdHex(articleId), bson.M{"$inc": bson.M{"content.hits": 1}})

	http.Redirect(w, r, article.OriginalUrl, http.StatusFound)
}

// URL: /a/{articleId}
// 显示文章
func showArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleId := vars["articleId"]
	c := DB.C(CONTENTS)

	article := Article{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("showArticleHandler:", err.Error())
		return
	}

	renderTemplate(w, r, "article/show.html", BASE, map[string]interface{}{
		"article": article,
		"active":  "article",
	})
}

// URL: /a/{articleId}/edit
// 编辑主题
func editArticleHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)

	articleId := mux.Vars(r)["articleId"]

	c := DB.C(CONTENTS)
	var article Article
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		message(w, r, "没有该文章", "没有该文章,不能编辑", "error")
		return
	}

	if !article.CanEdit(user.Username) {
		message(w, r, "没用该权限", "对不起,你没有权限编辑该文章", "error")
		return
	}

	var categorys []ArticleCategory
	c = DB.C(ARTICLE_CATEGORIES)
	c.Find(nil).All(&categorys)

	var choices []wtforms.Choice

	for _, category := range categorys {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("title", "标题", article.Title, wtforms.Required{}),
		wtforms.NewTextField("original_source", "原始出处", article.OriginalSource, wtforms.Required{}),
		wtforms.NewTextField("original_url", "原始链接", article.OriginalUrl, wtforms.URL{}),
		wtforms.NewSelectField("category", "分类", choices, article.CategoryId.Hex()),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			categoryId := bson.ObjectIdHex(form.Value("category"))
			c = DB.C(CONTENTS)
			err = c.Update(bson.M{"_id": article.Id_}, bson.M{"$set": bson.M{
				"categoryid":        categoryId,
				"originalsource":    form.Value("original_source"),
				"originalurl":       form.Value("original_url"),
				"content.title":     form.Value("title"),
				"content.updatedby": user.Id_.Hex(),
				"content.updatedat": time.Now(),
			}})

			if err != nil {
				fmt.Println("update error:", err.Error())
				return
			}

			http.Redirect(w, r, "/a/"+article.Id_.Hex(), http.StatusFound)
			return
		}
	}

	renderTemplate(w, r, "article/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "编辑",
		"action": "/a/" + articleId + "/edit",
		"active": "article",
	})
}

// URL: /a/{articleId}/delete
// 删除文章
func deleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)

	vars := mux.Vars(r)
	articleId := vars["articleId"]

	c := DB.C(CONTENTS)

	article := new(Article)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("deleteArticleHandler:", err.Error())
		return
	}

	if article.CanDelete(user.Username) {
		c.Remove(bson.M{"_id": article.Id_})

		c = DB.C(COMMENTS)
		c.Remove(bson.M{"contentid": article.Id_})

		http.Redirect(w, r, "/articles", http.StatusFound)
	}
}

/*
文章板块
*/

package gopher

import (
	"fmt"
	"html/template"
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
	c := DB.C("articlecategories")
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("title", "标题", "", wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", "", wtforms.Required{}),
		wtforms.NewTextField("original_source", "原始出处", "", wtforms.Required{}),
		wtforms.NewTextField("original_url", "原始链接", "", wtforms.URL{}),
		wtforms.NewSelectField("category", "分类", choices, ""),
	)

	if r.Method == "POST" && form.Validate(r) {
		session, _ := store.Get(r, "user")
		username, _ := session.Values["username"]
		username = username.(string)

		user := User{}
		c = DB.C("users")
		c.Find(bson.M{"username": username}).One(&user)

		c = DB.C("contents")

		id_ := bson.NewObjectId()

		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

		categoryId := bson.ObjectIdHex(form.Value("category"))
		err := c.Insert(&Article{
			Content: Content{
				Id_:       id_,
				Type:      TypeArticle,
				Title:     form.Value("title"),
				Markdown:  form.Value("content"),
				Html:      template.HTML(html),
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

	renderTemplate(w, r, "article/form.html", map[string]interface{}{
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

	c := DB.C("contents")

	pagination := NewPagination(c.Find(bson.M{"content.type": TypeArticle}).Sort("-content.createdat"), "/articles", PerPage)

	var articles []Article

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&articles)

	renderTemplate(w, r, "article/index.html", map[string]interface{}{
		"articles":   articles,
		"pagination": pagination,
		"page":       page,
		"active":     "article",
	})
}

// URL: /a/{articl_id}
// 显示文章
func showArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleId := vars["articleId"]
	c := DB.C("contents")

	article := Article{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("showArticleHandler:", err.Error())
		return
	}

	c.UpdateId(bson.ObjectIdHex(articleId), bson.M{"$inc": bson.M{"content.hits": 1}})

	renderTemplate(w, r, "article/show.html", map[string]interface{}{
		"article": article,
		"active":  "article",
	})
}

// URL: /a/{articleId}/edit
// 编辑主题
func editArticleHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)

	articleId := mux.Vars(r)["articleId"]

	c := DB.C("contents")
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
	c = DB.C("articlecategories")
	c.Find(nil).All(&categorys)

	var choices []wtforms.Choice

	for _, category := range categorys {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("title", "标题", article.Title, wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", article.Markdown, wtforms.Required{}),
		wtforms.NewTextField("original_source", "原始出处", article.OriginalSource, wtforms.Required{}),
		wtforms.NewTextField("original_url", "原始链接", article.OriginalUrl, wtforms.URL{}),
		wtforms.NewSelectField("category", "分类", choices, article.CategoryId.Hex()),
	)

	content := article.Markdown
	html := article.Html

	if r.Method == "POST" {
		if form.Validate(r) {
			html := form.Value("html")
			html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

			categoryId := bson.ObjectIdHex(form.Value("category"))
			c = DB.C("contents")
			err = c.Update(bson.M{"_id": article.Id_}, bson.M{"$set": bson.M{
				"categoryid":        categoryId,
				"originalsource":    form.Value("original_source"),
				"originalurl":       form.Value("original_url"),
				"content.title":     form.Value("title"),
				"content.markdown":  form.Value("content"),
				"content.html":      template.HTML(html),
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

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
	}

	renderTemplate(w, r, "article/form.html", map[string]interface{}{
		"form":    form,
		"title":   "编辑",
		"action":  "/a/" + articleId + "/edit",
		"html":    html,
		"content": content,
		"active":  "article",
	})
}

package main

import (
	"./wtforms"
	"github.com/gorilla/mux"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// URL: /article/new
// 新建文章
func newArticleHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := currentUser(r); !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	var categories []ArticleCategory
	c := db.C("articlecategories")
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
		c = db.C("users")
		c.Find(bson.M{"username": username}).One(&user)

		c = db.C("articles")

		Id_ := bson.NewObjectId()

		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

		categoryId := bson.ObjectIdHex(form.Value("category"))
		err := c.Insert(&Article{
			Id_:            Id_,
			CategoryId:     categoryId,
			UserId:         user.Id_,
			Title:          form.Value("title"),
			Markdown:       form.Value("content"),
			Html:           template.HTML(html),
			OriginalSource: form.Value("original_source"),
			OriginalUrl:    form.Value("original_url"),
			CreatedAt:      time.Now(),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/a/"+Id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(w, r, "article/form.html", map[string]interface{}{"form": form, "title": "新建", "action": "/article/new"})
}

// URL: /articles
// 列出所有文章
func listArticlesHandler(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("p")
	page := 1

	if p != "" {
		var err error
		page, err = strconv.Atoi(p)

		if err != nil {
			message(w, r, "页码错误", "页码错误", "error")
			return
		}
	}

	//	var hotNodes []Node
	//	c := db.C("nodes")
	//	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").Limit(10).All(&hotNodes)

	//	var status Status
	//	c = db.C("status")
	//	c.Find(nil).One(&status)

	c := db.C("articles")

	pagination := NewPagination(c.Find(nil).Sort("-createdat"), "/", PerPage)

	var articles []Article

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&articles)

	renderTemplate(w, r, "article/index.html", map[string]interface{}{"articles": articles, "pagination": pagination, "page": page})
}

// URL: /a/{articl_id}
// 显示文章
func showArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleId := vars["articleId"]
	c := db.C("articles")

	article := Article{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		println("err")
	}

	c = db.C("articles")
	c.UpdateId(bson.ObjectIdHex(articleId), bson.M{"$inc": bson.M{"hits": 1}})

	renderTemplate(w, r, "article/show.html", map[string]interface{}{"article": article})
}

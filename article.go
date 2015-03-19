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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// URL: /article/new
// 新建文章
func newArticleHandler(handler *Handler) {
	var categories []ArticleCategory
	c := handler.DB.C(ARTICLE_CATEGORIES)
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

	if handler.Request.Method == "POST" && form.Validate(handler.Request) {
		user, _ := currentUser(handler)

		c = handler.DB.C(CONTENTS)

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

		http.Redirect(handler.ResponseWriter, handler.Request, "/a/"+id_.Hex(), http.StatusFound)
		return
	}

	handler.renderTemplate("article/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "新建",
		"action": "/article/new",
		"active": "article",
	})
}

// URL: /article/go/{categoryId}
// 文章分类列表

func articlesInCategoryHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	categoryId := vars["categoryId"]

	c := handler.DB.C(ARTICLE_CATEGORIES)

	category := ArticleCategory{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(categoryId)}).One(&category)

	if err != nil {
		message(handler, "没有这个目录", "请联系管理员创建这个目录", "error")
		return
	}

	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	c = handler.DB.C(CONTENTS)

	pagination := NewPagination(c.Find(bson.M{"categoryid": bson.ObjectIdHex(categoryId), "content.type": TypeArticle}).Sort("-content.createdat"), "/", 20)
	var articles []Article
	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "没有找到页面", "没有找到页面", "error")
		return
	}
	err = query.(*mgo.Query).All(&articles)
	if err != nil {
		fmt.Println(err)
	}
	handler.renderTemplate("/article/index.html", BASE, map[string]interface{}{
		"articles": articles,
		"type":     category.Name,
	})
}

// URL: /articles
// 列出所有文章
func listArticlesHandler(handler *Handler) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	//	var hotNodes []Node
	//	c := db.C("nodes")
	//	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").Limit(10).All(&hotNodes)

	//	var status Status
	//	c = db.C("status")
	//	c.Find(nil).One(&status)

	c := handler.DB.C(CONTENTS)

	pagination := NewPagination(c.Find(bson.M{"content.type": TypeArticle}).Sort("-content.createdat"), "/articles", PerPage)

	var articles []Article

	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	query.(*mgo.Query).All(&articles)

	handler.renderTemplate("article/index.html", BASE, map[string]interface{}{
		"articles":   articles,
		"pagination": pagination,
		"page":       page,
		"active":     "article",
	})
}

// URL: /a{articleId}/redirect
// 转到原文链接
func redirectArticleHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	articleId := vars["articleId"]

	if !bson.IsObjectIdHex(articleId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}
	c := handler.DB.C(CONTENTS)

	article := new(Article)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("redirectArticleHandler:", err.Error())
		return
	}

	c.UpdateId(bson.ObjectIdHex(articleId), bson.M{"$inc": bson.M{"content.hits": 1}})

	http.Redirect(handler.ResponseWriter, handler.Request, article.OriginalUrl, http.StatusFound)
}

// URL: /a/{articleId}
// 显示文章
func showArticleHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	articleId := vars["articleId"]

	if !bson.IsObjectIdHex(articleId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}
	c := handler.DB.C(CONTENTS)

	article := Article{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("showArticleHandler:", err.Error())
		return
	}

	handler.renderTemplate("article/show.html", BASE, map[string]interface{}{
		"article": article,
		"active":  "article",
	})
}

// URL: /a/{articleId}/edit
// 编辑主题
func editArticleHandler(handler *Handler) {
	user, _ := currentUser(handler)

	articleId := mux.Vars(handler.Request)["articleId"]
	if !bson.IsObjectIdHex(articleId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	c := handler.DB.C(CONTENTS)
	var article Article
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		message(handler, "没有该文章", "没有该文章,不能编辑", "error")
		return
	}

	if !article.CanEdit(user.Username, handler.DB) {
		message(handler, "没用该权限", "对不起,你没有权限编辑该文章", "error")
		return
	}

	var categorys []ArticleCategory
	c = handler.DB.C(ARTICLE_CATEGORIES)
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

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			categoryId := bson.ObjectIdHex(form.Value("category"))
			c = handler.DB.C(CONTENTS)
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

			http.Redirect(handler.ResponseWriter, handler.Request, "/a/"+article.Id_.Hex(), http.StatusFound)
			return
		}
	}

	handler.renderTemplate("article/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "编辑",
		"action": "/a/" + articleId + "/edit",
		"active": "article",
	})
}

// URL: /a/{articleId}/delete
// 删除文章
func deleteArticleHandler(handler *Handler) {
	user, _ := currentUser(handler)

	vars := mux.Vars(handler.Request)
	articleId := vars["articleId"]
	if !bson.IsObjectIdHex(articleId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	c := handler.DB.C(CONTENTS)

	article := new(Article)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(articleId)}).One(&article)

	if err != nil {
		fmt.Println("deleteArticleHandler:", err.Error())
		return
	}

	if article.CanDelete(user.Username, handler.DB) {
		c.Remove(bson.M{"_id": article.Id_})

		c = handler.DB.C(COMMENTS)
		c.Remove(bson.M{"contentid": article.Id_})

		http.Redirect(handler.ResponseWriter, handler.Request, "/articles", http.StatusFound)
	}
}

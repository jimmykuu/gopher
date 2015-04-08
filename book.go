package gopher

import (
	"net/http"
	"strconv"

	"github.com/deferpanic/deferclient/deferclient"
	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2/bson"
)

// URL: /books
// 图书列表
func booksHandler(handler *Handler) {
	c := handler.DB.C(BOOKS)
	var chineseBooks []Book
	c.Find(bson.M{"language": "中文"}).All(&chineseBooks)

	var englishBooks []Book
	c.Find(bson.M{"language": "英文"}).All(&englishBooks)
	handler.renderTemplate("book/index.html", BASE, map[string]interface{}{
		"chineseBooks": chineseBooks,
		"englishBooks": englishBooks,
		"active":       "books",
	})
}

// URL: /book/{id}
// 显示图书详情
func showBookHandler(handler *Handler) {
	bookId := mux.Vars(handler.Request)["id"]

	if !bson.IsObjectIdHex(bookId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	c := handler.DB.C(BOOKS)
	var book Book
	c.Find(bson.M{"_id": bson.ObjectIdHex(bookId)}).One(&book)

	handler.renderTemplate("book/show.html", BASE, map[string]interface{}{
		"book":   book,
		"active": "books",
	})
}

// URL: /admin/book/{id}/edit
// 编辑图书
func editBookHandler(handler *Handler) {
	defer deferclient.Persist()

	bookId := mux.Vars(handler.Request)["id"]

	c := handler.DB.C(BOOKS)
	var book Book
	c.Find(bson.M{"_id": bson.ObjectIdHex(bookId)}).One(&book)

	form := wtforms.NewForm(
		wtforms.NewTextField("title", "书名", book.Title, wtforms.Required{}),
		wtforms.NewTextField("cover", "封面", book.Cover, wtforms.Required{}),
		wtforms.NewTextField("author", "作者", book.Author, wtforms.Required{}),
		wtforms.NewTextField("translator", "译者", book.Translator),
		wtforms.NewTextArea("introduction", "简介", book.Introduction),
		wtforms.NewTextField("pages", "页数", strconv.Itoa(book.Pages), wtforms.Required{}),
		wtforms.NewTextField("language", "语言", book.Language, wtforms.Required{}),
		wtforms.NewTextField("publisher", "出版社", book.Publisher),
		wtforms.NewTextField("publication_date", "出版年月日", book.PublicationDate),
		wtforms.NewTextField("isbn", "ISBN", book.ISBN),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			pages, _ := strconv.Atoi(form.Value("pages"))

			err := c.Update(bson.M{"_id": book.Id_}, bson.M{"$set": bson.M{
				"title":            form.Value("title"),
				"cover":            form.Value("cover"),
				"author":           form.Value("author"),
				"translator":       form.Value("translator"),
				"introduction":     form.Value("introduction"),
				"pages":            pages,
				"language":         form.Value("language"),
				"publisher":        form.Value("publisher"),
				"publication_date": form.Value("publication_date"),
				"isbn":             form.Value("isbn"),
			}})

			if err != nil {
				panic(err)
			}

			http.Redirect(handler.ResponseWriter, handler.Request, "/admin/books", http.StatusFound)
			return
		}
	}

	handler.renderTemplate("book/form.html", ADMIN, map[string]interface{}{
		"book":  book,
		"form":  form,
		"isNew": false,
	})
}

// URL: /book/{id}/delete
// 删除图书
func deleteBookHandler(handler *Handler) {
	id := mux.Vars(handler.Request)["id"]

	c := handler.DB.C(BOOKS)
	c.RemoveId(bson.ObjectIdHex(id))

	handler.ResponseWriter.Write([]byte("true"))
}

func listBooksHandler(handler *Handler) {
	c := handler.DB.C(BOOKS)
	var books []Book
	c.Find(nil).All(&books)

	handler.renderTemplate("book/list.html", ADMIN, map[string]interface{}{
		"books": books,
	})
}

func newBookHandler(handler *Handler) {
	defer deferclient.Persist()

	form := wtforms.NewForm(
		wtforms.NewTextField("title", "书名", "", wtforms.Required{}),
		wtforms.NewTextField("cover", "封面", "", wtforms.Required{}),
		wtforms.NewTextField("author", "作者", "", wtforms.Required{}),
		wtforms.NewTextField("translator", "译者", ""),
		wtforms.NewTextArea("introduction", "简介", ""),
		wtforms.NewTextField("pages", "页数", "", wtforms.Required{}),
		wtforms.NewTextField("language", "语言", "", wtforms.Required{}),
		wtforms.NewTextField("publisher", "出版社", ""),
		wtforms.NewTextField("publication_date", "出版年月日", ""),
		wtforms.NewTextField("isbn", "ISBN", ""),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			pages, _ := strconv.Atoi(form.Value("pages"))
			c := handler.DB.C(BOOKS)
			err := c.Insert(&Book{
				Id_:             bson.NewObjectId(),
				Title:           form.Value("title"),
				Cover:           form.Value("cover"),
				Author:          form.Value("author"),
				Translator:      form.Value("translator"),
				Pages:           pages,
				Language:        form.Value("language"),
				Publisher:       form.Value("publisher"),
				PublicationDate: form.Value("publication_date"),
				Introduction:    form.Value("introduction"),
				ISBN:            form.Value("isbn"),
			})

			if err != nil {
				panic(err)
			}
			http.Redirect(handler.ResponseWriter, handler.Request, "/admin/books", http.StatusFound)
			return
		}
	}

	handler.renderTemplate("book/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": true,
	})
}

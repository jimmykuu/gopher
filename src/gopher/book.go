package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

func booksHandler(w http.ResponseWriter, r *http.Request) {
	c := DB.C("books")
	var chineseBooks []Book
	c.Find(bson.M{"language": "中文"}).All(&chineseBooks)

	var englishBooks []Book
	c.Find(bson.M{"language": "英文"}).All(&englishBooks)
	renderTemplate(w, r, "book/index.html", map[string]interface{}{
		"chineseBooks": chineseBooks,
		"englishBooks": englishBooks,
		"active":       "books",
	})
}

func showBookHandler(w http.ResponseWriter, r *http.Request) {
	bookId := mux.Vars(r)["bookId"]

	c := DB.C("books")
	var book Book
	c.Find(bson.M{"_id": bson.ObjectIdHex(bookId)}).One(&book)

	renderTemplate(w, r, "book/show.html", map[string]interface{}{
		"book":   book,
		"active": "books",
	})
}

package gopher

import (
	"container/list"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
)

//rss contents of today

var contents []Topic

//缓存
var (
	cache     list.List
	beginTime time.Time
)

func rssHandler(w http.ResponseWriter, r *http.Request) {
	session, err := mgo.Dial(Config.DB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	now := time.Now()

	DB := session.DB("gopher")
	c := DB.C("contents")
	c.Find(nil).All(&contents)
	fmt.Println(contents)
	t, err := template.ParseFiles("templates/rss.xml")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, map[string]interface{}{
		"topics": contents,
	})

}

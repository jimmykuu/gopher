package gopher

import (
	"fmt"
	"html/template"
	"net/http"

	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
)

func rssHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("RSS")
	session, err := mgo.Dial(Config.DB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	DB := session.DB("gopher")
	c := DB.C("contents")
	var contents []Topic
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

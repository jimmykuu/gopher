package gopher

import (
	//"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 获取路由.
func getRoute() *mux.Router {
	r := mux.NewRouter()
	for _, route := range routes {
		r.HandleFunc(route.URL, handlerFun(route))
	}
	return r
}

// 测试权限拦截.
func TestPermission(t *testing.T) {
	db, _ := mgo.Dial(Config.DB)
	defer func() {
		db.DB("gopher").C(USERS).Remove(bson.M{"username": "没有中文名"})
	}()
	r := getRoute()
	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", Config.Host+"/topic/new", nil)
	if err != nil {
		panic(err)
	}
	r.ServeHTTP(res, req)
	if res.Code != http.StatusFound || res.Header().Get("Location") != "/signin" {
		t.Fatal("Autenticated permission failed.")
	}

	req, err = http.NewRequest("GET", Config.Host+"/admin/nodes", nil)
	if err != nil {
		panic(err)
	}
	db.DB("gopher").C(USERS).Insert(bson.M{"username": "没有中文名"})
	session, _ := store.Get(req, "user")
	session.Values["username"] = "没有中文名"
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	body := res.Body.String()
	if !regexp.MustCompile("对不起，你没有权限进行该操作").MatchString(body) {
		t.Fail()
	}

}

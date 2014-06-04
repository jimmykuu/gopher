/*
读取配置文件,设置URL,启动服务器
*/

package gopher

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handlerFun(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := NewHandler(w, r)
		if route.Permission == Everyone {
			route.HandlerFunc(handler)
		} else if route.Permission == Authenticated {
			_, ok := currentUser(r)

			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}

			route.HandlerFunc(handler)
		} else if route.Permission == Administrator {
			user, ok := currentUser(r)

			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}

			if !user.IsSuperuser {
				message(handler, "没有权限", "对不起，你没有权限进行该操作", "error")
				return
			}

			route.HandlerFunc(handler)
		}
	}
}

func StartServer() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	r := mux.NewRouter()
	for _, route := range routes {
		r.HandleFunc(route.URL, handlerFun(route))
	}

	http.Handle("/", r)

	fmt.Println("Server start on:", Config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil))
}

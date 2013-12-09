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

func handlerFun(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler.Permission == Everyone {
			handler.HandlerFunc(w, r)
		} else if handler.Permission == Authenticated {
			_, ok := currentUser(r)

			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}

			handler.HandlerFunc(w, r)
		} else if handler.Permission == Administrator {
			user, ok := currentUser(r)

			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}

			if !user.IsSuperuser {
				message(w, r, "没有权限", "对不起，你没有权限进行该操作", "error")
				return
			}

			handler.HandlerFunc(w, r)
		}
	}
}

func StartServer() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	r := mux.NewRouter()
	for _, handler := range handlers {
		r.HandleFunc(handler.URL, handlerFun(handler))
	}

	http.Handle("/", r)

	fmt.Println("Server start on:", Config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil))
}

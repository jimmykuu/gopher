/*
读取配置文件,设置URL,启动服务器
*/

package gopher

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.net/websocket"
	"github.com/gorilla/mux"
)

var (
	logger = log.New(os.Stdout, "[gopher]:", log.LstdFlags)
)

func handlerFun(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			defer func() {
				if e := recover(); e != nil {
					fmt.Println("panic:", e)
				}
			}()*/

		handler := NewHandler(w, r)
		defer handler.Session.Close()

		url := r.Method + " " + r.URL.Path
		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}
		logger.Println(url)
		if route.Permission&Everyone == Everyone {
			route.HandlerFunc(handler)
		}
		var (
			user *User
			ok   bool
		)
		if route.Permission&Authenticated == Authenticated {
			user, ok = currentUser(handler)
			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}

			if route.Permission&AdministratorOnly == AdministratorOnly {
				if !user.IsSuperuser {
					message(handler, "没有权限", "对不起，你没有权限进行该操作", "error")
					return
				}
			}

			route.HandlerFunc(handler)
		}
	}
}

func StartServer() {
	//http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.Handle("/get/package", websocket.Handler(getPackageHandler))

	r := mux.NewRouter()
	for _, route := range routes {
		r.HandleFunc(route.URL, handlerFun(route))
	}

	r.PathPrefix("/static/").HandlerFunc(fileHandler)
	http.Handle("/", r)

	logger.Println("Server start on:", Config.Port)
	// http server
	// err := http.ListenAndServeTLS(fmt.Sprintf(":%d", Config.Port), "cert.pem", "key.pem", nil)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil)
	if err != nil {
		logger.Fatal(err)
	}
}

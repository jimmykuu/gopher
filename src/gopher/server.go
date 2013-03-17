/*
读取配置文件,设置URL,启动服务器
*/

package gopher

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func StartServer() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	r := mux.NewRouter()
	for url, handler := range handlers {
		r.HandleFunc(url, handler)
	}

	http.Handle("/", r)

	fmt.Println("Server start on:", Config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil))
}

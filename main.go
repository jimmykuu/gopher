/*
读取配置文件,设置URL,启动服务器
*/

package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

var config map[string]string

// 初始化,读取配置文件
func init() {
	file, err := os.Open("config.json")
	if err != nil {
		println("配置文件读取失败")
		panic(err)
		os.Exit(1)
	}

	defer file.Close()

	dec := json.NewDecoder(file)

	err = dec.Decode(&config)

	if err != nil {
		println("配置文件读取失败")
		panic(err)
		os.Exit(1)
	}
}

func main() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	r := mux.NewRouter()
	for url, handler := range handlers {
		r.HandleFunc(url, handler)
	}

	http.Handle("/", r)

	port := config["port"]

	println("Listen", port)
	http.ListenAndServe(":"+port, nil)
}

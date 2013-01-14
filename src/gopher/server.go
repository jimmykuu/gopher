/*
读取配置文件,设置URL,启动服务器
*/

package gopher

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var config map[string]string
var analyticsCode template.HTML // 网站统计分析代码

// 初始化,读取配置文件
func init() {
	file, err := os.Open("etc/config.json")
	if err != nil {
		fmt.Println("配置文件读取失败:", err.Error())
		os.Exit(1)
	}

	defer file.Close()

	dec := json.NewDecoder(file)

	err = dec.Decode(&config)

	if err != nil {
		fmt.Println("配置文件解析失败:", err.Error())
		os.Exit(1)
	}

	analyticsFile := config["analytics_file"]

	if analyticsFile != "" {
		content, err := ioutil.ReadFile(analyticsFile)

		if err != nil {
			fmt.Println("统计分析文件没有找到:", err.Error())
			os.Exit(1)
		}

		analyticsCode = template.HTML(string(content))
	}
}

func StartServer() {
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	r := mux.NewRouter()
	for url, handler := range handlers {
		r.HandleFunc(url, handler)
	}

	http.Handle("/", r)

	port := config["port"]
	fmt.Println("Server start on:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

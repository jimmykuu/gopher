package gopher

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

type ConfigStruct struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	DB            string `json:"db"`
	CookieSecret  string `json:"cookie_secret"`
	SmtpUsername  string `json:"smtp_username"`
	SmtpPassword  string `json:"smtp_password"`
	SmtpHost      string `json:"smtp_host"`
	SmtpAddr      string `json:"smtp_addr"`
	FromEmail     string `json:"from_email"`
	Superusers    string `json:"superusers"`
	AnalyticsFile string `json:"analytics_file"`
}

var Config ConfigStruct
var analyticsCode template.HTML // 网站统计分析代码

func init() {
	file, err := os.Open("etc/config.json")
	if err != nil {
		fmt.Println("配置文件读取失败:", err.Error())
		os.Exit(1)
	}

	defer file.Close()

	dec := json.NewDecoder(file)

	err = dec.Decode(&Config)

	if err != nil {
		fmt.Println("配置文件解析失败:", err.Error())
		os.Exit(1)
	}

	if Config.AnalyticsFile != "" {
		content, err := ioutil.ReadFile(Config.AnalyticsFile)

		if err != nil {
			fmt.Println("统计分析文件没有找到:", err.Error())
			os.Exit(1)
		}

		analyticsCode = template.HTML(string(content))
	}
}

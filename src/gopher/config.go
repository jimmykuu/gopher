package gopher

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/bradrydzewski/go.auth"
)

type ConfigStruct struct {
	Host                       string `json:"host"`
	Port                       int    `json:"port"`
	DB                         string `json:"db"`
	CookieSecret               string `json:"cookie_secret"`
	SmtpUsername               string `json:"smtp_username"`
	SmtpPassword               string `json:"smtp_password"`
	SmtpHost                   string `json:"smtp_host"`
	SmtpAddr                   string `json:"smtp_addr"`
	FromEmail                  string `json:"from_email"`
	Superusers                 string `json:"superusers"`
	TimeZoneOffset             int64  `json:"time_zone_offset"`
	AnalyticsFile              string `json:"analytics_file"`
	ShareCodeFile              string `json:"share_code_file"`
	StaticFileVersion          int    `json:"static_file_version"`
	QiniuAccessKey             string `json:"qiniu_access_key"`
	QiniuSecretKey             string `json:"qiniu_secret_key"`
	GoGetPath                  string `json:"go_get_path"`
	PackagesDownloadPath       string `json:"packages_download_path"`
	PublicSalt                 string `json:"public_salt"`
	CookieSecure               bool   `json:"cookie_secure"`
	GithubClientId             string `json:"github_auth_client_id"`
	GithubClientSecret         string `json:"github_auth_client_secret"`
	GithubLoginRedirect        string `json:"github_login_redirect"`
	GithubLoginSuccessRedirect string `json:"github_login_success_redirect"`
}

var Config ConfigStruct
var analyticsCode template.HTML // 网站统计分析代码
var shareCode template.HTML     // 分享代码
var goVersion = runtime.Version()

func init() {
	file, err := os.Open("etc/config.json")
	if err != nil {
		log.Fatal("配置文件读取失败:", err.Error())
	}

	defer file.Close()

	dec := json.NewDecoder(file)

	err = dec.Decode(&Config)

	if err != nil {
		log.Fatal("配置文件解析失败:", err.Error())
	}

	if Config.AnalyticsFile != "" {
		content, err := ioutil.ReadFile(Config.AnalyticsFile)

		if err != nil {
			log.Fatal("统计分析文件没有找到:", err.Error())
		}

		analyticsCode = template.HTML(string(content))
	}

	if Config.ShareCodeFile != "" {
		content, err := ioutil.ReadFile(Config.ShareCodeFile)

		if err != nil {
			log.Fatal("分享代码文件没有找到:", err.Error())
		}

		shareCode = template.HTML(string(content))
	}
	if Config.GithubClientId == "" || Config.GithubClientSecret == "" {
		log.Fatal("没有配置github应用的参数")
	}
	auth.Config.CookieSecret = []byte(Config.CookieSecret)
	auth.Config.LoginRedirect = Config.GithubLoginRedirect
	auth.Config.LoginSuccessRedirect = Config.GithubLoginSuccessRedirect
	auth.Config.CookieSecure = Config.CookieSecure
	if !auth.Config.CookieSecure {
		fmt.Println("注意,cookie_secure设置为false,只能在本地环境下测试")
	}
	githubHandler = auth.Github(Config.GithubClientId, Config.GithubClientSecret, "user")
}

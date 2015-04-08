package gopher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"
	"time"

	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	BASE  = "base.html"
	ADMIN = "admin/base.html"
)

var funcMaps = template.FuncMap{
	"html": func(text string) template.HTML {
		return template.HTML(text)
	},
	"input": func(form wtforms.Form, fieldStr string) template.HTML {
		field, err := form.Field(fieldStr)
		if err != nil {
			panic(err)
		}

		errorClass := ""
		errorMessage := ""
		if field.HasErrors() {
			errorClass = "error "
			errorMessage = `<div class="ui red pointing above ui label">` + strings.Join(field.Errors(), ", ") + `</div>`
		}
		format := `<div class="%sfield">
			    %s
			    %s
				%s
		    </div>`

		return template.HTML(
			fmt.Sprintf(format,
				errorClass,
				field.RenderLabel(),
				field.RenderInput(),
				errorMessage,
			))
	},
	"loadtimes": func(startTime time.Time) string {
		// 加载时间
		return fmt.Sprintf("%dms", time.Now().Sub(startTime)/1000000)
	},
	"ads": func(position string, db *mgo.Database) []AD {
		c := db.C(ADS)
		var ads []AD
		c.Find(bson.M{"position": position}).Sort("index").All(&ads)

		return ads
	},
	"url": func(url string) string {
		// 没有http://或https://开头的增加http://
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			return url
		}

		return "http://" + url
	},
	"add": func(a, b int) int {
		// 加法运算
		return a + b
	},
	"formatdate": func(t time.Time) string {
		// 格式化日期
		return t.Format(time.RFC822)
	},
	"formattime": func(t time.Time) string {
		// 格式化时间
		now := time.Now()
		duration := now.Sub(t)
		if duration.Seconds() < 60 {
			return fmt.Sprintf("刚刚")
		} else if duration.Minutes() < 60 {
			return fmt.Sprintf("%.0f 分钟前", duration.Minutes())
		} else if duration.Hours() < 24 {
			return fmt.Sprintf("%.0f 小时前", duration.Hours())
		}

		t = t.Add(time.Hour * time.Duration(Config.TimeZoneOffset))
		return t.Format("2006-01-02 15:04")
	},
	"nl2br": func(text string) template.HTML {
		return template.HTML(strings.Replace(text, "\n", "<br>", -1))
	},
}

// 解析模板
func parseTemplate(file, baseFile string, data map[string]interface{}) []byte {
	var buf bytes.Buffer
	t := template.New(file).Funcs(funcMaps)
	baseBytes, err := ioutil.ReadFile("templates/" + baseFile)
	if err != nil {
		panic(err)
	}
	t, err = t.Parse(string(baseBytes))
	if err != nil {
		panic(err)
	}
	t, err = t.ParseFiles("templates/" + file)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// 渲染模板，并放入一些模板常用变量
func renderTemplate(handler Handler, file, baseFile string, data map[string]interface{}) {
	_, isPresent := data["signout"]

	// 如果isPresent==true，说明在执行登出操作
	if !isPresent {
		// 加入用户信息
		user, ok := currentUser(&handler)

		if ok {
			data["curr_user"] = user
			data["username"] = user.Username
			data["isSuperUser"] = user.IsSuperuser
			data["email"] = user.Email
			data["fansCount"] = len(user.Fans)
			data["followCount"] = len(user.Follow)
		}
	}

	data["utils"] = utils

	data["analyticsCode"] = analyticsCode
	data["shareCode"] = shareCode
	data["staticFileVersion"] = Config.StaticFileVersion
	data["goVersion"] = goVersion
	data["startTime"] = handler.StartTime
	data["db"] = handler.DB
	data["host"] = Config.Host

	_, ok := data["active"]
	if !ok {
		data["active"] = ""
	}

	page := parseTemplate(file, baseFile, data)
	handler.ResponseWriter.Write(page)
}

func renderJson(handler Handler, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	handler.ResponseWriter.Header().Set("Content-Type", "application/json")
	handler.ResponseWriter.Write(b)
}

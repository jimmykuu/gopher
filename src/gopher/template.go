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
)

const (
	BASE  = "base.html"
	ADMIN = "admin/base.html"
)

var funcMaps = template.FuncMap{
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
		return fmt.Sprintf("%dms", time.Now().Sub(startTime)/1000000)
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
		user, ok := currentUser(handler)

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

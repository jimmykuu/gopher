package gopher

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

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
	// t, err := template.ParseFiles("templates/"+baseFile, "templates/"+file)
	// if err != nil {
	// 	panic(err)
	// }
	err = t.Execute(&buf, data)

	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// 渲染模板，并放入一些模板常用变量
func renderTemplate(w http.ResponseWriter, r *http.Request, file, baseFile string, data map[string]interface{}) {
	_, isPresent := data["signout"]

	// 如果isPresent==true，说明在执行登出操作
	if !isPresent {
		// 加入用户信息
		user, ok := currentUser(r)

		if ok {
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

	_, ok := data["active"]
	if !ok {
		data["active"] = ""
	}

	page := parseTemplate(file, baseFile, data)
	w.Write(page)
}

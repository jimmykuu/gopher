/*
包含请求上下文和处理函数的结构体.
*/
package gopher

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Handler 是包含一些请求上下文的结构体.
type Handler struct {
	http.ResponseWriter
	*http.Request
	*mgo.Session               //会话
	StartTime    time.Time     //接受请求时间
	DB           *mgo.Database //数据库
}

// 只用file作模板的简易渲染
func (handler *Handler) render(file string, datas ...map[string]interface{}) {
	var data = make(map[string]interface{})
	if len(datas) == 1 {
		data = datas[0]
	} else if len(datas) != 0 {
		panic("不能传入超过多个data map")
	}
	tpl, err := template.ParseFiles(file)
	if err != nil {
		panic(err)
	}
	tpl.Execute(handler.ResponseWriter, data)
}

// 渲染模板，并放入一些模板常用变量
func (handler *Handler) renderTemplate(file, baseFile string, datas ...map[string]interface{}) {
	var data = make(map[string]interface{})
	if len(datas) == 1 {
		data = datas[0]
	} else if len(datas) != 0 {
		panic("不能传入超过多个data map")
	}
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
	data["now"] = time.Now()

	var linksOnBottom []LinkExchange
	c := handler.DB.C(LINK_EXCHANGES)
	c.Find(bson.M{"is_on_bottom": true}).All(&linksOnBottom)

	data["linksOnBottom"] = linksOnBottom

	_, ok := data["active"]
	if !ok {
		data["active"] = ""
	}

	page := parseTemplate(file, baseFile, data)
	handler.ResponseWriter.Write(page)
}

// param 返回在url中name的值.
func (handler *Handler) param(name string) string {
	return mux.Vars(handler.Request)[name]
}

// 重定向.
func (handler *Handler) redirect(urlStr string, code int) {
	http.Redirect(handler.ResponseWriter, handler.Request, urlStr, code)
}

// 返回json数据.
func (handler *Handler) renderJson(data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	handler.ResponseWriter.Header().Set("Content-Type", "application/json")
	handler.ResponseWriter.Write(b)
}

// 返回文本数据
func (handler *Handler) renderText(text string) {
	handler.ResponseWriter.Write([]byte(text))
}

func (handler *Handler) notFound() {
	http.NotFound(handler.ResponseWriter, handler.Request)
}

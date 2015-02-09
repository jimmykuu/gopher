/*
包含请求上下文和处理函数的结构体.
*/
package gopher

import (
	"encoding/json"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
)

// Handler 是包含一些请求上下文的结构体.
type Handler struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	StartTime      time.Time     //接受请求时间
	Session        *mgo.Session  //会话
	DB             *mgo.Database //数据库
}

// 渲染模板，并放入一些模板常用变量
func (handler *Handler) renderTemplate(file, baseFile string, datas ...map[string]interface{}) {
	var data map[string]interface{}
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

	_, ok := data["active"]
	if !ok {
		data["active"] = ""
	}

	page := parseTemplate(file, baseFile, data)
	handler.ResponseWriter.Write(page)
}

func (handler *Handler) renderJson(data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	handler.ResponseWriter.Header().Set("Content-Type", "application/json")
	handler.ResponseWriter.Write(b)
}

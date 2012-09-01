/*
主题
*/

package main

import (
	"./wtforms"
	"code.google.com/p/gorilla/mux"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// URL: /
// 网站首页,列出按回帖时间倒序排列的第一页
func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("p")
	page := 1

	if p != "" {
		var err error
		page, err = strconv.Atoi(p)

		if err != nil {
			message(w, r, "页码错误", "页码错误", "error")
			return
		}
	}

	var hotNodes []Node
	c := db.C("nodes")
	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").Limit(10).All(&hotNodes)

	var status Status
	c = db.C("status")
	c.Find(nil).One(&status)

	c = db.C("topics")

	pagination := NewPagination(c.Find(nil).Sort("-latestrepliedat"), "/", PerPage)

	var topics []Topic

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&topics)

	renderTemplate(w, r, "index.html", map[string]interface{}{"nodes": hotNodes, "status": status, "topics": topics, "pagination": pagination, "page": page})
}

// URL: /topic/new
// 新建主题
func newTopicHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := currentUser(r); !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	nodeId := vars["node"]

	var nodes []Node
	c := db.C("nodes")
	c.Find(nil).All(&nodes)

	var choices []wtforms.Choice

	for _, node := range nodes {
		choices = append(choices, wtforms.Choice{Value: node.Id, Label: node.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewSelectField("node", "节点", choices, nodeId),
		wtforms.NewTextArea("title", "标题", "", &wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", ""),
	)

	if r.Method == "POST" {
		if form.Validate(r) {
			nodeId = form.Value("node")

			node := Node{}
			c := db.C("nodes")
			c.Find(bson.M{"id": nodeId}).One(&node)

			session, _ := store.Get(r, "user")
			username, _ := session.Values["username"]
			username = username.(string)

			user := User{}
			c = db.C("users")
			c.Find(bson.M{"username": username}).One(&user)

			c = db.C("topics")

			Id_ := bson.NewObjectId()

			now := time.Now()

			html := form.Value("html")
			html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

			err := c.Insert(&Topic{
				Id_:             Id_,
				NodeId:          node.Id_,
				UserId:          user.Id_,
				Title:           form.Value("title"),
				Markdown:        form.Value("content"),
				Html:            template.HTML(html),
				CreatedAt:       now,
				LatestRepliedAt: now})

			if err != nil {
				panic(err)
			}

			c = db.C("status")
			var status Status
			c.Find(nil).One(&status)

			c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"topiccount": 1}})

			http.Redirect(w, r, "/t/"+Id_.Hex(), http.StatusFound)
		}
	}

	renderTemplate(w, r, "topic/new.html", map[string]interface{}{"form": form})
}

// URL: /t/{topicId}
// 根据主题的ID,显示主题的信息及回复
func showTopicHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId := vars["topicId"]
	c := db.C("topics")

	topic := Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId)}).One(&topic)

	if err != nil {
		println("err")
	}

	c = db.C("topics")
	c.UpdateId(bson.ObjectIdHex(topicId), bson.M{"$inc": bson.M{"hits": 1}})

	renderTemplate(w, r, "topic/show.html", map[string]interface{}{"topic": topic})
}

// URL: /reply/{topicId}
// 回复主题
func replyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		vars := mux.Vars(r)
		topicId := vars["topicId"]

		user, ok := currentUser(r)

		if !ok {
			http.Redirect(w, r, "/t/"+topicId, http.StatusFound)
			return
		}

		content := r.FormValue("content")

		html := r.FormValue("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

		Id_ := bson.NewObjectId()
		now := time.Now()
		reply := Reply{
			Id_:       Id_,
			UserId:    user.Id_,
			TopicId:   bson.ObjectIdHex(topicId),
			Markdown:  content,
			Html:      template.HTML(html),
			CreatedAt: now,
		}
		c := db.C("replies")
		c.Insert(&reply)

		c = db.C("topics")
		c.Update(bson.M{"_id": bson.ObjectIdHex(topicId)}, bson.M{"$inc": bson.M{"replycount": 1}, "$set": bson.M{"latestreplyid": Id_.Hex(), "latestrepliedat": now}})

		c = db.C("status")
		var status Status
		c.Find(nil).One(&status)

		c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"replycount": 1}})

		http.Redirect(w, r, "/t/"+topicId, http.StatusFound)
	}
}

// URL: /go/{node}
// 列出节点下所有的主题
func topicInNodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeId := vars["node"]
	c := db.C("nodes")

	node := Node{}
	err := c.Find(bson.M{"id": nodeId}).One(&node)

	if err != nil {
		message(w, r, "没有此节点", "请联系管理员创建此节点", "error")
		return
	}

	p := r.FormValue("p")
	page := 1

	if p != "" {
		var err error
		page, err = strconv.Atoi(p)

		if err != nil {
			message(w, r, "页码错误", "页码错误", "error")
			return
		}
	}

	c = db.C("topics")

	pagination := NewPagination(c.Find(bson.M{"nodeid": node.Id_}).Sort("-latestrepliedat"), "/", 20)

	var topics []Topic

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "没有找到页面", "没有找到页面", "error")
		return
	}

	query.All(&topics)

	renderTemplate(w, r, "/topic/list.html", map[string]interface{}{"topics": topics, "node": node})
}

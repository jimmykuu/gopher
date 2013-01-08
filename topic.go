/*
主题
*/

package main

import (
	"./wtforms"
	"github.com/gorilla/mux"
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

	nodeId := mux.Vars(r)["node"]

	var nodes []Node
	c := db.C("nodes")
	c.Find(nil).All(&nodes)

	var choices []wtforms.Choice

	for _, node := range nodes {
		choices = append(choices, wtforms.Choice{Value: node.Id_.Hex(), Label: node.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewSelectField("node", "节点", choices, nodeId),
		wtforms.NewTextArea("title", "标题", "", &wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", ""),
	)

	if r.Method == "POST" && form.Validate(r) {
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

		nodeId := bson.ObjectIdHex(form.Value("node"))
		err := c.Insert(&Topic{
			Id_:             Id_,
			NodeId:          nodeId,
			UserId:          user.Id_,
			Title:           form.Value("title"),
			Markdown:        form.Value("content"),
			Html:            template.HTML(html),
			CreatedAt:       now,
			LatestRepliedAt: now,
		})

		if err != nil {
			panic(err)
		}

		// 增加Node.TopicCount
		c = db.C("nodes")
		c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})

		c = db.C("status")
		var status Status
		c.Find(nil).One(&status)

		c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"topiccount": 1}})

		http.Redirect(w, r, "/t/"+Id_.Hex(), http.StatusFound)
		return
	}

	renderTemplate(w, r, "topic/form.html", map[string]interface{}{"form": form, "title": "新建", "action": "/topic/new"})
}

// URL: /t/{topicId}/edit
// 编辑主题
func editTopicHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	topicId := mux.Vars(r)["topicId"]

	c := db.C("topics")
	var topic Topic
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId)}).One(&topic)

	if err != nil {
		message(w, r, "没有该主题", "没有该主题,不能编辑", "error")
		return
	}

	if !topic.CanEdit(user.Username) {
		message(w, r, "没用该权限", "对不起,你没有权限编辑该主题", "error")
		return
	}

	var nodes []Node
	c = db.C("nodes")
	c.Find(nil).All(&nodes)

	var choices []wtforms.Choice

	for _, node := range nodes {
		choices = append(choices, wtforms.Choice{Value: node.Id_.Hex(), Label: node.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewSelectField("node", "节点", choices, topic.NodeId.Hex()),
		wtforms.NewTextArea("title", "标题", topic.Title, &wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", topic.Markdown),
	)

	content := topic.Markdown
	html := topic.Html

	if r.Method == "POST" {
		if form.Validate(r) {
			html := form.Value("html")
			html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

			nodeId := bson.ObjectIdHex(form.Value("node"))
			c = db.C("topics")
			c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{
				"nodeid":    nodeId,
				"title":     form.Value("title"),
				"markdown":  form.Value("content"),
				"html":      template.HTML(html),
				"updatedat": time.Now(),
				"updatedby": user.Id_,
			}})

			// 如果两次的节点不同,更新节点的主题数量

			if topic.NodeId != nodeId {
				c = db.C("nodes")
				c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})
				c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})
			}

			http.Redirect(w, r, "/t/"+topic.Id_.Hex(), http.StatusFound)
			return
		}

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
	}

	renderTemplate(w, r, "topic/form.html", map[string]interface{}{"form": form, "title": "编辑", "action": "/t/" + topicId + "/edit", "html": html, "content": content})
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

// URL: /reply/{replyId}/delete
// 删除回复，只有管理员可以删除
func deleteReplyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	replyId := vars["replyId"]

	user, ok := currentUser(r)

	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没用该权限", "对不起,你没有权限删除该回复", "error")
		return
	}

	c := db.C("replies")
	var reply Reply
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(replyId)}).One(&reply)
	if err != nil {
		message(w, r, "该回复不存在", "不存在该回复", "error")
		return
	}

	err = c.Remove(bson.M{"_id": bson.ObjectIdHex(replyId)})

	if err != nil {
		message(w, r, "该回复不存在", "不存在该回复", "error")
		return
	}

	c = db.C("topics")
	// 减少该主题的回复数量
	c.Update(bson.M{"_id": reply.TopicId}, bson.M{"$inc": bson.M{"replycount": -1}})

	var topic Topic
	c.Find(bson.M{"_id": reply.TopicId}).One(&topic)

	if topic.LatestReplyId == replyId {
		if topic.ReplyCount == 0 {
			// 如果删除后没有回复，设置最后回复id为空，最后回复时间为创建时间
			c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{"latestreplyid": "", "latestrepliedat": reply.CreatedAt}})
		} else {
			// 如果删除的是该主题最后一个回复，设置主题的最新回复id，和时间
			var latestReply Reply
			c = db.C("replies")
			c.Find(bson.M{"topicid": topic.Id_}).Sort("-createdat").Limit(1).One(&latestReply)

			c = db.C("topics")
			c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{"latestreplyid": latestReply.Id_.Hex(), "latestrepliedat": latestReply.CreatedAt}})
		}
	}

	c = db.C("status")
	var status Status
	c.Find(nil).One(&status)
	c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"replycount": -1}})

	http.Redirect(w, r, "/t/"+reply.TopicId.Hex(), http.StatusFound)
}

/*
主题
*/

package gopher

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"html/template"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

func topicsHandler(w http.ResponseWriter, r *http.Request, conditions bson.M, sort string, url string, subActive string) {
	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	var hotNodes []Node
	c := DB.C("nodes")
	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").Limit(10).All(&hotNodes)

	var status Status
	c = DB.C("status")
	c.Find(nil).One(&status)

	c = DB.C("contents")

	pagination := NewPagination(c.Find(conditions).Sort(sort), url, PerPage)

	var topics []Topic

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	query.All(&topics)

	renderTemplate(w, r, "index.html", map[string]interface{}{
		"nodes":      hotNodes,
		"status":     status,
		"topics":     topics,
		"pagination": pagination,
		"page":       page,
		"active":     "topic",
		"subActive":  subActive,
	})
}

// URL: /
// 网站首页,列出按回帖时间倒序排列的第一页
func indexHandler(w http.ResponseWriter, r *http.Request) {
	topicsHandler(w, r, bson.M{"content.type": TypeTopic}, "-latestrepliedat", "/", "latestReply")
}

// URL: /topics/latest
// 最新发布的主题，按照发布时间倒序排列
func latestTopicsHandler(w http.ResponseWriter, r *http.Request) {
	topicsHandler(w, r, bson.M{"content.type": TypeTopic}, "-content.createdat", "/topics/latest", "latestCreate")
}

// URL: /topics/no_reply
// 无人回复的主题，按照发布时间倒序排列
func noReplyTopicsHandler(w http.ResponseWriter, r *http.Request) {
	topicsHandler(w, r, bson.M{"content.type": TypeTopic, "content.commentcount": 0}, "-content.createdat", "/topics/no_reply", "noReply")
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
	c := DB.C("nodes")
	c.Find(nil).All(&nodes)

	var choices = []wtforms.Choice{wtforms.Choice{}} // 第一个选项为空

	for _, node := range nodes {
		choices = append(choices, wtforms.Choice{Value: node.Id_.Hex(), Label: node.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewSelectField("node", "节点", choices, nodeId, &wtforms.Required{}),
		wtforms.NewTextArea("title", "标题", "", &wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", ""),
	)

	var content string
	var html template.HTML

	if r.Method == "POST" {
		if form.Validate(r) {
			session, _ := store.Get(r, "user")
			username, _ := session.Values["username"]
			username = username.(string)

			user := User{}
			c = DB.C("users")
			c.Find(bson.M{"username": username}).One(&user)

			c = DB.C("contents")

			id_ := bson.NewObjectId()

			now := time.Now()

			html2 := form.Value("html")
			html2 = strings.Replace(html2, "<pre>", `<pre class="prettyprint linenums">`, -1)

			nodeId := bson.ObjectIdHex(form.Value("node"))
			err := c.Insert(&Topic{
				Content: Content{
					Id_:       id_,
					Type:      TypeTopic,
					Title:     form.Value("title"),
					Markdown:  form.Value("content"),
					Html:      template.HTML(html2),
					CreatedBy: user.Id_,
					CreatedAt: now,
				},
				Id_:             id_,
				NodeId:          nodeId,
				LatestRepliedAt: now,
			})

			if err != nil {
				fmt.Println("newTopicHandler:", err.Error())
				return
			}

			// 增加Node.TopicCount
			c = DB.C("nodes")
			c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})

			c = DB.C("status")
			var status Status
			c.Find(nil).One(&status)

			c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"topiccount": 1}})

			http.Redirect(w, r, "/t/"+id_.Hex(), http.StatusFound)
			return
		}

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
		form.SetValue("html", "")
	}

	renderTemplate(w, r, "topic/form.html", map[string]interface{}{
		"form":    form,
		"title":   "新建",
		"html":    html,
		"content": content,
		"action":  "/topic/new",
		"active":  "topic",
	})
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

	c := DB.C("contents")
	var topic Topic
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": TypeTopic}).One(&topic)

	if err != nil {
		message(w, r, "没有该主题", "没有该主题,不能编辑", "error")
		return
	}

	if !topic.CanEdit(user.Username) {
		message(w, r, "没用该权限", "对不起,你没有权限编辑该主题", "error")
		return
	}

	var nodes []Node
	c = DB.C("nodes")
	c.Find(nil).All(&nodes)

	var choices = []wtforms.Choice{wtforms.Choice{}} // 第一个选项为空

	for _, node := range nodes {
		choices = append(choices, wtforms.Choice{Value: node.Id_.Hex(), Label: node.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewSelectField("node", "节点", choices, topic.NodeId.Hex(), &wtforms.Required{}),
		wtforms.NewTextArea("title", "标题", topic.Title, &wtforms.Required{}),
		wtforms.NewTextArea("content", "内容", topic.Markdown),
	)

	content := topic.Markdown
	html := topic.Html

	if r.Method == "POST" {
		if form.Validate(r) {
			html2 := form.Value("html")
			html2 = strings.Replace(html2, "<pre>", `<pre class="prettyprint linenums">`, -1)

			nodeId := bson.ObjectIdHex(form.Value("node"))
			c = DB.C("contents")
			c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{
				"nodeid":            nodeId,
				"content.title":     form.Value("title"),
				"content.markdown":  form.Value("content"),
				"content.html":      template.HTML(html2),
				"content.updatedat": time.Now(),
				"content.updatedby": user.Id_.Hex(),
			}})

			// 如果两次的节点不同,更新节点的主题数量
			if topic.NodeId != nodeId {
				c = DB.C("nodes")
				c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})
				c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})
			}

			http.Redirect(w, r, "/t/"+topic.Id_.Hex(), http.StatusFound)
			return
		}

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
		form.SetValue("html", "")
	}

	renderTemplate(w, r, "topic/form.html", map[string]interface{}{
		"form":    form,
		"title":   "编辑",
		"action":  "/t/" + topicId + "/edit",
		"html":    html,
		"content": content,
		"active":  "topic",
	})
}

// URL: /t/{topicId}
// 根据主题的ID,显示主题的信息及回复
func showTopicHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId := vars["topicId"]
	c := DB.C("contents")

	topic := Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": TypeTopic}).One(&topic)

	if err != nil {
		fmt.Println("showTopicHandler:", err.Error())
		return
	}

	c.UpdateId(bson.ObjectIdHex(topicId), bson.M{"$inc": bson.M{"content.hits": 1}})

	renderTemplate(w, r, "topic/show.html", map[string]interface{}{
		"topic":  topic,
		"active": "topic",
	})
}

// URL: /go/{node}
// 列出节点下所有的主题
func topicInNodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeId := vars["node"]
	c := DB.C("nodes")

	node := Node{}
	err := c.Find(bson.M{"id": nodeId}).One(&node)

	if err != nil {
		message(w, r, "没有此节点", "请联系管理员创建此节点", "error")
		return
	}

	page, err := getPage(r)

	if err != nil {
		message(w, r, "页码错误", "页码错误", "error")
		return
	}

	c = DB.C("contents")

	pagination := NewPagination(c.Find(bson.M{"nodeid": node.Id_, "content.type": TypeTopic}).Sort("-latestrepliedat"), "/", 20)

	var topics []Topic

	query, err := pagination.Page(page)
	if err != nil {
		message(w, r, "没有找到页面", "没有找到页面", "error")
		return
	}

	query.All(&topics)

	renderTemplate(w, r, "/topic/list.html", map[string]interface{}{
		"topics": topics,
		"node":   node,
		"active": "topic",
	})
}

// URL: /t/{topicId}/delete
// 删除主题
func deleteTopicHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)

	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	if !user.IsSuperuser {
		message(w, r, "没有该权限", "对不起,你没有权限删除该评论", "error")
		return
	}

	vars := mux.Vars(r)
	topicId := bson.ObjectIdHex(vars["topicId"])
	c := DB.C("contents")

	topic := Topic{}

	err := c.Find(bson.M{"_id": topicId, "content.type": TypeTopic}).One(&topic)

	if err != nil {
		fmt.Println("deleteTopic:", err.Error())
		return
	}

	// Node统计数减一
	c = DB.C("nodes")
	c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})

	c = DB.C("status")
	var status Status
	c.Find(nil).One(&status)
	// 统计的主题数减一
	c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"topiccount": -1}})
	// 减去统计的回复数减去该主题的回复数
	c.Update(bson.M{"_id": status.Id_}, bson.M{"$inc": bson.M{"replycount": -topic.CommentCount}})

	//删除评论
	c = DB.C("comments")
	if topic.CommentCount > 0 {
		c.Remove(bson.M{"contentid": topic.Id_})
	}

	// 删除Topic记录
	c = DB.C("topics")
	c.Remove(bson.M{"_id": topicId})

	http.Redirect(w, r, "/", http.StatusFound)
}

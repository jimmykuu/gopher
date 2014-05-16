/*
主题
*/

package gopher

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func topicsHandler(w http.ResponseWriter, r *http.Request, conditions bson.M, sort string, url string, subActive string) {
	session, err := mgo.Dial(Config.DB)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	DB := session.DB("gopher")
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

	var linkExchanges []LinkExchange
	c = DB.C("link_exchanges")
	c.Find(nil).All(&linkExchanges)

	renderTemplate(w, r, "index.html", BASE, map[string]interface{}{
		"nodes":         hotNodes,
		"status":        status,
		"topics":        topics,
		"linkExchanges": linkExchanges,
		"pagination":    pagination,
		"page":          page,
		"active":        "topic",
		"subActive":     subActive,
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

			c.Update(nil, bson.M{"$inc": bson.M{"topiccount": 1}})

			http.Redirect(w, r, "/t/"+id_.Hex(), http.StatusFound)
			return
		}

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
		form.SetValue("html", "")
	}

	renderTemplate(w, r, "topic/form.html", BASE, map[string]interface{}{
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
	user, _ := currentUser(r)

	topicId := mux.Vars(r)["topicId"]

	c := DB.C("contents")
	var topic Topic
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": TypeTopic}).One(&topic)

	if err != nil {
		message(w, r, "没有该主题", "没有该主题,不能编辑", "error")
		return
	}

	if !topic.CanEdit(user.Username) {
		message(w, r, "没有该权限", "对不起,你没有权限编辑该主题", "error")
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

	renderTemplate(w, r, "topic/form.html", BASE, map[string]interface{}{
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
	cusers := DB.C("users")
	topic := Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": TypeTopic}).One(&topic)

	if err != nil {
		fmt.Println("showTopicHandler:", err.Error())
		return
	}

	c.UpdateId(bson.ObjectIdHex(topicId), bson.M{"$inc": bson.M{"content.hits": 1}})

	user, has := currentUser(r)
	if has {
		replies := user.RecentReplies
		ats := user.RecentAts
		pos := -1
		repliesDisactive := map[int]bool{}
		for k, v := range replies {
			if v == topicId {
				pos = k
				repliesDisactive[k] = true
			}
		}
		if pos != -1 {
			for pos, _ := range repliesDisactive {
				if pos == len(replies)-1 {
					replies = replies[:pos]
				} else {
					replies = append(replies[:pos], replies[pos+1:]...)
				}
			}
			cusers.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"recentreplies": replies}})
		}
		pos = -1
		atsDisactive := map[int]bool{}
		for k, v := range ats {
			if v == topicId {
				pos = k
				atsDisactive[pos] = true
			}
		}
		if pos != -1 {
			for pos, _ := range atsDisactive {
				if pos == len(ats)-1 {
					ats = ats[:pos]
				} else {
					ats = append(ats[:pos], ats[pos+1:]...)
				}
			}
			cusers.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"recentats": ats}})
		}
	}
	renderTemplate(w, r, "topic/show.html", BASE, map[string]interface{}{
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

	renderTemplate(w, r, "/topic/list.html", BASE, map[string]interface{}{
		"topics": topics,
		"node":   node,
		"active": "topic",
	})
}

// URL: /t/{topicId}/delete
// 删除主题
func deleteTopicHandler(w http.ResponseWriter, r *http.Request) {
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
	// 统计的主题数减一，减去统计的回复数减去该主题的回复数
	c.Update(nil, bson.M{"$inc": bson.M{"topiccount": -1, "replycount": -topic.CommentCount}})

	//删除评论
	c = DB.C("comments")
	if topic.CommentCount > 0 {
		c.Remove(bson.M{"contentid": topic.Id_})
	}

	// 删除Topic记录
	c = DB.C("contents")
	c.Remove(bson.M{"_id": topicId})

	http.Redirect(w, r, "/", http.StatusFound)
}

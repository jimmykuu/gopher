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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func topicsHandler(handler Handler, conditions bson.M, sort string, url string, subActive string) {
	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	var hotNodes []Node
	c := handler.DB.C(NODES)
	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").Limit(10).All(&hotNodes)

	var status Status
	c = handler.DB.C(STATUS)
	c.Find(nil).One(&status)

	c = handler.DB.C(CONTENTS)

	var topTopics []Topic

	if page == 1 {
		c.Find(bson.M{"is_top": true}).Sort(sort).All(&topTopics)

		var objectIds []bson.ObjectId
		for _, topic := range topTopics {
			objectIds = append(objectIds, topic.Id_)
		}
		if len(topTopics) > 0 {
			conditions["_id"] = bson.M{"$not": bson.M{"$in": objectIds}}
		}
	}

	pagination := NewPagination(c.Find(conditions).Sort(sort), url, PerPage)

	var topics []Topic

	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	query.(*mgo.Query).All(&topics)

	var linkExchanges []LinkExchange
	c = handler.DB.C(LINK_EXCHANGES)
	c.Find(nil).All(&linkExchanges)

	topics = append(topTopics, topics...)

	renderTemplate(handler, "index.html", BASE, map[string]interface{}{
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
func indexHandler(handler Handler) {
	topicsHandler(handler, bson.M{"content.type": TypeTopic}, "-latestrepliedat", "/", "latestReply")
}

// URL: /topics/latest
// 最新发布的主题，按照发布时间倒序排列
func latestTopicsHandler(handler Handler) {
	topicsHandler(handler, bson.M{"content.type": TypeTopic}, "-content.createdat", "/topics/latest", "latestCreate")
}

// URL: /topics/no_reply
// 无人回复的主题，按照发布时间倒序排列
func noReplyTopicsHandler(handler Handler) {
	topicsHandler(handler, bson.M{"content.type": TypeTopic, "content.commentcount": 0}, "-content.createdat", "/topics/no_reply", "noReply")
}

// URL: /topic/new
// 新建主题
func newTopicHandler(handler Handler) {
	nodeId := mux.Vars(handler.Request)["node"]

	var nodes []Node
	c := handler.DB.C(NODES)
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

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			session, _ := store.Get(handler.Request, "user")
			username, _ := session.Values["username"]
			username = username.(string)

			user := User{}
			c = handler.DB.C(USERS)
			c.Find(bson.M{"username": username}).One(&user)

			c = handler.DB.C(CONTENTS)

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
			c = handler.DB.C(NODES)
			c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})

			c = handler.DB.C(STATUS)

			c.Update(nil, bson.M{"$inc": bson.M{"topiccount": 1}})

			http.Redirect(handler.ResponseWriter, handler.Request, "/t/"+id_.Hex(), http.StatusFound)
			return
		}

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
		form.SetValue("html", "")
	}

	renderTemplate(handler, "topic/form.html", BASE, map[string]interface{}{
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
func editTopicHandler(handler Handler) {
	user, _ := currentUser(handler)

	topicId := bson.ObjectIdHex(mux.Vars(handler.Request)["topicId"])

	c := handler.DB.C(CONTENTS)
	var topic Topic
	err := c.Find(bson.M{"_id": topicId, "content.type": TypeTopic}).One(&topic)

	if err != nil {
		message(handler, "没有该主题", "没有该主题,不能编辑", "error")
		return
	}

	if !topic.CanEdit(user.Username, handler.DB) {
		message(handler, "没有该权限", "对不起,你没有权限编辑该主题", "error")
		return
	}

	var nodes []Node
	c = handler.DB.C(NODES)
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

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			html2 := form.Value("html")
			html2 = strings.Replace(html2, "<pre>", `<pre class="prettyprint linenums">`, -1)

			nodeId := bson.ObjectIdHex(form.Value("node"))
			c = handler.DB.C(CONTENTS)
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
				c = handler.DB.C(NODES)
				c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})
				c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})
			}

			http.Redirect(handler.ResponseWriter, handler.Request, "/t/"+topic.Id_.Hex(), http.StatusFound)
			return
		}

		content = form.Value("content")
		html = template.HTML(form.Value("html"))
		form.SetValue("html", "")
	}

	renderTemplate(handler, "topic/form.html", BASE, map[string]interface{}{
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
func showTopicHandler(handler Handler) {
	vars := mux.Vars(handler.Request)
	topicId := vars["topicId"]
	c := handler.DB.C(CONTENTS)
	cusers := handler.DB.C(USERS)
	topic := Topic{}

	if !bson.IsObjectIdHex(topicId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": TypeTopic}).One(&topic)

	if err != nil {
		panic(err)
		return
	}

	c.UpdateId(bson.ObjectIdHex(topicId), bson.M{"$inc": bson.M{"content.hits": 1}})

	user, has := currentUser(handler)

	//去除新消息的提醒
	if has {
		replies := user.RecentReplies
		ats := user.RecentAts
		pos := -1

		for k, v := range replies {
			if v.ContentId == topicId {
				pos = k
				break
			}
		}

		//数组的删除不是这么删的,早知如此就应该存链表了

		if pos != -1 {
			if pos == len(replies)-1 {
				replies = replies[:pos]
			} else {
				replies = append(replies[:pos], replies[pos+1:]...)
			}
			cusers.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"recentreplies": replies}})

		}

		pos = -1

		for k, v := range ats {
			if v.ContentId == topicId {
				pos = k
				break
			}
		}

		if pos != -1 {
			if pos == len(ats)-1 {
				ats = ats[:pos]
			} else {
				ats = append(ats[:pos], ats[pos+1:]...)
			}

			cusers.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"recentats": ats}})
		}
	}

	renderTemplate(handler, "topic/show.html", BASE, map[string]interface{}{
		"topic":  topic,
		"active": "topic",
	})
}

// URL: /go/{node}
// 列出节点下所有的主题
func topicInNodeHandler(handler Handler) {
	vars := mux.Vars(handler.Request)
	nodeId := vars["node"]
	c := handler.DB.C(NODES)

	node := Node{}
	err := c.Find(bson.M{"id": nodeId}).One(&node)

	if err != nil {
		message(handler, "没有此节点", "请联系管理员创建此节点", "error")
		return
	}

	page, err := getPage(handler.Request)

	if err != nil {
		message(handler, "页码错误", "页码错误", "error")
		return
	}

	c = handler.DB.C(CONTENTS)

	pagination := NewPagination(c.Find(bson.M{"nodeid": node.Id_, "content.type": TypeTopic}).Sort("-latestrepliedat"), "/", 20)

	var topics []Topic

	query, err := pagination.Page(page)
	if err != nil {
		message(handler, "没有找到页面", "没有找到页面", "error")
		return
	}

	query.(*mgo.Query).All(&topics)

	renderTemplate(handler, "/topic/list.html", BASE, map[string]interface{}{
		"topics": topics,
		"node":   node,
		"active": "topic",
	})
}

// URL: /t/{topicId}/collect/
// 将主题收藏至当前用户的收藏夹
func collectTopicHandler(handler Handler) {
	vars := mux.Vars(handler.Request)
	topicId := vars["topicId"]
	t := time.Now()
	user, _ := currentUser(handler)
	for _, v := range user.TopicsCollected {
		if v.TopicId == topicId {
			return
		}
	}
	user.TopicsCollected = append(user.TopicsCollected, CollectTopic{topicId, t})
	c := handler.DB.C(USERS)
	c.UpdateId(user.Id_, bson.M{"$set": bson.M{"topicscollected": user.TopicsCollected}})
	http.Redirect(handler.ResponseWriter, handler.Request, "/member/"+user.Username+"/collect?p=1", http.StatusFound)
}

// URL: /t/{topicId}/delete
// 删除主题
func deleteTopicHandler(handler Handler) {
	vars := mux.Vars(handler.Request)
	topicId := bson.ObjectIdHex(vars["topicId"])

	c := handler.DB.C(CONTENTS)

	topic := Topic{}

	err := c.Find(bson.M{"_id": topicId, "content.type": TypeTopic}).One(&topic)

	if err != nil {
		fmt.Println("deleteTopic:", err.Error())
		return
	}

	// Node统计数减一
	c = handler.DB.C(NODES)
	c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})

	c = handler.DB.C(STATUS)
	// 统计的主题数减一，减去统计的回复数减去该主题的回复数
	c.Update(nil, bson.M{"$inc": bson.M{"topiccount": -1, "replycount": -topic.CommentCount}})

	//删除评论
	c = handler.DB.C(COMMENTS)
	if topic.CommentCount > 0 {
		c.Remove(bson.M{"contentid": topic.Id_})
	}

	// 删除Topic记录
	c = handler.DB.C(CONTENTS)
	c.Remove(bson.M{"_id": topic.Id_})

	http.Redirect(handler.ResponseWriter, handler.Request, "/", http.StatusFound)
}

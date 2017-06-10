/*
主题
*/

package actions

import (
	"sort"
	"strconv"

	"github.com/tango-contrib/renders"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

const PerPage = 20

type City struct {
	Name        string `bson:"_id"`
	MemberCount int    `bson:"count"`
}

type ByCount []City

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].MemberCount > a[j].MemberCount }

type Topic struct {
	RenderBase
}

// 按条件列出所有主题
func (a *Topic) list(conditions bson.M, sortBy string, url string, subActive string) error {
	pageStr, err := a.Forms().String("p")
	if err != nil {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	if page <= 0 {
		page = 1
	}

	var nodes []models.Node

	session, DB := models.GetSessionAndDB()
	defer session.Close()

	c := DB.C(models.NODES)
	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").All(&nodes)

	var status models.Status
	c = DB.C(models.STATUS)
	c.Find(nil).One(&status)

	c = DB.C(models.CONTENTS)

	var topTopics []models.Topic

	if page == 1 {
		c.Find(bson.M{"is_top": true}).Sort(sortBy).All(&topTopics)

		var objectIds []bson.ObjectId
		for _, topic := range topTopics {
			objectIds = append(objectIds, topic.Id_)
		}
		if len(topTopics) > 0 {
			conditions["_id"] = bson.M{"$not": bson.M{"$in": objectIds}}
		}
	}

	pagination := NewPagination(c.Find(conditions).Sort(sortBy), url, PerPage)

	var topics []models.Topic

	query, err := pagination.Page(page)
	if err != nil {
		return err
	}

	query.(*mgo.Query).All(&topics)

	var linkExchanges []models.LinkExchange
	c = DB.C(models.LINK_EXCHANGES)
	c.Find(bson.M{"is_on_home": true}).All(&linkExchanges)

	topics = append(topTopics, topics...)

	c = DB.C(models.USERS)

	var cities []City
	c.Pipe([]bson.M{bson.M{
		"$group": bson.M{
			"_id":   "$location",
			"count": bson.M{"$sum": 1},
		},
	}}).All(&cities)

	sort.Sort(ByCount(cities))

	var hotCities []City

	count := 0
	for _, city := range cities {
		if city.Name != "" {
			hotCities = append(hotCities, city)

			count += 1
		}

		if count == 10 {
			break
		}
	}

	return a.Render("index.html", renders.T{
		"title":         "首页",
		"db":            DB,
		"nodes":         nodes,
		"cities":        hotCities,
		"status":        status,
		"topics":        topics,
		"linkExchanges": linkExchanges,
		"pagination":    pagination,
		"page":          page,
		"active":        "topic",
		"subActive":     subActive,
	})
}

type LatestReplyTopics struct {
	Topic
}

func (a *LatestReplyTopics) Get() error {
	return a.list(bson.M{"content.type": models.TypeTopic}, "-latestrepliedat", "/", "latestReply")
}

// 最新发布的主题，按照发布时间倒序排列
// URL: /topics/latest
type LatestTopics struct {
	Topic
}

func (a *LatestTopics) Get() error {
	return a.list(bson.M{"content.type": models.TypeTopic}, "-content.createdat", "/topics/latest", "latestCreate")
}

// URL: /topics/no_reply
// 无人回复的主题，按照发布时间倒序排列
type NoReplyTopics struct {
	Topic
}

func (a *NoReplyTopics) Get() error {
	return a.list(bson.M{"content.type": models.TypeTopic, "content.commentcount": 0}, "-content.createdat", "/topics/no_reply", "noReply")
}

type ShowTopic struct {
	RenderBase
}

// Get /t/:topicId 显示主题
func (a *ShowTopic) Get() error {
	topicID := a.Param("topicID")

	if !bson.IsObjectIdHex(topicID) {
		a.NotFound("参数错误")
		return nil
	}

	session, DB := models.GetSessionAndDB()
	defer session.Close()

	c := DB.C(models.CONTENTS)

	topic := models.Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicID), "content.type": models.TypeTopic}).One(&topic)

	if err != nil {
		a.NotFound(err.Error())
		return nil
	}

	// 点击数 +1
	c.UpdateId(bson.ObjectIdHex(topicID), bson.M{"$inc": bson.M{"content.hits": 1}})
	return a.Render("topic/show.html", renders.T{
		"title":    topic.Title,
		"topic":    topic,
		"comments": topic.Comments(DB),
		"db":       DB,
	})
}

/*
// URL: /topic/new
// 新建主题
func newTopicHandler(handler *Handler) {
	user, _ := currentUser(handler)
	if user.IsBlocked {
		handler.Redirect("/user/blocked")
		return
	}

	nodeId := mux.Vars(handler.Request)["node"]

	var nodes []Node
	c := handler.DB.C(NODES)
	c.Find(nil).All(&nodes)

	var choices = []wtforms.Choice{wtforms.Choice{}} // 第一个选项为空

	for _, node := range nodes {
		choices = append(choices, wtforms.Choice{Value: node.Id_.Hex(), Label: node.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewSelectField("node", "节点", choices, nodeId, &wtforms.Required{}),
		wtforms.NewTextArea("title", "标题", "", &wtforms.Required{}),
		wtforms.NewTextArea("editormd-markdown-doc", "内容", ""),
		wtforms.NewTextArea("editormd-html-code", "HTML", ""),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			c = handler.DB.C(CONTENTS)

			id_ := bson.NewObjectId()

			now := time.Now()

			nodeId := bson.ObjectIdHex(form.Value("node"))
			err := c.Insert(&Topic{
				Content: Content{
					Id_:       id_,
					Type:      TypeTopic,
					Title:     form.Value("title"),
					Markdown:  form.Value("editormd-markdown-doc"),
					Html:      template.HTML(form.Value("editormd-html-code")),
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
	}

	handler.renderTemplate("topic/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "新建",
		"action": "/topic/new",
		"active": "topic",
	})
}

// URL: /t/{topicId}/edit
// 编辑主题
func editTopicHandler(handler *Handler) {
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
		wtforms.NewSelectField("node", "节点", choices, topic.NodeId.Hex(), &wtforms.Required{}),
		wtforms.NewTextArea("title", "标题", topic.Title, &wtforms.Required{}),
		wtforms.NewTextArea("editormd-markdown-doc", "内容", topic.Markdown),
		wtforms.NewTextArea("editormd-html-code", "html", ""),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			nodeId := bson.ObjectIdHex(form.Value("node"))
			c = handler.DB.C(CONTENTS)
			c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{
				"nodeid":            nodeId,
				"content.title":     form.Value("title"),
				"content.markdown":  form.Value("editormd-markdown-doc"),
				"content.html":      template.HTML(form.Value("editormd-html-code")),
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
	}

	handler.renderTemplate("topic/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "编辑",
		"action": "/t/" + topicId + "/edit",
		"active": "topic",
	})
}

// URL: /go/{node}
// 列出节点下所有的主题
func topicInNodeHandler(handler *Handler) {
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

	handler.renderTemplate("/topic/list.html", BASE, map[string]interface{}{
		"topics": topics,
		"node":   node,
		"active": "topic",
	})
}

// URL: /t/{topicId}/collect/
// 将主题收藏至当前用户的收藏夹
func collectTopicHandler(handler *Handler) {
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
func deleteTopicHandler(handler *Handler) {
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

// 列出置顶的主题
func listTopTopicsHandler(handler *Handler) {
	var topics []Topic
	c := handler.DB.C(CONTENTS)
	c.Find(bson.M{"content.type": TypeTopic, "is_top": true}).All(&topics)

	handler.renderTemplate("/topic/top_list.html", ADMIN, map[string]interface{}{
		"topics": topics,
	})
}

// 设置置顶
func setTopTopicHandler(handler *Handler) {
	topicId := bson.ObjectIdHex(mux.Vars(handler.Request)["id"])
	c := handler.DB.C(CONTENTS)
	c.Update(bson.M{"_id": topicId}, bson.M{"$set": bson.M{"is_top": true}})
	handler.Redirect("/t/" + topicId.Hex())
}

// 取消置顶
func cancelTopTopicHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	topicId := bson.ObjectIdHex(vars["id"])

	c := handler.DB.C(CONTENTS)
	c.Update(bson.M{"_id": topicId}, bson.M{"$set": bson.M{"is_top": false}})
	handler.Redirect("/admin/top/topics")
}
*/

package actions

import (
	"strings"

	"gitea.com/tango/renders"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

type City struct {
	Name        string `bson:"_id"`
	MemberCount int    `bson:"count"`
}

type ByCount []City

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].MemberCount > a[j].MemberCount }

// ShowTopic 显示主题
type ShowTopic struct {
	RenderBase
}

// Get /t/:topicID 显示主题
func (a *ShowTopic) Get() error {
	topicID := a.Param("topicID")

	if !bson.IsObjectIdHex(topicID) {
		a.NotFound("参数错误")
		return nil
	}

	c := a.DB.C(models.CONTENTS)

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
		"comments": topic.Comments(a.DB),
	})
}

// NewTopic 新建主题
type NewTopic struct {
	AuthRenderBase
}

// Get /topic/new 新建主题页面
func (a *NewTopic) Get() error {
	return a.Render("topic/form.html", renders.T{
		"title":  "新建主题",
		"action": "new",
	})
}

// EditTopic 编辑主题
type EditTopic struct {
	AuthRenderBase
}

// Get /t/:tipicId 编辑主题页面
func (a *EditTopic) Get() error {
	topicID := a.Param("topicID")
	if !bson.IsObjectIdHex(topicID) {
		a.NotFound("参数错误")
		return nil
	}

	return a.Render("topic/form.html", renders.T{
		"title":   "编辑主题",
		"action":  "edit",
		"topicID": topicID,
	})
}

// NodeTopics 节点下的所有主题
type NodeTopics struct {
	RenderBase
}

// Get /go/:node
func (a *NodeTopics) Get() error {
	nodeId := a.Param("node")

	c := a.DB.C(models.NODES)

	node := models.Node{}
	err := c.Find(bson.M{"id": nodeId}).One(&node)

	if err != nil {
		a.NotFound("没有此节点")
		return nil
	}

	topics, pagination, err := GetTopics(a, a.DB, bson.M{"nodeid": node.Id_, "content.type": models.TypeTopic})

	return a.Render("index.html", renders.T{
		"title":      node.Name + "主题列表",
		"topics":     topics,
		"node":       node,
		"pagination": pagination,
		"url":        "/go/" + nodeId,
	})
}

// SearchTopic 检索主题
type SearchTopic struct {
	RenderBase
}

// Get: /search
func (a *SearchTopic) Get() error {
	page := a.FormInt("p", 1)
	if page <= 0 {
		page = 1
	}

	q := a.Form("q")
	keywords := strings.Split(q, " ")

	var noSpaceKeywords []string

	for _, keyword := range keywords {
		temp := strings.TrimSpace(keyword)
		if temp != "" {
			noSpaceKeywords = append(noSpaceKeywords, temp)
		}
	}

	var titleConditions []bson.M
	var markdownConditions []bson.M

	for _, keyword := range noSpaceKeywords {
		titleConditions = append(titleConditions, bson.M{"content.title": bson.M{"$regex": bson.RegEx{keyword, "i"}}})
		markdownConditions = append(markdownConditions, bson.M{"content.markdown": bson.M{"$regex": bson.RegEx{keyword, "i"}}})
	}

	var pagination *Pagination

	var conditions bson.M

	if len(noSpaceKeywords) == 0 {
		conditions = bson.M{"content.type": models.TypeTopic}
	} else {
		conditions = bson.M{
			"$and": []bson.M{
				bson.M{"content.type": models.TypeTopic},
				bson.M{"$or": []bson.M{
					bson.M{"$and": titleConditions},
					bson.M{"$and": markdownConditions},
				},
				},
			},
		}
	}

	topics, pagination, err := GetTopics(a, a.DB, conditions)

	if err != nil {
		a.NotFound(err.Error())
		return nil
	}

	return a.Render("topic/result.html", renders.T{
		"url":        "/search?q=" + q,
		"q":          q,
		"topics":     topics,
		"pagination": pagination,
		"page":       page,
	})
}

type FormInt interface {
	FormInt(key string, defaults ...int) int
}

// GetTopics 查询主题
func GetTopics(ctx FormInt, db *mgo.Database, conditions bson.M) ([]models.Topic, *Pagination, error) {
	page := ctx.FormInt("p", 1)
	if page <= 0 {
		page = 1
	}

	c := db.C(models.CONTENTS)
	var pagination = NewPagination(c.Find(conditions).Sort("-latestrepliedat"), PerPage)

	var topics []models.Topic

	query, err := pagination.Page(page)
	if err != nil {
		return nil, nil, err
	}

	query.All(&topics)

	return topics, pagination, nil
}

// CollectTopic 收藏主题
type CollectTopic struct {
	AuthRenderBase
}

/*

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

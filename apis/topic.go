package apis

import (
	"fmt"
	"html/template"
	"time"

	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

// 获取主题信息
type GetTopic struct {
	Base
	binding.Binder
}

// Get /api/topic/:topicId
func (a *GetTopic) Get() interface{} {
	topicId := a.Param("topicId")
	if !bson.IsObjectIdHex(topicId) {
		return map[string]interface{}{
			"status":  0,
			"message": "错误的主题 id",
		}
	}

	c := a.DB.C(models.CONTENTS)

	topic := models.Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": models.TypeTopic}).One(&topic)

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "没找到该主题",
		}
	}

	if !topic.CanEdit(a.User.Username, a.DB) {
		return map[string]interface{}{
			"status":  0,
			"message": "对不起，你没有权限编辑该主题",
		}
	}

	return map[string]interface{}{
		"title":    topic.Title,
		"node_id":  topic.NodeId,
		"markdown": topic.Markdown,
	}
}

type TopicForm struct {
	Title    string `json:"title"`
	NodeId   string `json:"node_id"`
	Markdown string `json:"markdown"`
	Html     string `json:"html"`
}

// NewTopic 新建主题
type NewTopic struct {
	Base
	binding.Binder
}

// Post /topic/new 新建主题
func (a *NewTopic) Post() interface{} {
	if a.User.IsBlocked {
		return map[string]interface{}{
			"status":  0,
			"message": "当前账户被禁言，不能新建主题",
		}
	}

	var form TopicForm
	a.ReadJSON(&form)

	// TODO form 校验

	var c = a.DB.C(models.CONTENTS)

	id_ := bson.NewObjectId()

	now := time.Now()

	nodeId := bson.ObjectIdHex(form.NodeId)
	err := c.Insert(&models.Topic{
		Content: models.Content{
			Id_:       id_,
			Type:      models.TypeTopic,
			Title:     form.Title,
			Markdown:  form.Markdown,
			Html:      template.HTML(form.Html),
			CreatedBy: a.User.Id_,
			CreatedAt: now,
		},
		Id_:             id_,
		NodeId:          nodeId,
		LatestRepliedAt: now,
	})

	if err != nil {
		fmt.Println("newTopicHandler:", err.Error())
		return map[string]interface{}{
			"status":  0,
			"message": "主题新建错误" + err.Error(),
		}
	}

	// 增加Node.TopicCount
	c = a.DB.C(models.NODES)
	c.Update(bson.M{"_id": nodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})

	// 增加总主题数
	c = a.DB.C(models.STATUS)
	c.Update(nil, bson.M{"$inc": bson.M{"topiccount": 1}})
	return map[string]interface{}{
		"status":   1,
		"message":  "主题新建成功",
		"topic_id": id_.Hex(),
	}
}

// EditTopic 编辑主题
type EditTopic struct {
	Base
}

// Post /topic/:topicId/edit
func (a *EditTopic) Post() interface{} {
	topicId := a.Param("topicId")

	if !bson.IsObjectIdHex(topicId) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	c := a.DB.C(models.CONTENTS)

	topic := models.Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicId), "content.type": models.TypeTopic}).One(&topic)

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "没有该主题",
		}
	}

	if !topic.CanEdit(a.User.Username, a.DB) {
		return map[string]interface{}{
			"status":  0,
			"message": "对不起，你没有权限编辑该主题",
		}
	}

	var form TopicForm
	a.ReadJSON(&form)

	// TODO form 校验

	var newNodeId = bson.ObjectIdHex(form.NodeId)

	c = a.DB.C(models.CONTENTS)
	c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{
		"nodeid":            newNodeId,
		"content.title":     form.Title,
		"content.markdown":  form.Markdown,
		"content.html":      template.HTML(form.Html),
		"content.updatedat": time.Now(),
		"content.updatedby": a.User.Id_.Hex(),
	}})

	// 如果两次的节点不同,更新节点的主题数量
	if topic.NodeId != newNodeId {
		c = a.DB.C(models.NODES)
		c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})
		c.Update(bson.M{"_id": newNodeId}, bson.M{"$inc": bson.M{"topiccount": 1}})
	}

	return map[string]interface{}{
		"status":   1,
		"message":  "主题编辑成功",
		"topic_id": topicId,
	}
}

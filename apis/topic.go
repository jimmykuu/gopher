package apis

import (
	"fmt"
	"html/template"
	"time"

	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

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

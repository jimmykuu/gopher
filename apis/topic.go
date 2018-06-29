package apis

import (
	"fmt"
	"html/template"
	"time"

	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

// Topic 主题
type Topic struct {
	Base
	binding.Binder
}

// Get /api/topic/:topicID
func (a *Topic) Get() interface{} {
	topicID := a.Param("topicID")
	if !bson.IsObjectIdHex(topicID) {
		return map[string]interface{}{
			"status":  0,
			"message": "错误的主题 id",
		}
	}

	c := a.DB.C(models.CONTENTS)

	topic := models.Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicID), "content.type": models.TypeTopic}).One(&topic)

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

// Delete /topic/:topicID
func (a *Topic) Delete() interface{} {
	topicID := a.Param("topicID")
	if !bson.IsObjectIdHex(topicID) {
		return map[string]interface{}{
			"status":  0,
			"message": "错误的主题 id",
		}
	}

	c := a.DB.C(models.CONTENTS)

	topic := models.Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicID), "content.type": models.TypeTopic}).One(&topic)

	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "没有找到主题",
		}
	}

	if !topic.CanDelete(a.User.Username, a.DB) {
		return map[string]interface{}{
			"status":  0,
			"message": "没有删除该主题的权限",
		}
	}

	// Node统计数减一
	c = a.DB.C(models.NODES)
	c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})

	c = a.DB.C(models.STATUS)
	// 统计的主题数减一，减去统计的回复数减去该主题的回复数
	c.Update(nil, bson.M{"$inc": bson.M{"topiccount": -1, "replycount": -topic.CommentCount}})

	//删除评论
	c = a.DB.C(models.COMMENTS)
	if topic.CommentCount > 0 {
		c.Remove(bson.M{"contentid": topic.Id_})
	}

	// 删除Topic记录
	c = a.DB.C(models.CONTENTS)
	c.Remove(bson.M{"_id": topic.Id_})

	return map[string]interface{}{
		"status": 1,
	}
}

// TopicForm 主题表单，新建和编辑共用
type TopicForm struct {
	Title    string `json:"title"`
	NodeID   string `json:"node_id"`
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
}

// Post /topic/new 新建主题
func (a *Topic) Post() interface{} {
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

	id := bson.NewObjectId()

	now := time.Now()

	nodeID := bson.ObjectIdHex(form.NodeID)
	err := c.Insert(&models.Topic{
		Content: models.Content{
			Id_:       id,
			Type:      models.TypeTopic,
			Title:     form.Title,
			Markdown:  form.Markdown,
			Html:      template.HTML(form.HTML),
			CreatedBy: a.User.Id_,
			CreatedAt: now,
		},
		Id_:             id,
		NodeId:          nodeID,
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
	c.Update(bson.M{"_id": nodeID}, bson.M{"$inc": bson.M{"topiccount": 1}})

	// 增加总主题数
	c = a.DB.C(models.STATUS)
	c.Update(nil, bson.M{"$inc": bson.M{"topiccount": 1}})
	return map[string]interface{}{
		"status":   1,
		"message":  "主题新建成功",
		"topic_id": id.Hex(),
	}
}

// Put /topic/:topicId/edit
func (a *Topic) Put() interface{} {
	topicID := a.Param("topicID")

	if !bson.IsObjectIdHex(topicID) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	c := a.DB.C(models.CONTENTS)

	topic := models.Topic{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(topicID), "content.type": models.TypeTopic}).One(&topic)

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

	var newNodeID = bson.ObjectIdHex(form.NodeID)

	c = a.DB.C(models.CONTENTS)
	c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{
		"nodeid":            newNodeID,
		"content.title":     form.Title,
		"content.markdown":  form.Markdown,
		"content.html":      template.HTML(form.HTML),
		"content.updatedat": time.Now(),
		"content.updatedby": a.User.Id_.Hex(),
	}})

	// 如果两次的节点不同,更新节点的主题数量
	if topic.NodeId != newNodeID {
		c = a.DB.C(models.NODES)
		c.Update(bson.M{"_id": topic.NodeId}, bson.M{"$inc": bson.M{"topiccount": -1}})
		c.Update(bson.M{"_id": newNodeID}, bson.M{"$inc": bson.M{"topiccount": 1}})
	}

	return map[string]interface{}{
		"status":   1,
		"message":  "主题编辑成功",
		"topic_id": topicID,
	}
}

// CollectTopic 收藏主题
type CollectTopic struct {
	Base
	binding.Binder
}

// Get /t/:topicID/collect 收藏主题
func (ct *CollectTopic) Get() interface{} {
	topicID := ct.Param("topicID")

	if !bson.IsObjectIdHex(topicID) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	user := ct.User
	var collected bool
	for _, v := range user.TopicsCollected {
		if v.TopicId == topicID {
			collected = true
			break
		}
	}

	if !collected {
		t := time.Now()
		collectTopic := models.CollectTopic{
			TopicId:       topicID,
			TimeCollected: t,
		}
		user.TopicsCollected = append(user.TopicsCollected, collectTopic)

		c := ct.DB.C(models.USERS)
		err := c.UpdateId(user.Id_, bson.M{"$set": bson.M{"topicscollected": user.TopicsCollected}})

		if err != nil {
			return map[string]interface{}{
				"status":  0,
				"message": "用户不存在",
			}
		}
	}

	return map[string]interface{}{
		"status": 1,
	}
}

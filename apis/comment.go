package apis

import (
	"html/template"
	"time"

	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

// Comment 评论
type Comment struct {
	Base
	binding.Binder
}

// Get /comment/:commentId 获取一条评论信息
func (a *Comment) Get() interface{} {
	commentIdStr := a.Param("commentId")
	if !bson.IsObjectIdHex(commentIdStr) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	commentId := bson.ObjectIdHex(commentIdStr)

	c := a.DB.C(models.COMMENTS)

	var comment models.Comment
	err := c.Find(bson.M{"_id": commentId}).One(&comment)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "该评论不存在",
		}
	}

	if !comment.CanDeleteOrEdit(a.User.Username, a.DB) {
		return map[string]interface{}{
			"status":  0,
			"message": "没有权限编辑该评论",
		}
	}

	return map[string]interface{}{
		"status":   1,
		"markdown": comment.Markdown,
		"html":     comment.Html,
	}
}

// Post /comment/:contentId 发表评论
func (a *Comment) Post() interface{} {
	if a.User.IsBlocked {
		return map[string]interface{}{
			"status":  0,
			"message": "当前账户被禁言，不能发表回复",
		}
	}

	contentIdStr := a.Param("contentId")
	if !bson.IsObjectIdHex(contentIdStr) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	contentId := bson.ObjectIdHex(contentIdStr)

	var temp map[string]interface{}
	c := a.DB.C(models.CONTENTS)
	c.Find(bson.M{"_id": contentId}).One(&temp)

	temp2 := temp["content"].(map[string]interface{})
	type_ := temp2["type"].(int)

	var form struct {
		Markdown string `json:"markdown"`
		Html     string `json:"html"`
	}

	a.ReadJSON(&form)

	c.Update(bson.M{"_id": contentId}, bson.M{"$inc": bson.M{"content.commentcount": 1}})

	commentId := bson.NewObjectId()
	now := time.Now()

	c = a.DB.C(models.COMMENTS)
	c.Insert(&models.Comment{
		Id_:       commentId,
		Type:      type_,
		ContentId: contentId,
		Markdown:  form.Markdown,
		Html:      template.HTML(form.Html),
		CreatedBy: a.User.Id_,
		CreatedAt: now,
	})
	if type_ == models.TypeTopic {
		// 修改最后回复用户Id和时间
		c = a.DB.C(models.CONTENTS)
		c.Update(bson.M{"_id": contentId}, bson.M{"$set": bson.M{"latestreplierid": a.User.Id_.Hex(), "latestrepliedat": now}})

		// 修改总的回复数量
		c = a.DB.C(models.STATUS)
		c.Update(nil, bson.M{"$inc": bson.M{"replycount": 1}})
	}

	return map[string]interface{}{
		"status":  1,
		"message": "发表评论成功",
	}
}

// Put /api/comment/:commentId 编辑评论
func (a *Comment) Put() interface{} {
	commentIdStr := a.Param("commentId")
	if !bson.IsObjectIdHex(commentIdStr) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	commentId := bson.ObjectIdHex(commentIdStr)

	c := a.DB.C(models.COMMENTS)

	comment := models.Comment{}

	err := c.Find(bson.M{"_id": commentId}).One(&comment)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "没有找到该评论",
		}
	}

	if !comment.CanDeleteOrEdit(a.User.Username, a.DB) {
		return map[string]interface{}{
			"status":  0,
			"message": "没有权限编辑该评论",
		}
	}

	var form struct {
		Markdown string `json:"markdown"`
		Html     string `json:"html"`
	}

	a.ReadJSON(&form)

	c.Update(bson.M{"_id": commentId}, bson.M{"$set": bson.M{
		"markdown":  form.Markdown,
		"html":      template.HTML(form.Html),
		"updatedby": a.User.Id_.Hex(),
		"updatedat": time.Now(),
	}})

	return map[string]interface{}{
		"status": 1,
		"html":   form.Html,
	}
}

// Delete /api/comment/:commentId 删除一条评论
func (a *Comment) Delete() interface{} {
	commentIdStr := a.Param("commentId")
	if !bson.IsObjectIdHex(commentIdStr) {
		return map[string]interface{}{
			"status":  0,
			"message": "参数错误",
		}
	}

	commentId := bson.ObjectIdHex(commentIdStr)

	c := a.DB.C(models.COMMENTS)

	var comment models.Comment
	err := c.Find(bson.M{"_id": commentId}).One(&comment)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": "该评论不存在",
		}
	}

	if !comment.CanDeleteOrEdit(a.User.Username, a.DB) {
		return map[string]interface{}{
			"status":  0,
			"message": "没有权限删除该评论",
		}
	}

	c.Remove(bson.M{"_id": commentId})

	c = a.DB.C(models.CONTENTS)
	c.Update(bson.M{"_id": comment.ContentId}, bson.M{"$inc": bson.M{"content.commentcount": -1}})

	if comment.Type == models.TypeTopic {
		var topic models.Topic
		c.Find(bson.M{"_id": comment.ContentId}).One(&topic)
		if topic.LatestReplierId == comment.CreatedBy.Hex() {
			if topic.CommentCount == 0 {
				// 如果删除后没有回复，设置最后回复id为空，最后回复时间为创建时间
				c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{"latestreplierid": "", "latestrepliedat": topic.CreatedAt}})
			} else {
				// 如果删除的是该主题最后一个回复，设置主题的最新回复id，和时间
				var latestComment models.Comment
				c = a.DB.C(models.COMMENTS)
				c.Find(bson.M{"contentid": topic.Id_}).Sort("-createdat").Limit(1).One(&latestComment)

				c = a.DB.C(models.CONTENTS)
				c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{"latestreplierid": latestComment.CreatedBy.Hex(), "latestrepliedat": latestComment.CreatedAt}})
			}
		}

		// 修改总的回复数量
		c = a.DB.C(models.STATUS)
		c.Update(nil, bson.M{"$inc": bson.M{"replycount": -1}})
	}

	return map[string]interface{}{
		"status": 1,
	}
}

package apis

import (
	"html/template"
	"time"

	"github.com/tango-contrib/binding"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

// Comment 发表评论
type Comment struct {
	Base
	binding.Binder
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

/*
// URL: /comment/{commentId}/delete
// 删除评论
func deleteCommentHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	var commentId string = vars["commentId"]

	c := handler.DB.C(COMMENTS)
	var comment Comment
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(commentId)}).One(&comment)

	if err != nil {
		message(handler, "评论不存在", "该评论不存在", "error")
		return
	}

	c.Remove(bson.M{"_id": comment.Id_})

	c = handler.DB.C(CONTENTS)
	c.Update(bson.M{"_id": comment.ContentId}, bson.M{"$inc": bson.M{"content.commentcount": -1}})

	if comment.Type == TypeTopic {
		var topic Topic
		c.Find(bson.M{"_id": comment.ContentId}).One(&topic)
		if topic.LatestReplierId == comment.CreatedBy.Hex() {
			if topic.CommentCount == 0 {
				// 如果删除后没有回复，设置最后回复id为空，最后回复时间为创建时间
				c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{"latestreplierid": "", "latestrepliedat": topic.CreatedAt}})
			} else {
				// 如果删除的是该主题最后一个回复，设置主题的最新回复id，和时间
				var latestComment Comment
				c = handler.DB.C("comments")
				c.Find(bson.M{"contentid": topic.Id_}).Sort("-createdat").Limit(1).One(&latestComment)

				c = handler.DB.C("contents")
				c.Update(bson.M{"_id": topic.Id_}, bson.M{"$set": bson.M{"latestreplierid": latestComment.CreatedBy.Hex(), "latestrepliedat": latestComment.CreatedAt}})
			}
		}

		// 修改中的回复数量
		c = handler.DB.C(STATUS)
		c.Update(nil, bson.M{"$inc": bson.M{"replycount": -1}})
	}

	var url string
	switch comment.Type {
	case TypeArticle:
		url = "/a/" + comment.ContentId.Hex()
	case TypeTopic:
		url = "/t/" + comment.ContentId.Hex()
	case TypePackage:
		url = "/p/" + comment.ContentId.Hex()
	}

	http.Redirect(handler.ResponseWriter, handler.Request, url, http.StatusFound)
}

// URL: /comment/:id.json
// 获取comment的内容
func commentJsonHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	var id string = vars["id"]

	c := handler.DB.C(COMMENTS)
	var comment Comment
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&comment)

	if err != nil {
		return
	}

	data := map[string]string{
		"markdown": comment.Markdown,
	}

	handler.renderJson(data)
}

// URL: /commeint/:id/edit
// 编辑comment
func editCommentHandler(handler *Handler) {
	if handler.Request.Method != "POST" {
		return
	}
	vars := mux.Vars(handler.Request)
	var id string = vars["id"]

	c := handler.DB.C(COMMENTS)

	user, _ := currentUser(handler)

	comment := Comment{}

	c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&comment)

	if !comment.CanDeleteOrEdit(user.Username, handler.DB) {
		return
	}

	markdown := handler.Request.FormValue("editormd-edit-markdown-doc")
	html := handler.Request.FormValue("editormd-edit-html-code")

	c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{
		"markdown":  markdown,
		"html":      template.HTML(html),
		"updatedby": user.Id_.Hex(),
		"updatedat": time.Now(),
	}})

	var temp map[string]interface{}
	c = handler.DB.C(CONTENTS)
	c.Find(bson.M{"_id": comment.ContentId}).One(&temp)

	temp2 := temp["content"].(map[string]interface{})
	type_ := temp2["type"].(int)

	var url string
	switch type_ {
	case TypeArticle:
		url = "/a/" + comment.ContentId.Hex()
	case TypeTopic:
		url = "/t/" + comment.ContentId.Hex()
	case TypePackage:
		url = "/p/" + comment.ContentId.Hex()
	}

	http.Redirect(handler.ResponseWriter, handler.Request, url, http.StatusFound)
}
*/

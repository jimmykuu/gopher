package gopher

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

// URL: /comment/{contentId}
// 评论，不同内容共用一个评论方法
func commentHandler(handler Handler) {
	if handler.Request.Method != "POST" {
		return
	}

	user, _ := currentUser(handler)

	vars := mux.Vars(handler.Request)
	contentIdStr := vars["contentId"]
	contentId := bson.ObjectIdHex(contentIdStr)

	var temp map[string]interface{}
	c := handler.DB.C(CONTENTS)
	c.Find(bson.M{"_id": contentId}).One(&temp)

	temp2 := temp["content"].(map[string]interface{})
	var contentCreator bson.ObjectId
	contentCreator = temp2["createdby"].(bson.ObjectId)
	type_ := temp2["type"].(int)

	var url string
	switch type_ {
	case TypeArticle:
		url = "/a/" + contentIdStr
	case TypeTopic:
		url = "/t/" + contentIdStr
	case TypePackage:
		url = "/p/" + contentIdStr
	}

	c.Update(bson.M{"_id": contentId}, bson.M{"$inc": bson.M{"content.commentcount": 1}})

	content := handler.Request.FormValue("content")

	html := handler.Request.FormValue("html")
	html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

	Id_ := bson.NewObjectId()
	now := time.Now()

	c = handler.DB.C(COMMENTS)
	c.Insert(&Comment{
		Id_:       Id_,
		Type:      type_,
		ContentId: contentId,
		Markdown:  content,
		Html:      template.HTML(html),
		CreatedBy: user.Id_,
		CreatedAt: now,
	})

	if type_ == TypeTopic {
		// 修改最后回复用户Id和时间
		c = handler.DB.C(CONTENTS)
		c.Update(bson.M{"_id": contentId}, bson.M{"$set": bson.M{"latestreplierid": user.Id_.Hex(), "latestrepliedat": now}})

		// 修改中的回复数量
		c = handler.DB.C(STATUS)
		c.Update(nil, bson.M{"$inc": bson.M{"replycount": 1}})
		/*mark ggaaooppeenngg*/
		//修改用户的最近回复

		c = handler.DB.C(USERS)
		//查找评论中at的用户,并且更新recentAts
		users := findAts(content)
		for _, v := range users {
			var user User
			err := c.Find(bson.M{"username": v}).One(&user)
			if err != nil {
				fmt.Println(err)
			} else {
				user.RecentAts = append(user.RecentAts, At{user.Username, contentIdStr, Id_.Hex()})
				if err = c.Update(bson.M{"username": user.Username}, bson.M{"$set": bson.M{"recentats": user.RecentAts}}); err != nil {
					fmt.Println(err)
				}
			}
		}

		//修改用户的最近回复
		//该最近回复提醒通过url被点击的时候会被disactive
		//更新最近的评论
		//自己的评论就不提示了
		tempTitle := temp2["title"].(string)

		if contentCreator.Hex() != user.Id_.Hex() {
			var recentreplies []Reply
			var Creater User
			err := c.Find(bson.M{"_id": contentCreator}).One(&Creater)
			if err != nil {
				fmt.Println(err)
			}
			recentreplies = Creater.RecentReplies
			//添加最近评论所在的主题id
			duplicate := false
			for _, v := range recentreplies {
				if contentIdStr == v.ContentId {
					duplicate = true
				}
			}
			//如果回复的主题有最近回复的话就不添加进去，因为在同一主题下就能看到
			if !duplicate {
				recentreplies = append(recentreplies, Reply{contentIdStr, tempTitle})

				if err = c.Update(bson.M{"_id": contentCreator}, bson.M{"$set": bson.M{"recentreplies": recentreplies}}); err != nil {
					fmt.Println(err)
				}
			}
		}

	}

	http.Redirect(handler.ResponseWriter, handler.Request, url, http.StatusFound)
}

// URL: /comment/{commentId}/delete
// 删除评论
func deleteCommentHandler(handler Handler) {
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
func commentJsonHandler(handler Handler) {
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
		"html":     string(comment.Html),
	}

	renderJson(handler, data)
}

// URL: /commeint/:id/edit
// 编辑comment
func editCommentHandler(handler Handler) {
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

	content := handler.Request.FormValue("content")

	html := handler.Request.FormValue("html")
	html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)

	c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{
		"markdown":  content,
		"html":      html,
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

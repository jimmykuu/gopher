// 提供 RESTful API 服务

package gopher

import (
	"gopkg.in/mgo.v2/bson"
)

func apiTopicsHandler(handler *Handler) {
	c := handler.DB.C(CONTENTS)

	var topics []Topic

	c.Find(bson.M{"content.type": TypeTopic}).Sort("-latestrepliedat").Limit(20).All(&topics)

	result := []map[string]interface{}{}
	for _, topic := range topics {
		creater := topic.Content.Creater(handler.DB)

		result = append(result, map[string]interface{}{
			"id":            topic.Id_.Hex(),
			"title":         topic.Content.Title,
			"markdown":      topic.Content.Markdown,
			"html":          topic.Content.Html,
			"comment_count": topic.Content.CommentCount,
			"created_by": map[string]interface{}{
				"id":       creater.Id_.Hex(),
				"username": creater.Username,
				"avatar":   creater.AvatarImgSrc(48),
			},
			"created_at":        topic.Content.CreatedAt.Format("2006-01-02 15:04:05"),
			"latest_replied_at": topic.LatestRepliedAt.Format("2006-01-02 15:04:05"),
		})
	}

	handler.renderJson(result)
}

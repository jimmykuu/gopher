package gopher

import (
	"container/list"
	"text/template"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	contents   []Topic   //当日内容
	cache      list.List //缓存
	latestTime time.Time //最后更新时间
)

// 返回今天凌晨00:00时间
func Dawn() time.Time {
	now := time.Now()
	t := now.Round(24 * time.Hour)
	if t.After(now) {
		t = t.AddDate(0, 0, -1)
	}
	t = t.Add(-time.Hour * time.Duration(t.Hour()))
	return t
}

func init() {
	latestTime = Dawn()
	latestTime = latestTime.AddDate(0, 0, -7) // 初次启动获取最近一周的内容.
}

var flag bool

func RssRefresh() {
	for {
		now := time.Now()
		if now.After(latestTime) {
			session, err := mgo.Dial(Config.DB)
			if err != nil {
				logger.Println(err)
				latestTime = now.Add(time.Hour) //一小时后重新连接.
				continue
			} else {
				c := session.DB("gopher").C("contents")
				c.Find(bson.M{"content.createdat": bson.M{"$gt": latestTime}}).Sort("-content.createdat").All(&contents)
				latestTime = now
				cache.PushBack(contents)
				if cache.Len() > 7 {
					cache.Remove(cache.Front())
				}
			}
		}
		time.Sleep(24 * time.Hour)
	}
}

func getFromCache() []Topic {
	var topics []Topic
	for e := cache.Back(); e != nil; e = e.Prev() {
		ts := e.Value.([]Topic)
		topics = append(topics, ts...)
	}
	return topics
}

func rssHandler(handler *Handler) {

	t, err := template.ParseFiles("templates/rss.xml")
	if err != nil {
		logger.Println(err)
	}
	rssTopics := getFromCache()
	handler.ResponseWriter.Header().Set("Content-Type", "application/xml")
	t.Execute(handler.ResponseWriter, map[string]interface{}{
		"date":   latestTime,
		"topics": rssTopics,
		"utils":  utils,
	})
}

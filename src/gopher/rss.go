package gopher

import (
	"container/list"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"labix.org/v2/mgo/bson"
)

var (
	//当日内容
	contents []Topic
	//缓存
	cache list.List
	//最后更新时间
	latestTime *time.Time
)

//今天凌晨零点时间
func Dawn() *time.Time {
	now := time.Now()
	t := now.Round(24 * time.Hour)
	if t.After(now) {
		t = t.AddDate(0, 0, -1)
	}
	return &t
}

func init() {
	latestTime = Dawn()
}

func RssRefresh() {
	now := time.Now()
	if now.After(*latestTime) {
		latestTime = &now
		c := DB.C("contents")
		c.Find(bson.M{"content.createdat": bson.M{"$gt": latestTime}}).All(&contents)
		fmt.Print(contents)
		cache.PushBack(contents)
		if cache.Len() > 7 {
			cache.Remove(cache.Front())
		}
	}

	time.Sleep(24 * time.Hour)
}

func getFromCache() []Topic {
	var topics []Topic
	for e := cache.Back(); e != nil; e = e.Prev() {
		ts := e.Value.([]Topic)
		topics = append(topics, ts...)
	}
	return topics
}

func rssHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/rss.xml")
	if err != nil {
		fmt.Println(err)
	}
	rssTopics := getFromCache()
	t.Execute(w, map[string]interface{}{
		"date":   *latestTime,
		"topics": rssTopics,
		"utils":  utils,
	})

}

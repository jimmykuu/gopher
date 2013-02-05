/*
把mongodb中的topics, articles, sites, packages里的内容都移到contents中
*/

package main

import (
	"encoding/json"
	"fmt"
	"gopher"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"time"
)

var (
	config map[string]string
	db     *mgo.Database
)

// 站点
type OldSite struct {
	Id_         bson.ObjectId `bson:"_id"`
	Name        string
	Url         string
	Description string
	CategoryId  bson.ObjectId
	UserId      bson.ObjectId
}

// 评论
type OldComment struct {
	Id_       bson.ObjectId `bson:"_id"`
	UserId    bson.ObjectId
	Markdown  string
	Html      template.HTML
	CreatedAt time.Time
}

// 文章
type OldArticle struct {
	Id_            bson.ObjectId `bson:"_id"`
	CategoryId     bson.ObjectId
	UserId         bson.ObjectId
	Title          string
	Markdown       string
	Html           template.HTML
	OriginalSource string
	OriginalUrl    string
	CreatedAt      time.Time
	Hits           int
	Comments       []OldComment
}

type OldPackage struct {
	Id_        bson.ObjectId `bson:"_id"`
	UserId     bson.ObjectId
	CategoryId bson.ObjectId
	Name       string
	Url        string
	Markdown   string
	Html       template.HTML
	CreatedAt  time.Time
}

// 主题
type OldTopic struct {
	Id_             bson.ObjectId `bson:"_id"`
	NodeId          bson.ObjectId
	UserId          bson.ObjectId
	Title           string
	Markdown        string
	Html            template.HTML
	CreatedAt       time.Time
	ReplyCount      int
	LatestReplyId   string
	LatestRepliedAt time.Time
	Hits            int
}

// 回复
type OldReply struct {
	Id_       bson.ObjectId `bson:"_id"`
	UserId    bson.ObjectId
	TopicId   bson.ObjectId
	Markdown  string
	Html      template.HTML
	CreatedAt time.Time
}

func init() {
	file, err := os.Open("etc/config.json")
	if err != nil {
		fmt.Println("配置文件读取失败:", err.Error())
		os.Exit(1)
	}

	defer file.Close()

	dec := json.NewDecoder(file)

	err = dec.Decode(&config)

	if err != nil {
		fmt.Println("配置文件解析失败:", err.Error())
		os.Exit(1)
	}

	if config["db"] == "" {
		fmt.Println("数据库地址还没有配置,请到config.json内配置db字段.")
		os.Exit(1)
	}

	session, err := mgo.Dial(config["db"])
	if err != nil {
		fmt.Println("MongoDB连接失败:", err.Error())
	}

	session.SetMode(mgo.Monotonic, true)

	db = session.DB("gopher")
}

func moveSites() {
	var sites []OldSite
	c := db.C("sites")
	c.Find(nil).All(&sites)

	c = db.C("contents")
	for _, site := range sites {
		c.Insert(&gopher.Site{
			Id_: site.Id_,
			Content: gopher.Content{
				Id_:       site.Id_,
				Type:      gopher.TypeSite,
				Title:     site.Name,
				Markdown:  site.Description,
				CreatedBy: site.UserId,
				CreatedAt: time.Now(),
			},
			Url:        site.Url,
			CategoryId: site.CategoryId,
		})
	}
}

func moveArticles() {
	var articles []OldArticle
	c := db.C("articles")
	c.Find(nil).All(&articles)

	c1 := db.C("contents")
	c2 := db.C("comments")
	for _, article := range articles {
		c1.Insert(&gopher.Article{
			Content: gopher.Content{
				Id_:          article.Id_,
				Type:         gopher.TypeArticle,
				Title:        article.Title,
				Markdown:     article.Markdown,
				Html:         article.Html,
				Hits:         article.Hits,
				CommentCount: len(article.Comments),
				CreatedBy:    article.UserId,
				CreatedAt:    article.CreatedAt,
			},
			Id_:            article.Id_,
			CategoryId:     article.CategoryId,
			OriginalSource: article.OriginalSource,
			OriginalUrl:    article.OriginalUrl,
		})

		for _, comment := range article.Comments {
			c2.Insert(&gopher.Comment{
				Id_:       comment.Id_,
				Type:      gopher.TypeArticle,
				ContentId: article.Id_,
				Markdown:  comment.Markdown,
				Html:      comment.Html,
				CreatedBy: comment.UserId,
				CreatedAt: comment.CreatedAt,
			})
		}
	}
}

func movePackages() {
	var packages []OldPackage
	c := db.C("packages")
	c.Find(nil).All(&packages)

	c = db.C("contents")
	for _, package_ := range packages {
		c.Insert(&gopher.Package{
			Content: gopher.Content{
				Id_:       package_.Id_,
				Type:      gopher.TypePackage,
				Title:     package_.Name,
				Markdown:  package_.Markdown,
				Html:      package_.Html,
				CreatedBy: package_.UserId,
				CreatedAt: package_.CreatedAt,
			},
			Id_:        package_.Id_,
			CategoryId: package_.CategoryId,
			Url:        package_.Url,
		})
	}
}

func moveTopics() {
	var topics []OldTopic
	c := db.C("topics")
	c.Find(nil).All(&topics)

	replyIdAndReplierId := make(map[string]string)

	var replies []OldReply
	c = db.C("replies")
	c.Find(nil).All(&replies)

	c = db.C("comments")
	for _, reply := range replies {
		replyIdAndReplierId[reply.Id_.Hex()] = reply.UserId.Hex()

		c.Insert(&gopher.Comment{
			Id_:       reply.Id_,
			Type:      gopher.TypeTopic,
			ContentId: reply.TopicId,
			Markdown:  reply.Markdown,
			Html:      reply.Html,
			CreatedBy: reply.UserId,
			CreatedAt: reply.CreatedAt,
		})
	}

	c = db.C("contents")
	for _, topic := range topics {
		latestReplierId := ""
		if topic.LatestReplyId != "" {
			latestReplierId = replyIdAndReplierId[topic.LatestReplyId]
		}
		c.Insert(&gopher.Topic{
			Content: gopher.Content{
				Id_:          topic.Id_,
				Type:         gopher.TypeTopic,
				Title:        topic.Title,
				Markdown:     topic.Markdown,
				Html:         topic.Html,
				Hits:         topic.Hits,
				CommentCount: topic.ReplyCount,
				CreatedBy:    topic.UserId,
				CreatedAt:    topic.CreatedAt,
			},
			Id_:             topic.Id_,
			NodeId:          topic.NodeId,
			LatestReplierId: latestReplierId,
			LatestRepliedAt: topic.LatestRepliedAt,
		})
	}
}

func main() {
	moveTopics()
	moveSites()
	moveArticles()
	movePackages()
}

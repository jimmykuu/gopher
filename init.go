package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/conf"
	"github.com/jimmykuu/gopher/models"
)

var (
	analyticsCode template.HTML // 网站统计分析代码
	shareCode     template.HTML // 分享代码
)

func init() {
	conf.Version = Version

	err := conf.InitConfig("etc/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	analyticsCode = getDefaultCode(conf.Config.AnalyticsFile)

	if conf.Config.DB == "" {
		fmt.Println("数据库地址还没有配置,请到config.json内配置db字段.")
		os.Exit(1)
	}

	session, err := mgo.Dial(conf.Config.DB)
	if err != nil {
		fmt.Println("MongoDB连接失败:", err.Error())
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	db := session.DB("gopher")

	models.DB = conf.Config.DB
	models.PublicSalt = conf.Config.PublicSalt

	utils = &Utils{}

	// 如果没有status,创建
	var status models.Status
	c := db.C(models.STATUS)
	err = c.Find(nil).One(&status)

	if err != nil {
		c.Insert(&models.Status{
			Id_:        bson.NewObjectId(),
			UserCount:  0,
			TopicCount: 0,
			ReplyCount: 0,
			UserIndex:  0,
		})
	}

	// 检查是否有超级账户设置
	var superusers []string
	for _, username := range strings.Split(conf.Config.Superusers, ",") {
		username = strings.TrimSpace(username)
		if username != "" {
			superusers = append(superusers, username)
		}
	}

	if len(superusers) == 0 {
		fmt.Println("你没有设置超级账户,请在config.json中的superusers中设置,如有多个账户,用逗号分开")
	}

	c = db.C(models.USERS)
	var users []models.User
	c.Find(bson.M{"issuperuser": true}).All(&users)

	// 如果mongodb中的超级用户不在配置文件中,取消超级用户
	for _, user := range users {
		if !stringInArray(superusers, user.Username) {
			c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"issuperuser": false}})
		}
	}

	// 设置超级用户
	for _, username := range superusers {
		c.Update(bson.M{"username": username, "issuperuser": false}, bson.M{"$set": bson.M{"issuperuser": true}})
	}
}

func getDefaultCode(path string) (code template.HTML) {
	if path != "" {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			logger.Fatal("文件 " + path + " 没有找到")
		}
		code = template.HTML(string(content))
	}
	return
}

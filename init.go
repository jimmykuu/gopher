package gopher

import (
	"fmt"
	"os"
	"strings"

	"github.com/deferpanic/deferclient/deferstats"
	"github.com/gorilla/sessions"
	. "github.com/qiniu/api.v6/conf"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	parseJsonFile("etc/config.json", &Config)
	analyticsCode = getDefaultCode(Config.AnalyticsFile)
	configGithubAuth()

	if Config.DB == "" {
		fmt.Println("数据库地址还没有配置,请到config.json内配置db字段.")
		os.Exit(1)
	}

	session, err := mgo.Dial(Config.DB)
	if err != nil {
		panic(err)
		fmt.Println("MongoDB连接失败:", err.Error())
		os.Exit(1)
	}

	session.SetMode(mgo.Monotonic, true)

	db := session.DB("gopher")

	store = sessions.NewCookieStore([]byte(Config.CookieSecret))

	utils = &Utils{}

	// 如果没有status,创建
	var status Status
	c := db.C(STATUS)
	err = c.Find(nil).One(&status)

	if err != nil {
		c.Insert(&Status{
			Id_:        bson.NewObjectId(),
			UserCount:  0,
			TopicCount: 0,
			ReplyCount: 0,
			UserIndex:  0,
		})
	}

	// 检查是否有超级账户设置
	var superusers []string
	for _, username := range strings.Split(Config.Superusers, ",") {
		username = strings.TrimSpace(username)
		if username != "" {
			superusers = append(superusers, username)
		}
	}

	if len(superusers) == 0 {
		fmt.Println("你没有设置超级账户,请在config.json中的superusers中设置,如有多个账户,用逗号分开")
	}

	c = db.C(USERS)
	var users []User
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

	// 生成users.json字符串
	generateUsersJson(db)

	if Config.DeferPanicApiKey != "" {
		dps = deferstats.NewClient(Config.DeferPanicApiKey)
	}

	ACCESS_KEY = Config.QiniuAccessKey
	SECRET_KEY = Config.QiniuSecretKey
}

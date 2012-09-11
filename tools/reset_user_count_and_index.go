/*
由于刚开始程序有Bug,导致用户的是第几位和用户数量统计有误
根据用户注册时间重设用户是第几位会员,以及用户数量
*/

package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

var db *mgo.Database

type User struct {
	Id_          bson.ObjectId `bson:"_id"`
	Username     string
	Password     string
	Email        string
	Website      string
	Location     string
	Tagline      string
	Bio          string
	Twitter      string
	Weibo        string
	JoinedAt     time.Time
	Follow       []string
	Fans         []string
	IsSuperuser  bool
	IsActive     bool
	ValidateCode string
	ResetCode    string
	Index        int
}

type Status struct {
	Id_        bson.ObjectId `bson:"_id"`
	UserCount  int
	TopicCount int
	ReplyCount int
	UserIndex  int
}

func init() {
	session, err := mgo.Dial("")
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	db = session.DB("gopher")
}

func main() {
	c := db.C("users")
	var users []User
	c.Find(nil).Sort("createdat").All(&users)

	for index, user := range users {
		println(user.Username, user.Index, index)
		// 更新index
		c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"index": index + 1}})
	}

	userIndex := len(users)
	// 更新统计信息,通过验证的用户才统计总数
	count, err := c.Find(bson.M{"isactive": true}).Count()
	if err != nil {
		panic(err)
	}
	println("total users:", count)

	c = db.C("status")

	c.Update(nil, bson.M{"$set": bson.M{"userindex": userIndex, "usercount": count}})
}

/*
重新设置所有节点的topiccount的值
*/

package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var db *mgo.Database

type Node struct {
	Id_         bson.ObjectId `bson:"_id"`
	Id          string
	Name        string
	Description string
	TopicCount  int
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
	c := db.C("nodes")
	var nodes []Node
	c.Find(nil).All(&nodes)

	c2 := db.C("topics")
	for _, node := range nodes {
		count, err := c2.Find(bson.M{"nodeid": node.Id_}).Count()
		if err != nil {
			panic(err)
		}
		println(node.Name, count)
		c.Update(bson.M{"_id": node.Id_}, bson.M{"$set": bson.M{"topiccount": count}})
	}
}

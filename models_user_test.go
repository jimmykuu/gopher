package gopher

import (
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestAt(t *testing.T) {
	db, _ := mgo.Dial(Config.DB)
	cu := db.DB("gopher").C(USERS)

	defer func() {
		cu.Remove(bson.M{"username": "testUser"})
	}()

	cu.Insert(bson.M{"username": "testUser"})
	u := &User{Username: "testUser"}
	err := u.AtBy(cu, "commentUser", "cid", "ccid")
	if err != nil {
		t.Fatal(err)
	}
	u = new(User)
	cu.Find(bson.M{"username": "testUser"}).One(u)

	if len(u.RecentAts) != 1 {
		t.Fatal("User.AtBy failed")
	}
}

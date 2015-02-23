package gopher

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 用于存储playground产生的代码
type Code struct {
	Id_     bson.ObjectId `bson:"_id"`
	Content string
}

// 保存代码
func (c *Code) Save(db *mgo.Database) error {
	cln := db.C(CODE)
	return cln.Insert(c)
}

// 通过Id获取代码
func GetCodeById(hex string, db *mgo.Database) (*Code, error) {
	cln := db.C(CODE)
	id := bson.ObjectIdHex(hex)
	code := new(Code)
	err := cln.FindId(id).One(code)
	return code, err
}

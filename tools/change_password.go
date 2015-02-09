package main

import (
	"crypto/md5"
	"fmt"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"gopher"
)

func main() {
	session, err := mgo.Dial(gopher.Config.DB)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	fmt.Println(session)

	c := session.DB("gopher").C(gopher.USERS)
	var users []gopher.User
	c.Find(nil).All(&users)

	for _, user := range users {
		salt := strings.Replace(uuid.NewUUID().String(), "-", "", -1)

		saltPasswordMd5 := fmt.Sprintf("%x", md5.Sum([]byte(user.Password+salt)))
		newPassword := fmt.Sprintf("%x", md5.Sum([]byte(saltPasswordMd5+gopher.Config.PublicSalt)))
		err := c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{
			"salt":     salt,
			"password": newPassword,
		}})

		if err != nil {
			panic(err)
		}
	}
}

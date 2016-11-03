package models

import (
	"gopkg.in/mgo.v2"
)

var (
	DB         string
	PublicSalt string
)

func GetSessionAndDB() (*mgo.Session, *mgo.Database) {
	session, err := mgo.Dial(DB)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return session, session.DB("gopher")
}

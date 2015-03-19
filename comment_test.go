package gopher

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	Url "net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Post是辅助函数,适用于一般方法的post,支持url.Values和符合json或者url-encoded格式的io.Reader.
func Post(url string, data interface{}) (*http.Request, error) {
	if data == nil {
		return http.NewRequest("POST", url, nil)
	}
	switch form := data.(type) {
	case Url.Values:
		body := strings.NewReader(form.Encode())
		req, err := http.NewRequest("POST", url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			return nil, err
		}
		return req, err
	case io.Reader:
		bufReader := bufio.NewReader(form)
		b, err := bufReader.Peek(1)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", url, bufReader)
		if err != nil {
			return nil, err
		}
		// json-encoded
		if b[0] == '{' || b[0] == '[' {
			req.Header.Set("Content-Type", "application/json")
		} else {
			// url-encoded
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		return req, nil
	default:
		return nil, fmt.Errorf("data type %#v is not supported.", reflect.TypeOf(data))
	}
}
func TestCommentAt(t *testing.T) {

	db, _ := mgo.Dial(Config.DB)

	//插入一条主题,一个回复.
	//和两个用户.

	topicId := bson.NewObjectId()
	contentId := bson.NewObjectId()
	nodeId := bson.NewObjectId()
	userId := bson.NewObjectId()
	commenterId := bson.NewObjectId()

	//　插入主题
	contentsC := db.DB("gopher").C(CONTENTS)
	contentsC.Insert(&Topic{
		Content: Content{
			Id_:       contentId,
			CreatedBy: userId,
			Type:      TypeTopic,
		},
		Id_:             topicId,
		NodeId:          nodeId,
		LatestRepliedAt: time.Now(),
	},
	)
	defer func() {
		contentsC.Remove(bson.M{"_id": topicId})
	}()

	//  插入用户
	usersC := db.DB("gopher").C(USERS)
	usersC.Insert(bson.M{"_id": userId, "username": "user"})
	usersC.Insert(bson.M{"_id": commenterId, "username": "commenter"})
	usersC.Insert(bson.M{"username": "3rd_user"})
	defer func() {
		usersC.Remove(bson.M{"username": "commenter"})
		usersC.Remove(bson.M{"username": "user"})
		usersC.Remove(bson.M{"username": "3rd_user"})
	}()
	form := Url.Values{}
	form.Add("content", "@3rd_user comments")
	req, err := Post(Config.Host+"/comment/"+topicId.Hex(), form)
	if err != nil {
		t.Fatal(err)
	}
	session, _ := store.Get(req, "user")
	session.Values["username"] = "commenter"
	res := httptest.NewRecorder()
	r := getRoute()
	r.ServeHTTP(res, req)

	defer func() {
		c := db.DB("gopher").C(COMMENTS)
		c.Remove(bson.M{"createdby": commenterId})
	}()

	t.Log(res.Code)
	t.Log(res.Body.String())
	t.Log(res.Header().Get("Location"))
	user, err := getUserByName(usersC, "3rd_user")
	if err != nil {
		t.Fatal(err)
	}

	if len(user.RecentAts) != 1 || user.RecentAts[0].User != "commenter" {
		t.Fatal("at update failed")
	}
}

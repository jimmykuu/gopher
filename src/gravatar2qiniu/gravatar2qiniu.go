/*
把用户的Gravatar头像都下载并上传到七牛云存储
*/

package main

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"crypto/md5"
	"fmt"
	"github.com/jimmykuu/webhelpers"
	. "github.com/qiniu/api/conf"
	qiniu_io "github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
	"gopher"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

func main() {
	ACCESS_KEY = gopher.Config.QiniuAccessKey
	SECRET_KEY = gopher.Config.QiniuSecretKey

	var policy = rs.PutPolicy{
		Scope: "gopher",
	}

	c := gopher.DB.C("users")
	var users []gopher.User
	c.Find(nil).All(&users)

	for _, user := range users {
		url := webhelpers.Gravatar(user.Email, 256)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("get gravatar image error:", url, err.Error())
			return
		}

		filename := strings.Replace(uuid.NewUUID().String(), "-", "", -1) + ".jpg"
		body, _ := ioutil.ReadAll(resp.Body)

		h := md5.New()
		h.Write(body)
		md5Str := fmt.Sprintf("%x", h.Sum(nil))

		if md5Str != "ac83818c6d5b6aca4b6f796b6d3cb338" {
			// 不是默认头像，上传
			key := "avatar/" + filename
			ret := new(qiniu_io.PutRet)

			buf := bytes.NewBuffer(body)

			err = qiniu_io.Put(nil, ret, policy.Token(nil), key, buf, nil)
			if err == nil {
				c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": filename}})
				fmt.Printf("upload %s's avatar success: %s\n", user.Email, filename)
			} else {
				fmt.Printf("upload %s' avatar error: %s\n", user.Email, err.Error())
			}
		}

		resp.Body.Close()
	}
}

/*
把用户的Gravatar头像都下载并上传到七牛云存储
*/

package main

import (
	"crypto/md5"
	"fmt"
	"github.com/jimmykuu/webhelpers"
	. "github.com/qiniu/api/conf"
	qiniu_io "github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
	"gopher"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

func main() {
	ACCESS_KEY = gopher.Config.QiniuAccessKey
	SECRET_KEY = gopher.Config.QiniuSecretKey

	extra := &qiniu_io.PutExtra{
		Bucket:         "gopher",
		MimeType:       "",
		CustomMeta:     "",
		CallbackParams: "",
	}

	var policy = rs.PutPolicy{
		Scope: "gopher",
	}

	c := gopher.DB.C("users")
	var users []gopher.User
	c.Find(nil).Limit(2).All(&users)

	for _, user := range users {
		url := webhelpers.Gravatar(user.Email, 256)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("get gravatar image error:", url, err.Error())
			return
		}

		fmt.Println(url)
		// [inline; filename="8c8dbc14a00702dd50e9fd597c2b4df8.jpg"]
		temp := strings.Split(resp.Header["Content-Disposition"][0], "=")[1]
		filename := string([]byte(temp)[1 : len(temp)-1])
		fmt.Println(filename)

		h := md5.New()
		io.Copy(h, resp.Body)
		md5Str := fmt.Sprintf("%x", h.Sum(nil))

		if md5Str != "ac83818c6d5b6aca4b6f796b6d3cb338" {
			// 不是默认头像，上传
			key := "avatar/" + filename
			ret := new(qiniu_io.PutRet)

			err = qiniu_io.Put(nil, ret, policy.Token(), key, resp.Body, extra)

			if err == nil {
				c.Update(bson.M{"_id": user.Id_}, bson.M{"$set": bson.M{"avatar": filename}})
				fmt.Printf("upload %s's avatar success\n", user.Email)
			} else {
				fmt.Printf("upload %s' avatar error: %s\n", user.Email, err.Error())
			}
		}

		resp.Body.Close()
	}
}

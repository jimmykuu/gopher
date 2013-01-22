/*
一些辅助方法
*/

package gopher

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/smtp"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	PerPage = 20
)

var (
	db          *mgo.Database
	store       *sessions.CookieStore
	fileVersion map[string]string = make(map[string]string) // {path: version}
	utils       *Utils
)

func uuid() string {
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%x", buf)
}

type Utils struct {
}

func (u *Utils) Gravatar(email string, size uint16) string {
	h := md5.New()
	io.WriteString(h, email)
	return fmt.Sprintf("http://www.gravatar.com/avatar/%x?s=%d", h.Sum(nil), size)
}

func (u *Utils) StaticUrl(path string) string {
	version, ok := fileVersion[path]
	if ok {
		return "/static/" + path + "?v=" + version
	}

	file, err := os.Open("static/" + path)

	if err != nil {
		return "/static/" + path
	}

	h := md5.New()

	_, err = io.Copy(h, file)

	version = fmt.Sprintf("%x", h.Sum(nil))[:5]

	fileVersion[path] = version

	return "/static/" + path + "?v=" + version
}

func (u *Utils) Index(index int) int {
	return index + 1
}

func (u *Utils) FormatTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)
	if duration.Seconds() < 60 {
		return fmt.Sprintf("刚刚")
	} else if duration.Minutes() < 60 {
		return fmt.Sprintf("%.0f 分钟前", duration.Minutes())
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%.0f 小时前", duration.Hours())
	}

	return t.Format("2006-01-02 15:04")
}

func (u *Utils) UserInfo(username string) template.HTML {
	c := db.C("users")

	user := User{}
	// 检查用户名
	c.Find(bson.M{"username": username}).One(&user)

	format := `<div>
      <a href="/member/%s"><img class="gravatar" src="%s" style="float:left;"></a>
      <h3><a href="/member/%s">%s</a></h3>
      <div class="clearfix"></div>
    </div>`

	return template.HTML(fmt.Sprintf(format, username, u.Gravatar(user.Email, 50), username, username))
}

func (u *Utils) AssertUser(i interface{}) *User {
	v, _ := i.(User)
	return &v
}

func (u *Utils) AssertNode(i interface{}) *Node {
	v, _ := i.(Node)
	return &v
}

func (u *Utils) AssertTopic(i interface{}) *Topic {
	v, _ := i.(Topic)
	return &v
}

func (u *Utils) AssertArticle(i interface{}) *Article {
	v, _ := i.(Article)
	return &v
}

func (u *Utils) AssertPackage(i interface{}) *Package {
	v, _ := i.(Package)
	return &v
}

func message(w http.ResponseWriter, r *http.Request, title string, message string, class string) {
	renderTemplate(w, r, "message.html", map[string]interface{}{"title": title, "message": template.HTML(message), "class": class})
}

func sendMail(subject string, message string, to []string) {
	auth := smtp.PlainAuth(
		"",
		config["smtp_username"],
		config["smtp_password"],
		config["smtp_host"],
	)
	msg := fmt.Sprintf("To: %s\r\nFrom: jimmykuu@126.com\r\nSubject: %s\r\nContent-Type: text/html\r\n\r\n%s", strings.Join(to, ";"), subject, message)
	err := smtp.SendMail(config["smtp_addr"], auth, config["from_email"], to, []byte(msg))
	if err != nil {
		panic(err)
	}
}

// 检查一个string元素是否在数组里面
func stringInArray(a []string, x string) bool {
	sort.Strings(a)
	index := sort.SearchStrings(a, x)

	if index == 0 {
		if a[0] == x {
			return true
		}

		return false
	} else if index > len(a)-1 {
		return false
	}

	return true
}

func init() {

	if config["db"] == "" {
		fmt.Println("数据库地址还没有配置,请到config.json内配置db字段.")
		os.Exit(1)
	}

	session, err := mgo.Dial(config["db"])
	if err != nil {
		fmt.Println("MongoDB连接失败:", err.Error())
	}

	session.SetMode(mgo.Monotonic, true)

	db = session.DB("gopher")

	cookie_secret := config["cookie_secret"]

	store = sessions.NewCookieStore([]byte(cookie_secret))

	utils = &Utils{}

	// 如果没有status,创建
	var status Status
	c := db.C("status")
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
	for _, username := range strings.Split(config["superusers"], ",") {
		username = strings.TrimSpace(username)
		if username != "" {
			superusers = append(superusers, username)
		}
	}

	if len(superusers) == 0 {
		println("你没有设置超级账户,请在config.json中的superusers中设置,如有多个账户,用逗号分开")
	}

	c = db.C("users")
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
}

func parseTemplate(file string, data map[string]interface{}) []byte {
	var buf bytes.Buffer

	t, err := template.ParseFiles("templates/base.html", "templates/"+file)
	if err != nil {
		panic(err)
	}

	err = t.Execute(&buf, data)

	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func renderTemplate(w http.ResponseWriter, r *http.Request, file string, data map[string]interface{}) {
	_, isPresent := data["signout"]

	// 如果isPresent==true，说明在执行登出操作
	if !isPresent {
		// 加入用户信息
		user, ok := currentUser(r)

		if ok {
			data["username"] = user.Username
			data["isSuperUser"] = user.IsSuperuser
			data["email"] = user.Email
			data["fansCount"] = len(user.Fans)
			data["followCount"] = len(user.Follow)
		}
	}

	data["utils"] = utils

	data["analyticsCode"] = analyticsCode

	page := parseTemplate(file, data)
	w.Write(page)
}

func staticHandler(templateFile string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, r, templateFile, map[string]interface{}{})
	}
}

func yucVerifyFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("7250eb93b3c18cc9daa29cf58af7a004"))
}

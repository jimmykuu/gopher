package main

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

const (
	PerPage = 20
)

var (
	fileVersion map[string]string = make(map[string]string) // {path: version}
	utils       *Utils
	usersJson   []byte
)

type Utils struct {
}

func (u *Utils) UserInfo(username string, db *mgo.Database) template.HTML {
	c := db.C(models.USERS)

	user := models.User{}
	// 检查用户名
	c.Find(bson.M{"username": username}).One(&user)

	format := `<div>
        <a href="/member/%s"><img class="gravatar img-rounded" src="%s" class="gravatar"></a>
        <h4><a href="/member/%s">%s</a><br><small>%s</small></h4>
	<div class="clearfix">
	</div>
    </div>`

	return template.HTML(fmt.Sprintf(format, username, user.AvatarImgSrc(48), username, username, user.Tagline))
}

func (u *Utils) News(username string, db *mgo.Database) template.HTML {
	c := db.C(models.USERS)
	user := models.User{}
	//检查用户名
	c.Find(bson.M{"username": username}).One(&user)
	format := `<div>
		<hr>
		<a href="/member/%s/news#topic">新回复 <span class="badge pull-right">%d</span></a>
		<br>
		<a href="/member/%s/news#at">AT<span class="badge pull-right">%d</span></a>
	</div>
	`
	return template.HTML(fmt.Sprintf(format, username, len(user.RecentReplies), username, len(user.RecentAts)))
}

func (u *Utils) AssertUser(i interface{}) *models.User {
	v, _ := i.(models.User)
	return &v
}

func (u *Utils) AssertNode(i interface{}) *models.Node {
	v, _ := i.(models.Node)
	return &v
}

func (u *Utils) AssertTopic(i interface{}) *models.Topic {
	v, _ := i.(models.Topic)
	return &v
}

func (u *Utils) AssertArticle(i interface{}) *models.Article {
	v, _ := i.(models.Article)
	return &v
}

func (u *Utils) AssertPackage(i interface{}) *models.Package {
	v, _ := i.(models.Package)
	return &v
}

// 获取链接的页码，默认"?p=1"这种类型
func Page(r *http.Request) (int, error) {
	p := r.FormValue("p")
	page := 1

	if p != "" {
		var err error
		page, err = strconv.Atoi(p)

		if err != nil {
			return 0, err
		}
	}

	return page, nil
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

func getPage(r *http.Request) (page int, err error) {
	p := r.FormValue("p")
	page = 1

	if p != "" {
		page, err = strconv.Atoi(p)

		if err != nil {
			return
		}
	}

	return
}

//  提取评论中被at的用户名
func findAts(content string) []string {
	allAts := regexp.MustCompile(`@(\S*) `).FindAllStringSubmatch(content, -1)
	var users []string
	for _, v := range allAts {
		users = append(users, v[1])
	}
	return users
}

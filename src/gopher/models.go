/*
和MongoDB对应的struct
*/

package gopher

import (
	"html/template"
	"labix.org/v2/mgo/bson"
	"time"
)

// 用户
type User struct {
	Id_          bson.ObjectId `bson:"_id"`
	Username     string
	Password     string
	Email        string
	Website      string
	Location     string
	Tagline      string
	Bio          string
	Twitter      string
	Weibo        string
	JoinedAt     time.Time
	Follow       []string
	Fans         []string
	IsSuperuser  bool
	IsActive     bool
	ValidateCode string
	ResetCode    string
	Index        int
}

// 用户发表的最近10个主题
func (u *User) LatestTopics() *[]Topic {
	c := db.C("topics")
	var topics []Topic

	c.Find(bson.M{"userid": u.Id_}).Sort("-createdat").Limit(10).All(&topics)

	return &topics
}

// 用户的最近10个回复
func (u *User) LatestReplies() *[]Reply {
	c := db.C("replies")
	var replies []Reply

	c.Find(bson.M{"userid": u.Id_}).Sort("-createdat").Limit(10).All(&replies)

	return &replies
}

// 是否被某人关注
func (u *User) IsFollowedBy(who string) bool {
	for _, username := range u.Fans {
		if username == who {
			return true
		}
	}

	return false
}

// 是否关注某人
func (u *User) IsFans(who string) bool {
	for _, username := range u.Follow {
		if username == who {
			return true
		}
	}

	return false
}

// 节点
type Node struct {
	Id_         bson.ObjectId `bson:"_id"`
	Id          string
	Name        string
	Description string
	TopicCount  int
}

// 回复
type Reply struct {
	Id_       bson.ObjectId `bson:"_id"`
	UserId    bson.ObjectId
	TopicId   bson.ObjectId
	Markdown  string
	Html      template.HTML
	CreatedAt time.Time
	topic     Topic
}

// 该回复所属于的用户
func (r *Reply) User() *User {
	println(r)
	c := db.C("users")
	var user User
	c.Find(bson.M{"_id": r.UserId}).One(&user)
	return &user
}

// 该回复的主题
func (r *Reply) Topic() *Topic {
	if r.topic.Title == "" {
		c := db.C("topics")
		r.topic = Topic{}
		c.Find(bson.M{"_id": r.TopicId}).One(&r.topic)
	}

	return &r.topic
}

// 是否有权删除回复
func (r *Reply) CanDelete(username string) bool {
	var user User
	c := db.C("users")
	err := c.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	return user.IsSuperuser
}

// 主题
type Topic struct {
	Id_             bson.ObjectId `bson:"_id"`
	NodeId          bson.ObjectId
	UserId          bson.ObjectId
	Title           string
	Markdown        string
	Html            template.HTML
	CreatedAt       time.Time
	ReplyCount      int32
	LatestReplyId   string
	LatestRepliedAt time.Time
	Hits            int32
}

// 主题所属节点
func (t *Topic) Node() *Node {
	c := db.C("nodes")
	node := Node{}
	c.Find(bson.M{"_id": t.NodeId}).One(&node)

	return &node
}

// 主题的最近的一个回复
func (t *Topic) LatestReply() *Reply {
	if t.LatestReplyId == "" {
		return nil
	}

	c := db.C("replies")
	reply := Reply{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(t.LatestReplyId)}).One(&reply)

	if err != nil {
		return nil
	}

	return &reply
}

// 主题的作者
func (t *Topic) User() *User {
	c := db.C("users")
	user := User{}
	c.Find(bson.M{"_id": t.UserId}).One(&user)

	return &user
}

// 主题下的所有回复
func (t *Topic) Replies() *[]Reply {
	c := db.C("replies")
	var replies []Reply

	c.Find(bson.M{"topicid": t.Id_}).All(&replies)

	return &replies
}

// 是否有权编辑主题
func (t *Topic) CanEdit(username string) bool {
	var user User
	c := db.C("users")
	err := c.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	if user.IsSuperuser {
		return true
	}

	return t.UserId == user.Id_
}

// 状态,MongoDB中只存储一个状态
type Status struct {
	Id_        bson.ObjectId `bson:"_id"`
	UserCount  int
	TopicCount int
	ReplyCount int
	UserIndex  int
}

// 站点分类
type SiteCategory struct {
	Id_  bson.ObjectId `bson:"_id"`
	Name string
}

// 分类下的所有站点
func (sc *SiteCategory) Sites() *[]Site {
	var sites []Site
	c := db.C("sites")
	c.Find(bson.M{"categoryid": sc.Id_}).All(&sites)

	return &sites
}

// 站点
type Site struct {
	Id_         bson.ObjectId `bson:"_id"`
	Name        string
	Url         string
	Description string
	CategoryId  bson.ObjectId
	UserId      bson.ObjectId
}

// 是否有权编辑站点
func (s *Site) CanEdit(username string) bool {
	var user User
	c := db.C("users")
	err := c.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	if user.IsSuperuser {
		return true
	}

	return s.UserId == user.Id_
}

// 文章分类
type ArticleCategory struct {
	Id_  bson.ObjectId `bson:"_id"`
	Name string
}

// 文章
type Article struct {
	Id_            bson.ObjectId `bson:"_id"`
	CategoryId     bson.ObjectId
	UserId         bson.ObjectId
	Title          string
	Markdown       string
	Html           template.HTML
	OriginalSource string
	OriginalUrl    string
	CreatedAt      time.Time
	Hits           int32
	Comments       []Comment
}

// 是否有权编辑主题
func (a *Article) CanEdit(username string) bool {
	var user User
	c := db.C("users")
	err := c.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	if user.IsSuperuser {
		return true
	}

	return a.UserId == user.Id_
}

// 文章的提交人
func (a *Article) User() *User {
	c := db.C("users")
	user := User{}
	c.Find(bson.M{"_id": a.UserId}).One(&user)

	return &user
}

// 主题所属类型
func (a *Article) Category() *ArticleCategory {
	c := db.C("articlecategories")
	category := ArticleCategory{}
	c.Find(bson.M{"_id": a.CategoryId}).One(&category)

	return &category
}

// 评论
type Comment struct {
	Id_       bson.ObjectId `bson:"_id"`
	UserId    bson.ObjectId
	Markdown  string
	Html      template.HTML
	CreatedAt time.Time
}

// 评论人
func (c *Comment) User() *User {
	c_ := db.C("users")
	user := User{}
	c_.Find(bson.M{"_id": c.UserId}).One(&user)

	return &user
}

// 是否有权删除评论
func (c *Comment) CanDelete(username string) bool {
	var user User
	c_ := db.C("users")
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	return user.IsSuperuser
}

// 包分类
type PackageCategory struct {
	Id_          bson.ObjectId `bson:"_id"`
	Id           string
	Name         string
	PackageCount int
}

type Package struct {
	Id_        bson.ObjectId `bson:"_id"`
	UserId     bson.ObjectId
	CategoryId bson.ObjectId
	Name       string
	Url        string
	Markdown   string
	Html       template.HTML
	CreatedAt  time.Time
}

func (p *Package) Category() *PackageCategory {
	category := PackageCategory{}
	c := db.C("packagecategories")
	c.Find(bson.M{"_id": p.CategoryId}).One(&category)

	return &category
}

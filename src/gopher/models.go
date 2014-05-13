/*
和MongoDB对应的struct
*/

package gopher

import (
	"html/template"
	"time"

	"labix.org/v2/mgo/bson"
)

const (
	TypeTopic     = 'T'
	TypeArticle   = 'A'
	TypeSite      = 'S'
	TypePackage   = 'P'
	DefaultAvatar = "gopher_teal.jpg"
)

// 用户
type User struct {
	Id_            bson.ObjectId `bson:"_id"`
	Username       string
	Password       string
	Email          string
	Avatar         string
	Website        string
	Location       string
	Tagline        string
	Bio            string
	Twitter        string
	Weibo          string
	GitHubUsername string
	JoinedAt       time.Time
	Follow         []string
	Fans           []string
	//存储的是最近回复的主题的objectid.hex
	RecentReplies []string
	IsSuperuser   bool
	IsActive      bool
	ValidateCode  string
	ResetCode     string
	Index         int
}

// 是否是默认头像
func (u *User) IsDefaultAvatar(avatar string) bool {
	filename := u.Avatar
	if filename == "" {
		filename = DefaultAvatar
	}

	return filename == avatar
}

// 头像的图片地址
func (u *User) AvatarImgSrc() string {
	// 如果没有设置头像，用默认头像
	filename := u.Avatar
	if filename == "" {
		filename = DefaultAvatar
	}

	return "http://gopher.qiniudn.com/avatar/" + filename
}

// 用户发表的最近10个主题
func (u *User) LatestTopics() *[]Topic {
	c := DB.C("contents")
	var topics []Topic

	c.Find(bson.M{"content.createdby": u.Id_, "content.type": TypeTopic}).Sort("-content.createdat").Limit(10).All(&topics)

	return &topics
}

// 用户的最近10个回复
func (u *User) LatestReplies() *[]Comment {
	c := DB.C("comments")
	var replies []Comment

	c.Find(bson.M{"createdby": u.Id_, "type": TypeTopic}).Sort("-createdat").Limit(10).All(&replies)

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

// 通用的内容
type Content struct {
	Id_          bson.ObjectId // 同外层Id_
	Type         int
	Title        string
	Markdown     string
	Html         template.HTML
	CommentCount int
	Hits         int // 点击数量
	CreatedAt    time.Time
	CreatedBy    bson.ObjectId
	UpdatedAt    time.Time
	UpdatedBy    string
}

func (c *Content) Creater() *User {
	c_ := DB.C("users")
	user := User{}
	c_.Find(bson.M{"_id": c.CreatedBy}).One(&user)

	return &user
}

func (c *Content) Updater() *User {
	if c.UpdatedBy == "" {
		return nil
	}

	c_ := DB.C("users")
	user := User{}
	c_.Find(bson.M{"_id": bson.ObjectIdHex(c.UpdatedBy)}).One(&user)

	return &user
}

func (c *Content) Comments() *[]Comment {
	c_ := DB.C("comments")
	var comments []Comment

	c_.Find(bson.M{"contentid": c.Id_}).All(&comments)

	return &comments
}

// 是否有权编辑主题
func (c *Content) CanEdit(username string) bool {
	var user User
	c_ := DB.C("users")
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	if user.IsSuperuser {
		return true
	}

	return c.CreatedBy == user.Id_
}

func (c *Content) CanDelete(username string) bool {
	var user User
	c_ := DB.C("users")
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	return user.IsSuperuser
}

// 主题
type Topic struct {
	Content
	Id_             bson.ObjectId `bson:"_id"`
	NodeId          bson.ObjectId
	LatestReplierId string
	LatestRepliedAt time.Time
}

// 主题所属节点
func (t *Topic) Node() *Node {
	c := DB.C("nodes")
	node := Node{}
	c.Find(bson.M{"_id": t.NodeId}).One(&node)

	return &node
}

// 主题链接
func (t *Topic) Link(createdBy bson.ObjectId) string {
	return "http://golangtc.com/t/" + createdBy.Hex()

}

//格式化日期
func (t *Topic) Format(tm time.Time) string {
	return tm.Format("2006-01-02 15:04:05")
}

// 主题的最近的一个回复
func (t *Topic) LatestReplier() *User {
	if t.LatestReplierId == "" {
		return nil
	}

	c := DB.C("users")
	user := User{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(t.LatestReplierId)}).One(&user)

	if err != nil {
		return nil
	}

	return &user
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
	c := DB.C("contents")
	c.Find(bson.M{"categoryid": sc.Id_, "content.type": TypeSite}).All(&sites)

	return &sites
}

// 站点
type Site struct {
	Content
	Id_        bson.ObjectId `bson:"_id"`
	Url        string
	CategoryId bson.ObjectId
}

// 文章分类
type ArticleCategory struct {
	Id_  bson.ObjectId `bson:"_id"`
	Name string
}

// 文章
type Article struct {
	Content
	Id_            bson.ObjectId `bson:"_id"`
	CategoryId     bson.ObjectId
	OriginalSource string
	OriginalUrl    string
}

// 主题所属类型
func (a *Article) Category() *ArticleCategory {
	c := DB.C("articlecategories")
	category := ArticleCategory{}
	c.Find(bson.M{"_id": a.CategoryId}).One(&category)

	return &category
}

// 评论
type Comment struct {
	Id_       bson.ObjectId `bson:"_id"`
	Type      int
	ContentId bson.ObjectId
	Markdown  string
	Html      template.HTML
	CreatedBy bson.ObjectId
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

// 评论人
func (c *Comment) Creater() *User {
	c_ := DB.C("users")
	user := User{}
	c_.Find(bson.M{"_id": c.CreatedBy}).One(&user)

	return &user
}

// 是否有权删除评论，只允许管理员删除
func (c *Comment) CanDelete(username string) bool {
	var user User
	c_ := DB.C("users")
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	return user.IsSuperuser
}

// 主题
func (c *Comment) Topic() *Topic {
	// 内容
	var topic Topic
	c_ := DB.C("contents")
	c_.Find(bson.M{"_id": c.ContentId, "content.type": TypeTopic}).One(&topic)
	return &topic
}

// 包分类
type PackageCategory struct {
	Id_          bson.ObjectId `bson:"_id"`
	Id           string
	Name         string
	PackageCount int
}

type Package struct {
	Content
	Id_        bson.ObjectId `bson:"_id"`
	CategoryId bson.ObjectId
	Url        string
}

func (p *Package) Category() *PackageCategory {
	category := PackageCategory{}
	c := DB.C("packagecategories")
	c.Find(bson.M{"_id": p.CategoryId}).One(&category)

	return &category
}

type LinkExchange struct {
	Id_         bson.ObjectId `bson:"_id"`
	Name        string        `bson:"name"`
	URL         string        `bson:"url"`
	Description string        `bson:"description"`
	Logo        string        `bson:"logo"`
}

type AD struct {
	Id_      bson.ObjectId `bson:"_id"`
	Position string        `bson:"position"`
	Name     string        `bson:"name"`
	Code     string        `bson:"code"`
}

type Book struct {
	Id_             bson.ObjectId `bson:"_id"`
	Title           string        `bson:"title"`
	Cover           string        `bson:"cover"`
	Author          string        `bson:"author"`
	Translator      string        `bson:"translator"`
	Pages           int           `bson:"pages"`
	Introduction    string        `bson:"introduction"`
	Publisher       string        `bson:"publisher"`
	Language        string        `bson:"language"`
	PublicationDate string        `bson:"publication_date"`
	ISBN            string        `bson:"isbn"`
}

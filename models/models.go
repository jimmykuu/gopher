package models

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 加密密码，md5(md5(password + salt) + public_salt)
func encryptPassword(password, salt string) string {
	saltedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password))) + salt
	md5SaltedPassword := fmt.Sprintf("%x", md5.Sum([]byte(saltedPassword)))
	return fmt.Sprintf("%x", md5.Sum([]byte(md5SaltedPassword+PublicSalt)))
}

// var colors = []string{"#FFCC66", "#66CCFF", "#6666FF", "#FF8000", "#0080FF", "#008040", "#008080"}

// At 主题 id 和评论 id，用于定位到专门的评论
type At struct {
	User      string
	ContentId string
	CommentId string
}

// Reply 主题 id 和主题标题
type Reply struct {
	ContentId  string
	TopicTitle string
}

// CollectTopic 收藏的话题
type CollectTopic struct {
	TopicId       string
	TimeCollected time.Time
}

// Topic 获取收藏对应的话题
func (ct *CollectTopic) Topic(db *mgo.Database) *Topic {
	c := db.C(CONTENTS)
	var topic Topic
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(ct.TopicId), "content.type": TypeTopic}).One(&topic)
	if err != nil {
		panic(err)
		return nil
	}
	return &topic
}

// User 用户
type User struct {
	Id_             bson.ObjectId `bson:"_id"`
	Username        string        //如果关联社区帐号，默认使用社区的用户名
	Password        string
	Salt            string `bson:"salt"`
	Email           string //如果关联社区帐号,默认使用社区的邮箱
	Avatar          string
	Website         string
	Location        string
	Tagline         string
	Bio             string
	Twitter         string
	Weibo           string    // 微博
	GitHubUsername  string    // GitHub 用户名
	JoinedAt        time.Time // 加入时间
	Follow          []string
	Fans            []string
	RecentReplies   []Reply        //存储的是最近回复的主题的objectid.hex
	RecentAts       []At           //存储的是最近评论被AT的主题的objectid.hex
	TopicsCollected []CollectTopic //用户收藏的topic数组
	IsSuperuser     bool           // 是否是超级用户
	IsActive        bool
	IsBlocked       bool `bson:"is_blocked"` // 是否被禁止发帖,回帖
	ValidateCode    string
	ResetCode       string
	Index           int    // 第几个加入社区
	AccountRef      string // 帐号关联的社区
	IdRef           string // 关联社区的帐号
	LinkRef         string // 关联社区的主页链接
	OrgRef          string // 关联社区的组织或者公司
	PictureRef      string // 关联社区的头像链接
	Provider        string // 关联社区名称,比如 github.com
}

// AtBy 增加最近被@
func (u *User) AtBy(c *mgo.Collection, username, contentIdStr, commentIdStr string) error {
	if username == "" || contentIdStr == "" || commentIdStr == "" {
		return errors.New("string parameters can not be empty string")
	}

	if len(u.RecentAts) == 0 {
		var user User
		err := c.Find(bson.M{"username": u.Username}).One(&user)
		if err != nil {
			return err
		}
		u = &user
	}

	u.RecentAts = append(u.RecentAts, At{username, contentIdStr, commentIdStr})
	err := c.Update(bson.M{"username": u.Username}, bson.M{"$set": bson.M{"recentats": u.RecentAts}})
	if err != nil {
		return err
	}
	return nil
}

// IsDefaultAvatar 是否是默认头像
func (u *User) IsDefaultAvatar(avatar string) bool {
	filename := u.Avatar
	if filename == "" {
		filename = DefaultAvatar
	}

	return filename == avatar
}

// CheckPassword 检查密码是否正确
func (u User) CheckPassword(password string) bool {
	return u.Password == encryptPassword(password, u.Salt)
}

// AvatarImgSrc 头像的图片地址
func (u *User) AvatarImgSrc(size int) string {
	// 如果没有设置头像，使用从 http://identicon.relucks.org/ 下载的默认头像
	if u.Avatar == "" {
		return fmt.Sprintf("https://is.golangtc.com/avatar/%s.png?width=%d&height=%d&mode=fill", u.Username, size, size)
	}

	return fmt.Sprintf("https://is.golangtc.com/avatar/%s?width=%d&height=%d&mode=fill", u.Avatar, size, size)
}

// LatestTopics 用户发表的最近10个主题
func (u *User) LatestTopics(db *mgo.Database) *[]Topic {
	c := db.C("contents")
	var topics []Topic

	c.Find(bson.M{"content.createdby": u.Id_, "content.type": TypeTopic}).Sort("-content.createdat").Limit(10).All(&topics)

	return &topics
}

// LatestReplies 用户的最近10个回复
func (u *User) LatestReplies(db *mgo.Database) *[]Comment {
	c := db.C("comments")
	var replies []Comment

	c.Find(bson.M{"createdby": u.Id_, "type": TypeTopic}).Sort("-createdat").Limit(10).All(&replies)

	return &replies
}

// IsFollowedBy 是否被某人关注
func (u *User) IsFollowedBy(who string) bool {
	for _, username := range u.Fans {
		if username == who {
			return true
		}
	}

	return false
}

// IsFans 是否关注某人
func (u *User) IsFans(who string) bool {
	for _, username := range u.Follow {
		if username == who {
			return true
		}
	}

	return false
}

// getUserByName 通过用户名查找用户
func getUserByName(c *mgo.Collection, name string) (*User, error) {
	u := new(User)
	err := c.Find(bson.M{"username": name}).One(u)
	if err != nil {
		return nil, err
	}
	return u, nil

}

// Node 节点
type Node struct {
	Id_         bson.ObjectId `bson:"_id"`
	Id          string
	Name        string
	Description string
	TopicCount  int
}

// Content 通用的内容
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

// Creater 创建人
func (c *Content) Creater(db *mgo.Database) *User {
	c_ := db.C(USERS)
	user := User{}
	c_.Find(bson.M{"_id": c.CreatedBy}).One(&user)

	return &user
}

// Updater 编辑人
func (c *Content) Updater(db *mgo.Database) *User {
	if c.UpdatedBy == "" {
		return nil
	}

	c_ := db.C(USERS)
	user := User{}
	c_.Find(bson.M{"_id": bson.ObjectIdHex(c.UpdatedBy)}).One(&user)

	return &user
}

// Comments 获取评论
func (c *Content) Comments(db *mgo.Database) *[]Comment {
	c_ := db.C("comments")
	var comments []Comment

	c_.Find(bson.M{"contentid": c.Id_}).Sort("createdat").All(&comments)

	return &comments
}

// CanCollect 只能收藏未收藏过的主题
func (c *Content) CanCollect(username string, db *mgo.Database) bool {
	var user User
	c_ := db.C(USERS)
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}
	has := false
	for _, v := range user.TopicsCollected {
		if v.TopicId == c.Id_.Hex() {
			has = true
		}
	}
	return !has
}

// CanEdit 是否有权编辑主题
func (c *Content) CanEdit(username string, db *mgo.Database) bool {
	var user User
	c_ := db.C(USERS)
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	if user.IsSuperuser {
		return true
	}

	return c.CreatedBy == user.Id_
}

// CanDelete 用户是否可删除
func (c *Content) CanDelete(username string, db *mgo.Database) bool {
	var user User
	c_ := db.C("users")
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}

	return user.IsSuperuser
}

// Announcement 公告
type Announcement struct {
	Content
	Id_  bson.ObjectId `bson:"_id"`
	Slug string        `bson:"slug"` // 唯一值，通过 url 路径找到该条内容
}

// Topic 主题
type Topic struct {
	Content
	Id_             bson.ObjectId `bson:"_id"`
	NodeId          bson.ObjectId
	LatestReplierId string
	LatestRepliedAt time.Time
	IsTop           bool `bson:"is_top"` // 置顶
}

// Node 主题所属节点
func (t *Topic) Node(db *mgo.Database) *Node {
	c := db.C("nodes")
	node := Node{}
	c.Find(bson.M{"_id": t.NodeId}).One(&node)

	return &node
}

// Link 主题链接
func (t *Topic) Link(id bson.ObjectId) string {
	return "http://golangtc.com/t/" + id.Hex()

}

// Format 格式化日期
func (t *Topic) Format(tm time.Time) string {
	return tm.Format(time.RFC822)
}

// LatestReplier 主题的最近的一个回复
func (t *Topic) LatestReplier(db *mgo.Database) *User {
	if t.LatestReplierId == "" {
		return nil
	}

	c := db.C("users")
	user := User{}

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(t.LatestReplierId)}).One(&user)

	if err != nil {
		return nil
	}

	return &user
}

// Status 状态, MongoDB 中只存储一个状态
type Status struct {
	Id_        bson.ObjectId `bson:"_id"`
	UserCount  int
	TopicCount int
	ReplyCount int
	UserIndex  int
}

// SiteCategory 站点分类
type SiteCategory struct {
	Id_  bson.ObjectId `bson:"_id"`
	Name string
}

// Sites 分类下的所有站点
func (sc *SiteCategory) Sites(db *mgo.Database) *[]Site {
	var sites []Site
	c := db.C("contents")
	c.Find(bson.M{"categoryid": sc.Id_, "content.type": TypeSite}).All(&sites)

	return &sites
}

// Site 站点
type Site struct {
	Content
	Id_        bson.ObjectId `bson:"_id"`
	Url        string
	CategoryId bson.ObjectId
}

// TrimUrlHttpPrefix 取域名
func (s *Site) TrimUrlHttpPrefix() string {
	return strings.TrimPrefix(s.Url, "http://")
}

// ArticleCategory 文章分类
type ArticleCategory struct {
	Id_  bson.ObjectId `bson:"_id"`
	Name string
}

// Article 文章
type Article struct {
	Content
	Id_            bson.ObjectId `bson:"_id"`
	CategoryId     bson.ObjectId
	OriginalSource string
	OriginalUrl    string
}

// Category 主题所属类型
func (a *Article) Category(db *mgo.Database) *ArticleCategory {
	c := db.C("articlecategories")
	category := ArticleCategory{}
	c.Find(bson.M{"_id": a.CategoryId}).One(&category)

	return &category
}

// Comment 评论
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

// Creater 评论人
func (c *Comment) Creater(db *mgo.Database) *User {
	c_ := db.C("users")
	user := User{}
	c_.Find(bson.M{"_id": c.CreatedBy}).One(&user)

	return &user
}

// CanDeleteOrEdit 是否有权删除评论，管理员和作者可删除
func (c *Comment) CanDeleteOrEdit(username string, db *mgo.Database) bool {
	if c.Creater(db).Username == username {
		return true
	}

	var user User
	c_ := db.C("users")
	err := c_.Find(bson.M{"username": username}).One(&user)
	if err != nil {
		return false
	}
	return user.IsSuperuser
}

// Topic 主题
func (c *Comment) Topic(db *mgo.Database) *Topic {
	// 内容
	var topic Topic
	c_ := db.C("contents")
	c_.Find(bson.M{"_id": c.ContentId, "content.type": TypeTopic}).One(&topic)
	return &topic
}

// PackageCategory 包分类
type PackageCategory struct {
	Id_          bson.ObjectId `bson:"_id"`
	Id           string
	Name         string
	PackageCount int
}

// Package 包
type Package struct {
	Content
	Id_        bson.ObjectId `bson:"_id"`
	CategoryId bson.ObjectId
	Url        string
}

// Category 类目
func (p *Package) Category(db *mgo.Database) *PackageCategory {
	category := PackageCategory{}
	c := db.C("packagecategories")
	c.Find(bson.M{"_id": p.CategoryId}).One(&category)

	return &category
}

// LinkExchange 友链
type LinkExchange struct {
	Id_         bson.ObjectId `bson:"_id"`
	Name        string        `bson:"name"`
	URL         string        `bson:"url"`
	Description string        `bson:"description"`
	Logo        string        `bson:"logo"`
	IsOnHome    bool          `bson:"is_on_home"`   // 是否在首页右侧显示
	IsOnBottom  bool          `bson:"is_on_bottom"` // 是否在底部显示
}

// AD 广告
type AD struct {
	Id_      bson.ObjectId `bson:"_id"`
	Position string        `bson:"position"`
	Name     string        `bson:"name"`
	Code     string        `bson:"code"`
	Index    int           `bons:"index"`
}

// Book 书籍
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

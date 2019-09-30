package gopher

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/conf"
	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/webhelpers"
)

var (
	staticFiles = map[string]string{} // 静态文件的默认属性 {filepath: md5}
)

var Funcs = template.FuncMap{
	"asserttopic": func(i interface{}) *models.Topic {
		v, _ := i.(models.Topic)
		return &v
	},
	"assertuser": func(i interface{}) *models.User {
		v, _ := i.(models.User)
		return &v
	},
	"html": func(text string) template.HTML {
		return template.HTML(text)
	},
	"ads": func(position string, db *mgo.Database) []models.AD {
		c := db.C(models.ADS)
		var ads []models.AD
		c.Find(bson.M{"position": position}).Sort("index").All(&ads)

		count := len(ads)

		if count <= 1 {
			return ads
		}

		dayIndex := time.Now().YearDay() % count

		var sortAds = make([]models.AD, count)
		// 根据当天是一年的内第几天排序，每个广告都有机会排第一个

		for i, ad := range ads {
			sortAds[(i+count-dayIndex)%count] = ad
		}

		return sortAds
	},
	"url": func(url string) string {
		// 没有http://或https://开头的增加http://
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			return url
		}

		return "http://" + url
	},
	"add": func(a, b int) int {
		// 加法运算
		return a + b
	},
	"formatdate": func(t time.Time) string {
		// 格式化日期
		return t.Format(time.RFC822)
	},
	"formattime": func(t time.Time) string {
		// 格式化时间
		now := time.Now()
		duration := now.Sub(t)
		if duration.Seconds() < 60 {
			return fmt.Sprintf("刚刚")
		} else if duration.Minutes() < 60 {
			return fmt.Sprintf("%.0f 分钟前", duration.Minutes())
		} else if duration.Hours() < 24 {
			return fmt.Sprintf("%.0f 小时前", duration.Hours())
		}

		t = t.Add(time.Hour * time.Duration(conf.Config.TimeZoneOffset))
		return t.Format("2006-01-02 15:04")
	},
	"formatdatetime": func(t time.Time) string {
		// 格式化时间成 2006-01-02 15:04:05
		return t.Add(time.Hour * time.Duration(conf.Config.TimeZoneOffset)).Format("2006-01-02 15:04:05")
	},
	"nl2br": func(text string) template.HTML {
		return template.HTML(strings.Replace(text, "\n", "<br>", -1))
	},
	"truncate": func(text string, length int, indicator string) string {
		return webhelpers.Truncate(text, length, indicator)
	},
	"staticfile": func(path string) string {
		// 增加静态文件的版本，防止文件变化后浏览器不更新
		var filepath = filepath.Join("./static", path)
		var md5Str string

		file, err := os.Open(filepath)
		if err != nil {
			if conf.Config.Debug {
				panic(err)
			} else {
				logger.Println("没有找到静态文件", path, err)
			}

			md5Str = "nofile"
		} else {
			var ok bool
			if md5Str, ok = staticFiles[path]; !ok {
				// Debug 状态下，每次都会读取文件的 md5 值，非 Debug状态下，只有第一次读取
				md5h := md5.New()
				io.Copy(md5h, file)
				md5Str = fmt.Sprintf("%x", md5h.Sum([]byte{}))

				if !conf.Config.Debug {
					staticFiles[path] = md5Str
				}
			}
		}

		return fmt.Sprintf("/static/%s?%s", path, md5Str)
	},
}

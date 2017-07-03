package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jimmykuu/webhelpers"
	"github.com/lunny/tango"
	"github.com/tango-contrib/events"
	"github.com/tango-contrib/renders"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
	"github.com/jimmykuu/gopher/modules/static"
	"github.com/jimmykuu/gopher/modules/templates"
)

var (
	logger = log.New(os.Stdout, "[gopher]:", log.LstdFlags)
)

func main() {
	t := tango.Classic()
	t.Use(
		events.Events(),
		static.Static(),
		renders.New(renders.Options{
			Reload:     true,
			FileSystem: templates.FileSystem("templates"),
			Directory:  "templates",
			Funcs: template.FuncMap{
				"asserttopic": func(i interface{}) *models.Topic {
					v, _ := i.(models.Topic)
					return &v
				},
				"html": func(text string) template.HTML {
					return template.HTML(text)
				},
				"loadtimes": func(startTime time.Time) string {
					// 加载时间
					return fmt.Sprintf("%dms", time.Now().Sub(startTime)/1000000)
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

					t = t.Add(time.Hour * time.Duration(Config.TimeZoneOffset))
					return t.Format("2006-01-02 15:04")
				},
				"formatdatetime": func(t time.Time) string {
					// 格式化时间成 2006-01-02 15:04:05
					return t.Add(time.Hour * time.Duration(Config.TimeZoneOffset)).Format("2006-01-02 15:04:05")
				},
				"nl2br": func(text string) template.HTML {
					return template.HTML(strings.Replace(text, "\n", "<br>", -1))
				},
				"truncate": func(text string, length int, indicator string) string {
					return webhelpers.Truncate(text, length, indicator)
				},
				"include": func(filename string, data map[string]interface{}) template.HTML {
					// 加载局部模板，从 templates 中去寻找
					var buf bytes.Buffer
					t, err := template.ParseFiles("templates/" + filename)
					if err != nil {
						panic(err)
					}
					err = t.Execute(&buf, data)
					if err != nil {
						panic(err)
					}

					return template.HTML(buf.Bytes())
				},
			},
		}))

	setRoutes(t)

	t.Run()
}

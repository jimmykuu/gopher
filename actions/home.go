package actions

import (
	"sort"

	"github.com/tango-contrib/renders"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

const PerPage = 20

// Index 首页
type Index struct {
	RenderBase
}

// Get / 默认首页
func (a *Index) Get() error {

	// return a.list(bson.M{"content.type": models.TypeTopic}, "-latestrepliedat")

	page := a.FormInt("p", 1)
	if page <= 0 {
		page = 1
	}

	var nodes []models.Node

	c := a.DB.C(models.NODES)
	c.Find(bson.M{"topiccount": bson.M{"$gt": 0}}).Sort("-topiccount").All(&nodes)

	var status models.Status
	c = a.DB.C(models.STATUS)
	c.Find(nil).One(&status)

	c = a.DB.C(models.CONTENTS)

	var topTopics []models.Topic

	var conditions = bson.M{"content.type": models.TypeTopic}

	if page == 1 {
		c.Find(bson.M{"is_top": true}).Sort("-latestrepliedat").All(&topTopics)

		var objectIds []bson.ObjectId
		for _, topic := range topTopics {
			objectIds = append(objectIds, topic.Id_)
		}
		if len(topTopics) > 0 {
			conditions["_id"] = bson.M{"$not": bson.M{"$in": objectIds}}
		}
	}

	pagination := NewPagination(c.Find(conditions).Sort("-latestrepliedat"), PerPage)

	var topics []models.Topic

	query, err := pagination.Page(page)
	if err != nil {
		return err
	}

	query.(*mgo.Query).All(&topics)

	var linkExchanges []models.LinkExchange
	c = a.DB.C(models.LINK_EXCHANGES)
	c.Find(bson.M{"is_on_home": true}).All(&linkExchanges)

	topics = append(topTopics, topics...)

	c = a.DB.C(models.USERS)

	var cities []City
	c.Pipe([]bson.M{bson.M{
		"$group": bson.M{
			"_id":   "$location",
			"count": bson.M{"$sum": 1},
		},
	}}).All(&cities)

	sort.Sort(ByCount(cities))

	var hotCities []City

	count := 0
	for _, city := range cities {
		if city.Name != "" {
			hotCities = append(hotCities, city)

			count++
		}

		if count == 10 {
			break
		}
	}

	return a.Render("index.html", renders.T{
		"title":         "首页",
		"nodes":         nodes,
		"cities":        hotCities,
		"status":        status,
		"topics":        topics,
		"linkExchanges": linkExchanges,
		"pagination":    pagination,
		"url":           "/",
		"page":          page,
	})
}

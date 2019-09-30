package actions

import (
	"fmt"

	"gitea.com/tango/renders"
	"gopkg.in/mgo.v2/bson"

	"github.com/jimmykuu/gopher/models"
)

// Announcement 显示关于/FAQ等公告页面
type Announcement struct {
	RenderBase
}

// Get /:slug
func (a *Announcement) Get() error {
	slug := a.Param("slug")
	fmt.Println(">>>>", slug)
	var announcement Announcement
	c := a.DB.C(models.CONTENTS)

	err := c.Find(bson.M{"slug": slug, "content.type": models.TypeAnnouncement}).One(&announcement)

	if err != nil {
		a.NotFound(err.Error())
		return nil
	}

	return a.Render("announcement.html", renders.T{
		"announcement": announcement,
	})
}

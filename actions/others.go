package actions

import (
	"github.com/tango-contrib/renders"

	"github.com/jimmykuu/gopher/models"
)

// Link 友情链接
type Link struct {
	RenderBase
}

// Get /link
func (a *Link) Get() error {
	var links []models.LinkExchange
	c := a.DB.C(models.LINK_EXCHANGES)
	c.Find(nil).All(&links)

	return a.Render("others/link.html", renders.T{
		"links": links,
	})
}

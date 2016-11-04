package main

import (
	"github.com/lunny/tango"

	"github.com/jimmykuu/gopher/actions"
	"github.com/jimmykuu/gopher/apis"
)

func setRoutes(t *tango.Tango) {
	t.Any("/signin", new(actions.Signin))
	t.Any("/", new(actions.LatestTopics))

	t.Group("/api", func(g *tango.Group) {
		g.Any("/signin", new(apis.Signin))
	})
}

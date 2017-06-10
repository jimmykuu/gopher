package main

import (
	"github.com/lunny/tango"

	"github.com/jimmykuu/gopher/actions"
	"github.com/jimmykuu/gopher/apis"
)

func setRoutes(t *tango.Tango) {
	t.Any("/signin", new(actions.Signin))
	t.Any("/signup", new(actions.Signup))
	t.Any("/t/:topicID", new(actions.ShowTopic))
	t.Any("/", new(actions.LatestTopics))

	t.Group("/api", func(g *tango.Group) {
		g.Any("/signin", new(apis.Signin))
		g.Any("/signup", new(apis.Signup))
	})
}

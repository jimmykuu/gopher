package main

import (
	"github.com/lunny/tango"

	"github.com/jimmykuu/gopher/actions"
	"github.com/jimmykuu/gopher/apis"
	"github.com/jimmykuu/gopher/middlewares"
)

func setRoutes(t *tango.Tango) {
	t.Any("/signin", new(actions.Signin))
	t.Any("/signup", new(actions.Signup))
	t.Any("/t/:topicID", new(actions.ShowTopic))
	t.Any("/topic/new", new(actions.NewTopic))
	t.Any("/", new(actions.LatestTopics))

	t.Group("/api", func(g *tango.Group) {
		g.Use(middlewares.ApiAuthHandler())
		g.Any("/signin", new(apis.Signin))
		g.Any("/signup", new(apis.Signup))
		g.Any("/nodes", new(apis.NodeList))
		g.Post("/topic/new", new(apis.NewTopic))
	})
}

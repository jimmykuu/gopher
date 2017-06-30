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
	t.Any("/t/:topicId", new(actions.ShowTopic))
	t.Get("/topic/new", new(actions.NewTopic))
	t.Get("/t/:topicId/edit", new(actions.EditTopic))
	t.Any("/", new(actions.LatestTopics))

	t.Group("/api", func(g *tango.Group) {
		g.Use(middlewares.ApiAuthHandler())
		g.Any("/signin", new(apis.Signin))
		g.Any("/signup", new(apis.Signup))
		g.Any("/nodes", new(apis.NodeList))
		g.Post("/topic/new", new(apis.NewTopic))
		g.Get("/topic/:topicId", new(apis.GetTopic))
		g.Post("/topic/:topicId/edit", new(apis.EditTopic))
	})
}

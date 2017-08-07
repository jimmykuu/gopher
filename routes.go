package main

import (
	"github.com/lunny/tango"

	"github.com/jimmykuu/gopher/actions"
	"github.com/jimmykuu/gopher/apis"
	"github.com/jimmykuu/gopher/middlewares"
)

func setRoutes(t *tango.Tango) {
	t.Get("/signin", new(actions.Signin))
	t.Get("/signup", new(actions.Signup))
	t.Get("/t/:topicID", new(actions.ShowTopic))
	t.Get("/topic/new", new(actions.NewTopic))
	t.Get("/t/:topicID/edit", new(actions.EditTopic))
	t.Get("/go/:node", new(actions.NodeTopics))
	t.Get("/topics/latest", new(actions.LatestTopics))
	t.Get("/topics/no_reply", new(actions.NoReplyTopics))
	t.Get("/search", new(actions.SearchTopic))
	t.Get("/member/:username", new(actions.AccountIndex))
	t.Get("/", new(actions.LatestReplyTopics))

	t.Group("/api", func(g *tango.Group) {
		g.Use(middlewares.ApiAuthHandler())
		g.Post("/signin", new(apis.Signin))
		g.Post("/signup", new(apis.Signup))
		g.Get("/nodes", new(apis.NodeList))
		g.Post("/topics", new(apis.Topic))
		g.Put("/topics/:topicID", new(apis.Topic))
		g.Get("/topics/:topicID", new(apis.Topic))
		g.Delete("/topic/:topicID", new(apis.Topic))
		g.Post("/comment/:contentId", new(apis.Comment))
		g.Post("/upload/image", new(apis.UploadImage))
		g.Delete("/comment/:commentId", new(apis.Comment))
		g.Get("/comment/:commentId", new(apis.Comment))
		g.Put("/comments/:commentId", new(apis.Comment))
	})
}

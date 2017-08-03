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
	t.Get("/t/:topicId", new(actions.ShowTopic))
	t.Get("/topic/new", new(actions.NewTopic))
	t.Get("/t/:topicId/edit", new(actions.EditTopic))
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
		g.Post("/topic/new", new(apis.NewTopic))
		g.Get("/topic/:topicId", new(apis.GetTopic))
		g.Post("/topic/:topicId/edit", new(apis.EditTopic))
		g.Post("/comment/:contentId", new(apis.Comment))
		g.Post("/upload/image", new(apis.UploadImage))
		g.Delete("/comment/:commentId", new(apis.DeleteComment))
	})
}

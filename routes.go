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

	t.Get("/topic/new", new(actions.NewTopic))
	t.Get("/t/:topicID/edit", new(actions.EditTopic))
	t.Get("/t/:topicID", new(actions.ShowTopic))
	t.Get("/t/:topicID/collect", new(actions.CollectTopic))

	t.Get("/", new(actions.Index))
	t.Get("/go/:node", new(actions.NodeTopics))
	t.Get("/search", new(actions.SearchTopic))

	t.Get("/member/:username", new(actions.AccountIndex))
	t.Get("/member/:username/comments", new(actions.AccountComments))
	t.Get("/member/:username/collections", new(actions.AccountCollections))
	t.Get("/members", new(actions.LatestUsers))
	t.Get("/members/all", new(actions.AllUsers))

	t.Get("/:slug", new(actions.Announcement))

	t.Get("/user_center", new(actions.UserCenter))
	t.Get("/user_center/profile", new(actions.UserProfile))
	t.Get("/user_center/avatar", new(actions.ChangeAvatar))
	t.Get("/user_center/change_password", new(actions.ChangePassword))

	t.Group("/api", func(g *tango.Group) {
		g.Use(middlewares.ApiAuthHandler())
		g.Post("/signin", new(apis.Signin))
		g.Post("/signup", new(apis.Signup))

		g.Get("/nodes", new(apis.NodeList))
		g.Post("/topics", new(apis.Topic))
		g.Put("/topics/:topicID", new(apis.Topic))
		g.Get("/topics/:topicID", new(apis.Topic))
		g.Delete("/topic/:topicID", new(apis.Topic))

		g.Get("/topic/:topicID/collect", new(apis.CollectTopic))
		g.Get("/topic/:topicID/cancel_collect", new(apis.CancelCollectTopic))

		g.Post("/upload/image", new(apis.UploadImage))

		g.Post("/comments", new(apis.Comment))
		g.Delete("/comments/:commentID", new(apis.Comment))
		g.Get("/comments/:commentID", new(apis.Comment))
		g.Put("/comments/:commentID", new(apis.Comment))

		g.Get("/user_center/default_avatars", new(apis.DefaultAvatars))
		g.Get("/user_center/user_info", new(apis.UserInfo))
		g.Put("/user_center/profile", new(apis.UserProfile))
		g.Put("/user_center/change_password", new(apis.UserChangePassword))
		g.Post("/user_center/upload_avatar", new(apis.UploadAvatarImage))
		g.Put("/user_center/set_avatar", new(apis.SetAvatar))
		g.Get("/user_center/from_gravatar", new(apis.FromGravatar))
	})
}

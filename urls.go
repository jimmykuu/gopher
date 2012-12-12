/*
URL和Handler的Mapping
*/

package main

import (
	"net/http"
)

var (
	handlers = map[string]func(http.ResponseWriter, *http.Request){
		"/":      indexHandler,
		"/about": staticHandler("about.html"),
		"/faq":   staticHandler("faq.html"),

		"/admin":                        adminHandler,
		"/admin/nodes":                  adminListNodesHandler,
		"/admin/node/new":               adminNewNodeHandler,
		"/admin/site_categories":        adminListSiteCategoriesHandler,
		"/admin/site_category/new":      adminNewSiteCategoryHandler,
		"/admin/users":                  adminListUsersHandler,
		"/admin/user/{userId}/activate": adminActivateUserHandler,
		"/admin/article_categories":     adminListArticleCategoriesHandler,
		"/admin/article_category/new":   adminNewArticleCategoryHandler,

		"/signup":          signupHandler,
		"/signin":          signinHandler,
		"/signout":         signoutHandler,
		"/activate/{code}": activateHandler,
		"/forgot_password": forgotPasswordHandler,
		"/reset/{code}":    resetPasswordHandler,
		"/profile":         profileHandler,
		// "/profile/avatar":  changeAvatarHandler,

		"/nodes":     nodesHandler,
		"/go/{node}": topicInNodeHandler,

		"/topic/new":        newTopicHandler,
		"/new/{node}":       newTopicHandler,
		"/t/{topicId}":      showTopicHandler,
		"/t/{topicId}/edit": editTopicHandler,
		"/reply/{topicId}":  replyHandler,

		"/member/{username}":         memberInfoHandler,
		"/member/{username}/topics":  memberTopicsHandler,
		"/member/{username}/replies": memberRepliesHandler,
		"/follow/{username}":         followHandler,
		"/unfollow/{username}":       unfollowHandler,

		"/sites":              sitesHandler,
		"/site/new":           newSiteHandler,
		"/site/{siteId}/edit": editSiteHandler,

		"/article/new":   newArticleHandler,
		"/articles":      listArticlesHandler,
		"/a/{articleId}": showArticleHandler,
	}
)

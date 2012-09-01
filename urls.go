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

		"/admin":                   adminHandler,
		"/admin/nodes":             adminListNodesHandler,
		"/admin/node/new":          adminNewNodeHandler,
		"/admin/site_categories":   adminListSiteCategoriesHandler,
		"/admin/site_category/new": adminNewSiteCategoryHandler,

		"/signup":          signupHandler,
		"/signin":          signinHandler,
		"/signout":         signoutHandler,
		"/activate/{code}": activateHandler,
		"/forgot_password": forgotPasswordHandler,
		"/reset/{code}":    resetPasswordHandler,
		"/profile":         profileHandler,
		// "/profile/avatar":  changeAvatarHandler,

		"/topic/new":       newTopicHandler,
		"/nodes":           nodesHandler,
		"/go/{node}":       topicInNodeHandler,
		"/new/{node}":      newTopicHandler,
		"/t/{topicId}":     showTopicHandler,
		"/reply/{topicId}": replyHandler,

		"/member/{username}":         memberInfoHandler,
		"/member/{username}/topics":  memberTopicsHandler,
		"/member/{username}/replies": memberRepliesHandler,
		"/follow/{username}":         followHandler,
		"/unfollow/{username}":       unfollowHandler,

		"/sites":    sitesHandler,
		"/site/new": newSiteHandler,
	}
)

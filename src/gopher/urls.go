/*
URL和Handler的Mapping
*/

package gopher

import (
	"net/http"
)

var (
	handlers = map[string]func(http.ResponseWriter, *http.Request){
		"/":                    indexHandler,
		"/about":               staticHandler("about.html"),
		"/faq":                 staticHandler("faq.html"),
		"/yuc_verify_file.txt": yucVerifyFileHandler,

		"/admin":                        adminHandler,
		"/admin/nodes":                  adminListNodesHandler,
		"/admin/node/new":               adminNewNodeHandler,
		"/admin/site_categories":        adminListSiteCategoriesHandler,
		"/admin/site_category/new":      adminNewSiteCategoryHandler,
		"/admin/users":                  adminListUsersHandler,
		"/admin/user/{userId}/activate": adminActivateUserHandler,
		"/admin/article_categories":     adminListArticleCategoriesHandler,
		"/admin/article_category/new":   adminNewArticleCategoryHandler,
		"/admin/package_categories":     adminListPackageCategoriesHandler,
		"/admin/package_category/new":   adminNewPackageCategoryHandler,

		"/signup":          signupHandler,
		"/signin":          signinHandler,
		"/signout":         signoutHandler,
		"/activate/{code}": activateHandler,
		"/forgot_password": forgotPasswordHandler,
		"/reset/{code}":    resetPasswordHandler,
		"/profile":         profileHandler,
		"/change_password": changePasswordHandler,
		"/profile/avatar":  changeAvatarHandler,

		"/nodes":     nodesHandler,
		"/go/{node}": topicInNodeHandler,

		"/comment/{contentId}":        commentHandler,
		"/comment/{commentId}/delete": deleteCommentHandler,

		"/topic/new":        newTopicHandler,
		"/new/{node}":       newTopicHandler,
		"/t/{topicId}":      showTopicHandler,
		"/t/{topicId}/edit": editTopicHandler,

		"/member/{username}":         memberInfoHandler,
		"/member/{username}/topics":  memberTopicsHandler,
		"/member/{username}/replies": memberRepliesHandler,
		"/follow/{username}":         followHandler,
		"/unfollow/{username}":       unfollowHandler,

		"/sites":                sitesHandler,
		"/site/new":             newSiteHandler,
		"/site/{siteId}/edit":   editSiteHandler,
		"/site/{siteId}/delete": deleteSiteHandler,

		"/article/new":        newArticleHandler,
		"/articles":           listArticlesHandler,
		"/a/{articleId}":      showArticleHandler,
		"/a/{articleId}/edit": editArticleHandler,

		"/packages":              packagesHandler,
		"/package/new":           newPackageHandler,
		"/packages/{categoryId}": listPackagesHandler,
		"/p/{packageId}":         showPackageHandler,
		"/p/{packageId}/edit":    editPackageHandler,
	}
)

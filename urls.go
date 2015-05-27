/*
URL和Handler的Mapping
*/

package gopher

import (
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
)

// NewHandler返回含有请求上下文的Handler.
func NewHandler(w http.ResponseWriter, r *http.Request) *Handler {
	session, err := mgo.Dial(Config.DB)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return &Handler{
		ResponseWriter: w,
		Request:        r,
		StartTime:      time.Now(),
		Session:        session,
		DB:             session.DB("gopher"),
	}
}

// Redirect是重定向的简便方法.
func (h Handler) Redirect(urlStr string) {
	http.Redirect(h.ResponseWriter, h.Request, urlStr, http.StatusFound)
}

// HandlerFunc 是自定义的请求处理函数,接受*Handler作为参数.
type HandlerFunc func(*Handler)

// Route 是代表对应请求的路由规则以及权限的结构体.
type Route struct {
	URL         string
	Permission  PerType
	HandlerFunc HandlerFunc
}

func fileHandler(w http.ResponseWriter, req *http.Request) {
	url := req.Method + " " + req.URL.Path
	logger.Println(url)
	filePath := req.URL.Path[1:]
	http.ServeFile(w, req, filePath)
}

// 路由规则
var routes = []Route{
	{"/", Everyone, indexHandler},
	{"/about", Everyone, staticHandler("about.html")},
	{"/faq", Everyone, staticHandler("faq.html")},
	{"/timeline", Everyone, staticHandler("timeline.html")},
	{"/link", Everyone, linksHandler},
	{"/search", Everyone, searchHandler},
	{"/users.json", Everyone, usersJsonHandler},

	{"/topics.rss", Everyone, rssHandler},
	{"/admin", Administrator, adminHandler},
	{"/admin/nodes", Administrator, adminListNodesHandler},
	{"/admin/node/new", Administrator, adminNewNodeHandler},
	{"/admin/site_categories", Administrator, adminListSiteCategoriesHandler},
	{"/admin/site_category/new", Administrator, adminNewSiteCategoryHandler},
	{"/admin/users", Administrator, adminListUsersHandler},
	{"/admin/user/{userId}/activate", Administrator, adminActivateUserHandler},
	{"/admin/article_categories", Administrator, adminListArticleCategoriesHandler},
	{"/admin/article_category/new", Administrator, adminNewArticleCategoryHandler},
	{"/admin/package_categories", Administrator, adminListPackageCategoriesHandler},
	{"/admin/package_category/new", Administrator, adminNewPackageCategoryHandler},
	{"/admin/package_category/{id}/edit", Administrator, adminEditPackageCategoryHandler},
	{"/admin/link_exchanges", Administrator, adminListLinkExchangesHandler},
	{"/admin/link_exchange/new", Administrator, adminNewLinkExchangeHandler},
	{"/admin/link_exchange/{linkExchangeId}/edit", Administrator, adminEditLinkExchangeHandler},
	{"/admin/link_exchange/{linkExchangeId}/delete", Administrator, adminDeleteLinkExchangeHandler},
	{"/admin/ads", Administrator, adminListAdsHandler},
	{"/admin/ad/new", Administrator, adminNewAdHandler},
	{"/admin/ad/{id:[0-9a-f]{24}}/delete", Administrator, adminDeleteAdHandler},
	{"/admin/ad/{id:[0-9a-f]{24}}/edit", Administrator, adminEditAdHandler},
	{"/admin/book/new", Administrator, newBookHandler},
	{"/admin/books", Administrator, listBooksHandler},
	{"/admin/book/{id}/edit", Administrator, editBookHandler},
	{"/admin/book/{id}/delete", Administrator, deleteBookHandler},
	{"/admin/top/topics", Administrator, listTopTopicsHandler},
	{"/admin/topic/{id:[0-9a-f]{24}}/cancel/top", Administrator, cancelTopTopicHandler},
	{"/admin/topic/{id:[0-9a-f]{24}}/set/top", Administrator, setTopTopicHandler},

	{"/auth/signup", Everyone, authSignupHandler},
	{"/auth/login", Everyone, authLoginHandler},
	{"/signup", Everyone, signupHandler},
	{"/signin", Everyone, signinHandler},
	{"/signout", Authenticated, signoutHandler},
	{"/activate/{code}", Everyone, activateHandler},
	{"/forgot_password", Everyone, forgotPasswordHandler},
	{"/reset/{code}", Everyone, resetPasswordHandler},
	{"/profile", Authenticated, profileHandler},
	{"/change_password", Authenticated, changePasswordHandler},
	{"/profile/avatar", Authenticated, changeAvatarHandler},
	{"/profile/choose_default_avatar", Authenticated, chooseDefaultAvatar},
	{"/profile/avatar/gravatar", Authenticated, setAvatarFromGravatar},

	{"/nodes", Everyone, nodesHandler},
	{"/go/{node}", Everyone, topicInNodeHandler},

	{"/comment/{contentId:[0-9a-f]{24}}", Authenticated, commentHandler},
	{"/comment/{commentId:[0-9a-f]{24}}/delete", Administrator, deleteCommentHandler},
	{"/comment/{id:[0-9a-f]{24}}.json", Authenticated, commentJsonHandler},
	{"/comment/{id:[0-9a-f]{24}}/edit", Authenticated, editCommentHandler},

	{"/topics/latest", Everyone, latestTopicsHandler},
	{"/topics/no_reply", Everyone, noReplyTopicsHandler},
	{"/topic/new", Authenticated, newTopicHandler},
	{"/new/{node}", Authenticated, newTopicHandler},
	{"/t/{topicId:[0-9a-f]{24}}", Everyone, showTopicHandler},
	{"/t/{topicId:[0-9a-f]{24}}/edit", Authenticated, editTopicHandler},
	{"/t/{topicId:[0-9a-f]{24}}/collect", Authenticated, collectTopicHandler},
	{"/t/{topicId:[0-9a-f]{24}}/delete", Administrator, deleteTopicHandler},

	{"/member/{username}", Everyone, memberInfoHandler},
	{"/member/{username}/topics", Everyone, memberTopicsHandler},
	{"/member/{username}/replies", Everyone, memberRepliesHandler},
	{"/member/{username}/news", Everyone, memmberNewsHandler},
	{"/member/{username}/clear/{t}", Authenticated, memmberNewsClear},
	{"/member/{username}/collect", Everyone, memberTopicsCollectedHandler},
	{"/follow/{username}", Authenticated, followHandler},
	{"/unfollow/{username}", Authenticated, unfollowHandler},
	{"/members", Everyone, membersHandler},
	{"/members/all", Everyone, allMembersHandler},
	{"/members/city/{cityName}", Everyone, membersInTheSameCityHandler},

	{"/sites", Everyone, sitesHandler},
	{"/site/new", Authenticated, newSiteHandler},
	{"/site/{siteId:[0-9a-f]{24}}/edit", Authenticated, editSiteHandler},
	{"/site/{siteId:[0-9a-f]{24}}/delete", Administrator, deleteSiteHandler},
	{"/article/new", Authenticated, newArticleHandler},
	{"/article/go/{categoryId}", Everyone, articlesInCategoryHandler},
	{"/articles", Everyone, listArticlesHandler},
	{"/a/{articleId}", Everyone, showArticleHandler},
	{"/a/{articleId}/redirect", Everyone, redirectArticleHandler},
	{"/a/{articleId}/edit", Authenticated, editArticleHandler},
	{"/a/{articleId}/delete", Authenticated, deleteArticleHandler},

	{"/packages", Everyone, packagesHandler},
	{"/package/new", Authenticated, newPackageHandler},
	{"/package", Everyone, getPackageUrlHandler},
	{"/packages/{categoryId}", Everyone, listPackagesHandler},
	{"/p/{packageId}", Everyone, showPackageHandler},
	{"/p/{packageId}/edit", Authenticated, editPackageHandler},
	{"/p/{packageId}/delete", Administrator, deletePackageHandler},

	{"/books", Everyone, booksHandler},
	{"/book/{id}", Everyone, showBookHandler},

	{"/download", Everyone, downloadHandler},
	{"/download/package", Everyone, downloadPackagesHandler},
	{"/download/liteide", Everyone, downloadLiteIDEHandler},
	{"/play", Everyone, playGroundHandler},
	{"/play/{id:[0-9a-f]{24}}", Everyone, playGroundHandler},
	{"/play/share", Everyone, shareCodeHandler},
	{"/playsocket", Everyone, playSocketHandler},

	{"/upload/image", Authenticated, uploadImageHandler},

	{"/api/v1/topics", Everyone, apiTopicsHandler},
}

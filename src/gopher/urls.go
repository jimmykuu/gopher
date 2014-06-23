/*
URL和Handler的Mapping
*/

package gopher

import (
	"net/http"
	"time"

	"labix.org/v2/mgo"
)

type Handler struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	StartTime      time.Time
	Session        *mgo.Session
	DB             *mgo.Database
}

func NewHandler(w http.ResponseWriter, r *http.Request) Handler {
	session, err := mgo.Dial(Config.DB)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return Handler{
		ResponseWriter: w,
		Request:        r,
		StartTime:      time.Now(),
		Session:        session,
		DB:             session.DB("gopher"),
	}
}

type HandlerFunc func(Handler)

type Route struct {
	URL         string
	Permission  int
	HandlerFunc HandlerFunc
}

var (
	routes = []Route{
		{"/", Everyone, indexHandler},
		{"/about", Everyone, staticHandler("about.html")},
		{"/faq", Everyone, staticHandler("faq.html")},
		{"/timeline", Everyone, staticHandler("timeline.html")},
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
		{"/admin/ad/{id}/delete", Administrator, adminDeleteAdHandler},
		{"/admin/ad/{id}/edit", Administrator, adminEditAdHandler},
		{"/admin/book/new", Administrator, newBookHandler},
		{"/admin/books", Administrator, listBooksHandler},
		{"/admin/book/{id}/edit", Administrator, editBookHandler},
		{"/admin/book/{id}/delete", Administrator, deleteBookHandler},

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

		{"/nodes", Everyone, nodesHandler},
		{"/go/{node}", Everyone, topicInNodeHandler},

		{"/comment/{contentId}", Authenticated, commentHandler},
		{"/comment/{commentId}/delete", Administrator, deleteCommentHandler},

		{"/topics/latest", Everyone, latestTopicsHandler},
		{"/topics/no_reply", Everyone, noReplyTopicsHandler},
		{"/topic/new", Authenticated, newTopicHandler},
		{"/new/{node}", Authenticated, newTopicHandler},
		{"/t/{topicId}", Everyone, showTopicHandler},
		{"/t/{topicId}/edit", Authenticated, editTopicHandler},
		{"/t/{topicId}/delete", Administrator, deleteTopicHandler},

		{"/member/{username}", Everyone, memberInfoHandler},
		{"/member/{username}/topics", Everyone, memberTopicsHandler},
		{"/member/{username}/replies", Everyone, memberRepliesHandler},
		{"/member/{username}/news", Everyone, memmberNewsHandler},
		{"/follow/{username}", Authenticated, followHandler},
		{"/unfollow/{username}", Authenticated, unfollowHandler},
		{"/members", Everyone, membersHandler},
		{"/members/all", Everyone, allMembersHandler},
		{"/members/city/{cityName}", Everyone, membersInTheSameCityHandler},

		{"/sites", Everyone, sitesHandler},
		{"/site/new", Authenticated, newSiteHandler},
		{"/site/{siteId}/edit", Authenticated, editSiteHandler},
		{"/site/{siteId}/delete", Administrator, deleteSiteHandler},
		{"/article/new", Authenticated, newArticleHandler},
		{"/article/go/{categoryId}", Everyone, articlesInCategoryHandler},
		{"/articles", Everyone, listArticlesHandler},
		{"/a/{articleId}", Everyone, showArticleHandler},
		{"/a/{articleId}/redirect", Everyone, redirectArticleHandler},
		{"/a/{articleId}/edit", Authenticated, editArticleHandler},
		{"/a/{articleId}/delete", Authenticated, deleteArticleHandler},

		{"/packages", Everyone, packagesHandler},
		{"/package/new", Authenticated, newPackageHandler},
		{"/packages/{categoryId}", Everyone, listPackagesHandler},
		{"/p/{packageId}", Everyone, showPackageHandler},
		{"/p/{packageId}/edit", Authenticated, editPackageHandler},
		{"/p/{packageId}/delete", Administrator, deletePackageHandler},

		{"/books", Everyone, booksHandler},
		{"/book/{id}", Everyone, showBookHandler},

		{"/download", Everyone, downloadHandler},
	}
)

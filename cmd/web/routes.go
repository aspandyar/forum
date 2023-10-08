package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/showAll", app.allForum)

	mux.HandleFunc("/forum/view/", app.forumView)

	mux.HandleFunc("/forum/category", app.forumCategory)

	mux.HandleFunc("/user/signup", app.userSignup)
	mux.HandleFunc("/auth", app.handleGoogleAuth)
	mux.HandleFunc("/callback", app.handleGoogleCallback)

	mux.HandleFunc("/login/github/", app.gitHubLoginHandler)
	mux.HandleFunc("/loggedin", func(w http.ResponseWriter, r *http.Request) {
		app.loggedinHandler(w, r, "")
	})
	mux.HandleFunc("/login/github/callback", app.gitHubCallbackHandler)

	mux.HandleFunc("/user/login", app.userLogin)

	userLogout := http.HandlerFunc(app.userLogoutPost)
	mux.Handle("/user/logout", app.requireAuthentication(userLogout))

	moderAsk := http.HandlerFunc(app.moderAskHandler)
	mux.Handle("/moderation/ask", app.requireAuthentication(moderAsk))

	moderAccept := http.HandlerFunc(app.userModerationDone)
	mux.Handle("/moderation/accept/", app.requireAuthentication(moderAccept))

	forumAccept := http.HandlerFunc(app.forumAcceptHandler)
	mux.Handle("/moderation/forum/", app.requireAuthentication(forumAccept))

	forumCreate := http.HandlerFunc(app.handleForumCreate)
	mux.Handle("/forum/create", app.requireAuthentication(forumCreate))

	forumEdit := http.HandlerFunc(app.handleForumEdit)
	mux.Handle("/forum/edit/", app.requireAuthentication(forumEdit))

	forumRemove := http.HandlerFunc(app.handleForumRemove)
	mux.Handle("/forum/remove/", app.requireAuthentication(forumRemove))

	forumLikeStatus := http.HandlerFunc(app.forumIsLike)
	mux.Handle("/forum/like/", app.requireAuthentication(forumLikeStatus))

	forumCommentStatus := http.HandlerFunc(app.handleForumComment)
	mux.Handle("/forum/comment/", app.requireAuthentication(forumCommentStatus))

	forumCommentEditStatus := http.HandlerFunc(app.handleForumEditComment)
	mux.Handle("/forum/comment/edit/", app.requireAuthentication(forumCommentEditStatus))

	forumRemoveCommentStatus := http.HandlerFunc(app.ForumRemoveCommentPost)
	mux.Handle("/forum/comment/remove/", app.requireAuthentication(forumRemoveCommentStatus))

	forumLikeCommentStatus := http.HandlerFunc(app.forumIsLikeComment)
	mux.Handle("/forum/likeComment/", app.requireAuthentication(forumLikeCommentStatus))

	forumAllLikes := http.HandlerFunc(app.forumAllUserLikes)
	mux.Handle("/forum/allLikes", app.requireAuthentication(forumAllLikes))

	forumAllComments := http.HandlerFunc(app.forumAllUserComments)
	mux.Handle("/forum/all_comments", app.requireAuthentication(forumAllComments))

	forumAllPosts := http.HandlerFunc(app.forumAllUserPosts)
	mux.Handle("/forum/allPosts", app.requireAuthentication(forumAllPosts))

	userNotificationSection := http.HandlerFunc(app.userNotification)
	mux.Handle("/user/notification", app.requireAuthentication(userNotificationSection))

	userNotificationSectionRemove := http.HandlerFunc(app.userNotificationRemove)
	mux.Handle("/user/notification/remove/", app.requireAuthentication(userNotificationSectionRemove))

	return app.recoverPanic(app.logRequest(secureHeaders(rateLimitMiddleware(mux))))
}

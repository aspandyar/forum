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

	forumCreate := http.HandlerFunc(app.handleForumCreate)
	mux.Handle("/forum/create", app.requireAuthentication(forumCreate))

	forumEdit := http.HandlerFunc(app.handleForumEdit)
	mux.Handle("/forum/edit/", app.requireAuthentication(forumEdit))

	userLogout := http.HandlerFunc(app.userLogoutPost)
	mux.Handle("/user/logout", app.requireAuthentication(userLogout))

	forumLikeStatus := http.HandlerFunc(app.forumIsLike)
	mux.Handle("/forum/like/", app.requireAuthentication(forumLikeStatus))

	forumCommentStatus := http.HandlerFunc(app.handleForumComment)
	mux.Handle("/forum/comment/", app.requireAuthentication(forumCommentStatus))

	forumLikeCommentStatus := http.HandlerFunc(app.forumIsLikeComment)
	mux.Handle("/forum/likeComment/", app.requireAuthentication(forumLikeCommentStatus))

	forumAllLikes := http.HandlerFunc(app.forumAllUserLikes)
	mux.Handle("/forum/allLikes", app.requireAuthentication(forumAllLikes))

	forumAllPosts := http.HandlerFunc(app.forumAllUserPosts)
	mux.Handle("/forum/allPosts", app.requireAuthentication(forumAllPosts))

	return app.recoverPanic(app.logRequest(secureHeaders(rateLimitMiddleware(mux))))
}

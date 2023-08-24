package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home)

	mux.HandleFunc("/forum/view/", app.forumView)

	mux.HandleFunc("/user/signup", app.userSignup)
	mux.HandleFunc("/user/login", app.userLogin)

	forumCreate := http.HandlerFunc(app.handleForumCreate)
	mux.Handle("/forum/create", app.requireAuthentication(forumCreate))

	userLogout := http.HandlerFunc(app.userLogoutPost)
	mux.Handle("/user/logout", app.requireAuthentication(userLogout))

	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}

package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/forum/view/", app.forumView)
	mux.HandleFunc("/forum/create", app.handleForumCreate)

	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
package main

import "net/http"

func (app *application) moderAskHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.AskForModeration(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

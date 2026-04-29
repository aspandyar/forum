package main

import (
	"net/http"
	"strconv"
	"strings"
)

func (app *application) handleForumRemove(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "remove" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[3]
	forumID, err := strconv.Atoi(idStr)
	if err != nil || forumID < 1 {
		http.NotFound(w, r)
		return
	}

	userFromForum, err := app.forums.GetUserIDFromForum(forumID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	isOwn := app.isOwnForum(userFromForum, r)
	if !isOwn {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.forums.Remove(forumID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/showAll", http.StatusSeeOther)
}

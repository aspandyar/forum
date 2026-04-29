package main

import (
	"net/http"

	requestpkg "github.com/aspandyar/forum/internal/transport/http/request"
)

func pathParts(r *http.Request) []string {
	return requestpkg.PathParts(r)
}

func pathInt(parts []string, idx int) (int, error) {
	return requestpkg.PathInt(parts, idx)
}

func (app *application) sessionUserID(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, err
	}
	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

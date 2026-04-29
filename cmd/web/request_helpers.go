package main

import (
	"net/http"
	"strconv"
	"strings"
)

func pathParts(r *http.Request) []string {
	return strings.Split(r.URL.Path, "/")
}

func pathInt(parts []string, idx int) (int, error) {
	return strconv.Atoi(parts[idx])
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

package models

import (
	"net/http"
	"time"

	sessioncookie "github.com/aspandyar/forum/internal/transport/http/sessioncookie"
)

func SetSessionCookie(w http.ResponseWriter, token string, expiration time.Time) {
	sessioncookie.SetSessionCookie(w, token, expiration)
}

func ClearSessionCookie(w http.ResponseWriter) {
	sessioncookie.ClearSessionCookie(w)
}

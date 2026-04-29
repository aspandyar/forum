package sessioncookie

import (
	"net/http"
	"time"
)

func SetSessionCookie(w http.ResponseWriter, token string, expiration time.Time) {
	sessionCookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Expires:  expiration,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, sessionCookie)
}

func ClearSessionCookie(w http.ResponseWriter) {
	expiredCookie := &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, expiredCookie)
}

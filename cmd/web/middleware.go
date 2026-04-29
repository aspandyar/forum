package main

import (
	"net/http"
	"time"

	mw "github.com/aspandyar/forum/internal/transport/http/middleware"
)

var (
	limit  = 100
	period = time.Minute
)

var rateLimiter = mw.RateLimit(limit, period)

func rateLimitMiddleware(next http.Handler) http.Handler {
	return rateLimiter(next)
}

func secureHeaders(next http.Handler) http.Handler {
	return mw.SecureHeaders(next)
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return mw.LogRequest(app.infoLog)(next)
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return mw.RecoverPanic(app.serverError)(next)
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return mw.RequireAuthentication(app.isAuthenticated)(next)
}

package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	mw "github.com/aspandyar/forum/internal/transport/http/middleware"
)

func TestSecureHeadersSetsExpectedHeaders(t *testing.T) {
	h := secureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req, rr := newRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Header().Get("Referrer-Policy") == "" {
		t.Fatal("expected Referrer-Policy header")
	}
	if rr.Header().Get("X-Frame-Options") != "deny" {
		t.Fatalf("unexpected X-Frame-Options: %q", rr.Header().Get("X-Frame-Options"))
	}
}

func TestRateLimitMiddlewareBlocksAfterLimit(t *testing.T) {
	origLimit, origPeriod, origLimiter := limit, period, rateLimiter
	limit = 1
	period = time.Hour
	rateLimiter = mw.RateLimit(limit, period)
	t.Cleanup(func() {
		limit = origLimit
		period = origPeriod
		rateLimiter = origLimiter
	})

	h := rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1, rr1 := newRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusNoContent {
		t.Fatalf("first request code=%d, want %d", rr1.Code, http.StatusNoContent)
	}

	req2, rr2 := newRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request code=%d, want %d", rr2.Code, http.StatusTooManyRequests)
	}
}

func TestRecoverPanicReturnsServerError(t *testing.T) {
	app, _ := newWebTestApp(t)
	h := app.recoverPanic(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req, rr := newRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if rr.Header().Get("Connection") != "close" {
		t.Fatalf("expected Connection close header, got %q", rr.Header().Get("Connection"))
	}
}

func TestRequireAuthenticationRedirectsWithoutSession(t *testing.T) {
	app, _ := newWebTestApp(t)
	h := app.requireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req, rr := newRequest(http.MethodGet, "/protected", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusSeeOther)
	}
	if !strings.Contains(rr.Header().Get("Location"), "/user/login") {
		t.Fatalf("expected redirect to login, got %q", rr.Header().Get("Location"))
	}
}

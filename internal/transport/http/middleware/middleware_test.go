package middleware

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	h := RateLimit(1, time.Hour)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:2222"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusNoContent {
		t.Fatalf("first code=%d", rr1.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:2222"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second code=%d", rr2.Code)
	}
}

func TestSecureHeadersAndRequireAuthentication(t *testing.T) {
	secured := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rr := httptest.NewRecorder()
	secured.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Header().Get("X-Frame-Options") != "deny" {
		t.Fatalf("header mismatch: %q", rr.Header().Get("X-Frame-Options"))
	}

	authd := RequireAuthentication(func(*http.Request) bool { return false })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rr = httptest.NewRecorder()
	authd.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect for unauthenticated, got %d", rr.Code)
	}
}

func TestLogRequestAndRecoverPanic(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	logged := LogRequest(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rr := httptest.NewRecorder()
	logged.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("log middleware status=%d", rr.Code)
	}

	var captured error
	recovered := RecoverPanic(func(w http.ResponseWriter, err error) {
		captured = err
		http.Error(w, "boom", http.StatusInternalServerError)
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("panic-now")
	}))
	rr = httptest.NewRecorder()
	recovered.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("recover status=%d", rr.Code)
	}
	if captured == nil || !strings.Contains(captured.Error(), "panic-now") {
		t.Fatalf("expected captured panic error, got %v", captured)
	}
}

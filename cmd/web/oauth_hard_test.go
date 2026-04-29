package main

import (
	"errors"
	"net/http"
	"testing"
)

type failingRoundTripper struct{}

func (f failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("forced transport error")
}

func TestOAuthCallbacks_ErrorBranches(t *testing.T) {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = failingRoundTripper{}
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	app, _ := newWebTestApp(t)

	req, rr := newRequest(http.MethodGet, "/callback?code=abc", nil)
	app.handleGoogleCallback(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("handleGoogleCallback forced error status=%d", rr.Code)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic in gitHubCallbackHandler when transport fails")
		}
	}()
	req, rr = newRequest(http.MethodGet, "/login/github/callback?code=abc", nil)
	app.gitHubCallbackHandler(rr, req)
}

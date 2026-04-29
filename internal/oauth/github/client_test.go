package github

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withDefaultTransport(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	old := http.DefaultTransport
	http.DefaultTransport = rt
	t.Cleanup(func() { http.DefaultTransport = old })
}

func TestAccessTokenAndUserData(t *testing.T) {
	withDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.String(), "/access_token") {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"abc123"}`)),
			}, nil
		}
		if strings.Contains(req.URL.String(), "api.github.com/user") {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"id":1,"login":"octo"}`)),
			}, nil
		}
		t.Fatalf("unexpected url: %s", req.URL.String())
		return nil, nil
	}))

	token, err := AccessToken("cid", "secret", "code")
	if err != nil {
		t.Fatalf("AccessToken error: %v", err)
	}
	if token != "abc123" {
		t.Fatalf("token=%q, want abc123", token)
	}
	data, err := UserData(token)
	if err != nil {
		t.Fatalf("UserData error: %v", err)
	}
	if !strings.Contains(data, "octo") {
		t.Fatalf("unexpected user data: %q", data)
	}
}

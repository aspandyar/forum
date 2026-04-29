package google

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

func TestFetchUserInfo(t *testing.T) {
	withDefaultTransport(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(req.URL.String(), "accounts.google.com/o/oauth2/token"):
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"token123"}`)),
			}, nil
		case strings.Contains(req.URL.String(), "www.googleapis.com/oauth2/v2/userinfo"):
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"email":"u@example.com","name":"User"}`)),
			}, nil
		default:
			t.Fatalf("unexpected URL: %s", req.URL.String())
			return nil, nil
		}
	}))

	info, err := FetchUserInfo("cid", "secret", "http://localhost/cb", "code")
	if err != nil {
		t.Fatalf("FetchUserInfo error: %v", err)
	}
	if info.Email != "u@example.com" || info.Name != "User" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

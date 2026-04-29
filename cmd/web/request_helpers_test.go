package main

import (
	"net/http"
	"testing"
)

func TestPathHelpersAndSessionUserID(t *testing.T) {
	app, _ := newWebTestApp(t)
	userID := seedWebUser(t, app, "sess", "sess@example.com", 2)

	req, _ := newRequest(http.MethodGet, "/forum/view/123", nil)
	parts := pathParts(req)
	if len(parts) < 4 {
		t.Fatalf("unexpected path parts: %v", parts)
	}
	id, err := pathInt(parts, 3)
	if err != nil || id != 123 {
		t.Fatalf("pathInt id=%d err=%v", id, err)
	}

	if _, err := app.sessionUserID(req); err == nil {
		t.Fatal("expected missing session cookie error")
	}
	attachSessionCookie(t, app, req, userID)
	gotUser, err := app.sessionUserID(req)
	if err != nil {
		t.Fatalf("sessionUserID error: %v", err)
	}
	if gotUser != userID {
		t.Fatalf("userID=%d want=%d", gotUser, userID)
	}
}

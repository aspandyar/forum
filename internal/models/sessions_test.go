package models

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateSessionID(t *testing.T) {
	a := generateSessionID()
	b := generateSessionID()
	if a == "" || b == "" {
		t.Fatal("expected non-empty session IDs")
	}
	if a == b {
		t.Fatal("expected unique session IDs")
	}
}

func TestCreateSessionReplacesExistingSession(t *testing.T) {
	db := newTestDB(t)
	model := &SessionModel{DB: db}
	u := seedUser(t, db, "sess-user")

	first, err := model.CreateSession(u)
	if err != nil {
		t.Fatalf("first CreateSession: %v", err)
	}
	second, err := model.CreateSession(u)
	if err != nil {
		t.Fatalf("second CreateSession: %v", err)
	}
	if first.Token == second.Token {
		t.Fatal("expected new session token when replacing active session")
	}

	var oldExpiryRaw string
	if err := db.QueryRow(`SELECT expiry FROM sessions WHERE token = ?`, first.Token).Scan(&oldExpiryRaw); err != nil {
		t.Fatalf("query old session expiry: %v", err)
	}
	oldExpiry, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", oldExpiryRaw)
	if err != nil {
		t.Fatalf("parse old session expiry: %v", err)
	}
	if oldExpiry.After(time.Now().Add(2 * time.Second)) {
		t.Fatalf("expected old session invalidated, got expiry %v", oldExpiry)
	}
}

func TestGetSessionAndActiveSession(t *testing.T) {
	db := newTestDB(t)
	model := &SessionModel{DB: db}
	u := seedUser(t, db, "active-user")

	active, err := model.GetActiveSession(u)
	if err != nil {
		t.Fatalf("GetActiveSession without session: %v", err)
	}
	if active != nil {
		t.Fatal("expected nil active session initially")
	}

	created, err := model.CreateSession(u)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	gotUser, expiry, err := model.GetSession(created.Token)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if gotUser != u {
		t.Fatalf("GetSession user = %d, want %d", gotUser, u)
	}
	if expiry.IsZero() {
		t.Fatal("expected non-zero session expiry")
	}

	active, err = model.GetActiveSession(u)
	if err != nil {
		t.Fatalf("GetActiveSession with session: %v", err)
	}
	if active == nil || active.Token != created.Token {
		t.Fatalf("unexpected active session: %#v", active)
	}
}

func TestSessionCookieHelpers(t *testing.T) {
	rr := httptest.NewRecorder()
	exp := time.Now().Add(10 * time.Minute)
	SetSessionCookie(rr, "token123", exp)

	headers := rr.Result().Header.Values("Set-Cookie")
	if len(headers) == 0 {
		t.Fatal("expected Set-Cookie header")
	}
	cookieHeader := strings.Join(headers, " ")
	if !strings.Contains(cookieHeader, "session=token123") {
		t.Fatalf("expected session cookie token, got %q", cookieHeader)
	}

	rr = httptest.NewRecorder()
	ClearSessionCookie(rr)
	headers = rr.Result().Header.Values("Set-Cookie")
	if len(headers) == 0 || !strings.Contains(headers[0], "session=") {
		t.Fatalf("expected cleared session cookie, got %v", headers)
	}
}

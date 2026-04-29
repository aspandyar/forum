package main

import (
	"net/http"
	"testing"
)

func TestMethodSwitchHandlersRejectUnsupportedMethods(t *testing.T) {
	app, _ := newWebTestApp(t)

	tests := []struct {
		name    string
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"addTags", "/admin/addTags", app.addTagsHandler},
		{"forumReport", "/moderation/report/1", app.forumReportHandler},
		{"forumCreate", "/forum/create", app.handleForumCreate},
		{"forumEdit", "/forum/edit/1", app.handleForumEdit},
		{"userSignup", "/user/signup", app.userSignup},
		{"userLogin", "/user/login", app.userLogin},
		{"forumComment", "/forum/comment/1", app.handleForumComment},
		{"forumEditComment", "/forum/comment/edit/1/1", app.handleForumEditComment},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, rr := newRequest(http.MethodDelete, tc.path, nil)
			tc.handler(rr, req)
			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status=%d, want %d", rr.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestAuthenticationAndOwnershipHelpers(t *testing.T) {
	app, _ := newWebTestApp(t)
	adminID := seedWebUser(t, app, "admin", "admin@example.com", 4)
	userID := seedWebUser(t, app, "user", "user@example.com", 2)

	req, _ := newRequest(http.MethodGet, "/", nil)
	if app.isAuthenticated(req) {
		t.Fatal("expected unauthenticated without session cookie")
	}
	if role := app.getRole(req); role != guestRole {
		t.Fatalf("role=%d, want guest role=%d", role, guestRole)
	}

	attachSessionCookie(t, app, req, userID)
	if !app.isAuthenticated(req) {
		t.Fatal("expected authenticated request with valid session")
	}
	if role := app.getRole(req); role != 2 {
		t.Fatalf("role=%d, want 2", role)
	}
	if !app.isOwnForum(userID, req) {
		t.Fatal("expected user to own their own forum")
	}
	if app.isOwnForum(adminID, req) {
		t.Fatal("expected regular user not to own admin forum")
	}

	reqAdmin, _ := newRequest(http.MethodGet, "/", nil)
	attachSessionCookie(t, app, reqAdmin, adminID)
	if !app.isOwnForum(userID, reqAdmin) {
		t.Fatal("expected admin to be considered owner for moderation checks")
	}
}

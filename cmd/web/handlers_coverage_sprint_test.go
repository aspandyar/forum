package main

import (
	"net/http"
	"strings"
	"testing"
	"text/template"
)

func addBaseTemplate(app *application, name string) {
	app.tempalteCache[name] = mustTemplate(`{{define "base"}}ok{{end}}`)
}

func TestListingHandlers_GuardBranches(t *testing.T) {
	app, _ := newWebTestApp(t)

	req, rr := newRequest(http.MethodGet, "/wrong", nil)
	app.home(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("home wrong path=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/wrong-show", nil)
	app.allForum(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("allForum wrong path=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/wrong-posts", nil)
	app.forumAllUserPosts(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("allPosts wrong path=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/wrong-likes", nil)
	app.forumAllUserLikes(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("allLikes wrong path=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/wrong-comments", nil)
	app.forumAllUserComments(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("allComments wrong path=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/wrong-category", nil)
	app.forumCategory(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("category wrong path=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/view", nil)
	app.forumView(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("forumView invalid path=%d", rr.Code)
	}
}

func TestListingHandlers_SimpleSuccessBranches(t *testing.T) {
	app, db := newWebTestApp(t)
	userID := seedWebUser(t, app, "u1", "u1@example.com", 2)
	_, err := app.forums.Insert("title", "body", "go", 7, userID, "")
	if err != nil {
		t.Fatalf("insert forum: %v", err)
	}
	addBaseTemplate(app, "category.tmpl.html")
	addBaseTemplate(app, "allForums.tmpl.html")
	addBaseTemplate(app, "allComments.tmpl.html")

	req, rr := newRequest(http.MethodGet, "/forum/category", nil)
	app.forumCategory(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("category GET status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/category", strings.NewReader("tags=go"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.forumCategory(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("category POST status=%d", rr.Code)
	}

	attachSessionCookie(t, app, req, userID)
	req, rr = newRequest(http.MethodGet, "/forum/allPosts", nil)
	attachSessionCookie(t, app, req, userID)
	app.forumAllUserPosts(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("allPosts status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/allLikes", nil)
	attachSessionCookie(t, app, req, userID)
	app.forumAllUserLikes(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("allLikes status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/all_comments", nil)
	attachSessionCookie(t, app, req, userID)
	app.forumAllUserComments(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("allComments status=%d", rr.Code)
	}

	_ = db
}

func TestAuthAndOAuthSimpleHandlers(t *testing.T) {
	app, _ := newWebTestApp(t)

	req, rr := newRequest(http.MethodGet, "/user/signup", nil)
	app.userSignupGet(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("signup get expected render error status 500, got %d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/user/login", nil)
	app.userLoginGet(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("login get expected render error status 500, got %d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/user/logout", nil)
	app.userLogoutPost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("logout missing cookie status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/auth", nil)
	app.handleGoogleAuth(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("google auth redirect status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/login/github/", nil)
	app.gitHubLoginHandler(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("github login redirect status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/loggedin", nil)
	app.loggedinHandler(rr, req, "")
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("loggedin empty data status=%d", rr.Code)
	}

	pass, err := generateRandomPassword(8)
	if err != nil || pass == "" {
		t.Fatalf("generateRandomPassword err=%v pass=%q", err, pass)
	}
}

func TestModerationAndForumGuardBranches(t *testing.T) {
	app, _ := newWebTestApp(t)

	req, rr := newRequest(http.MethodGet, "/moderation/ask", nil)
	app.moderAskHandler(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("moderAsk missing cookie status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/remove/bad", nil)
	app.handleForumRemove(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("forum remove bad path status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/user/notification/remove/abc", nil)
	app.userNotificationRemove(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("notification remove missing cookie status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/user/report/remove/1/2/3", nil)
	app.forumReportRemoveHandler(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("report remove missing cookie status=%d", rr.Code)
	}
}

func mustTemplate(src string) *template.Template {
	tpl, err := template.New("base").Parse(src)
	if err != nil {
		panic(err)
	}
	return tpl
}

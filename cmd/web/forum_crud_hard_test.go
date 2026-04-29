package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestForumCreateGetAndPostHardBranches(t *testing.T) {
	app, _ := newWebTestApp(t)
	addBaseTemplate(app, "create.tmpl.html")

	req, rr := newRequest(http.MethodGet, "/forum/create", nil)
	app.ForumCreateGet(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("ForumCreateGet status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/create", strings.NewReader("%%%"))
	req.Header.Set("Content-Type", "multipart/form-data")
	app.ForumCreatePost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ForumCreatePost malformed multipart status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/create", strings.NewReader("expires=bad"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ForumCreatePost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ForumCreatePost bad expires status=%d", rr.Code)
	}
}

func TestForumEditGetAndPostHardBranches(t *testing.T) {
	app, _ := newWebTestApp(t)
	addBaseTemplate(app, "edit.tmpl.html")
	addBaseTemplate(app, "create.tmpl.html")

	req, rr := newRequest(http.MethodGet, "/forum/edit/abc", nil)
	app.ForumEditGet(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("ForumEditGet bad id status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/edit/abc", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ForumEditPost(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("ForumEditPost bad id status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/edit/1", strings.NewReader("%%%"))
	req.Header.Set("Content-Type", "multipart/form-data")
	app.ForumEditPost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ForumEditPost malformed multipart status=%d", rr.Code)
	}
}

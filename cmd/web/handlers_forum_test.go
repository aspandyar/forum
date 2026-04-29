package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForumIsLikeRejectsNonPost(t *testing.T) {
	app := &application{}
	req := httptest.NewRequest(http.MethodGet, "/forum/like/1", nil)
	rr := httptest.NewRecorder()

	app.forumIsLike(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestForumIsLikeRejectsInvalidButton(t *testing.T) {
	app := &application{}
	req := httptest.NewRequest(http.MethodPost, "/forum/like/1", strings.NewReader("button=invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	app.forumIsLike(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestForumIsLikeCommentRejectsInvalidPath(t *testing.T) {
	app := &application{}
	req := httptest.NewRequest(http.MethodPost, "/forum/notLikeComment/1", strings.NewReader("button=like"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	app.forumIsLikeComment(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestForumCommentPostRejectsNonPost(t *testing.T) {
	app := &application{}
	req := httptest.NewRequest(http.MethodGet, "/forum/comment/1", nil)
	rr := httptest.NewRecorder()

	app.ForumCommentPost(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestForumEditCommentPostRejectsNonPost(t *testing.T) {
	app := &application{}
	req := httptest.NewRequest(http.MethodGet, "/forum/comment/edit/1/1", nil)
	rr := httptest.NewRecorder()

	app.ForumEditCommentPost(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

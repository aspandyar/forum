package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestForumInteractionEditRemoveGuardBranches(t *testing.T) {
	app, db := newWebTestApp(t)
	uid := seedWebUser(t, app, "intu", "intu@example.com", 2)
	forumID, err := app.forums.Insert("t", "c", "go", 7, uid, "")
	if err != nil {
		t.Fatalf("insert forum: %v", err)
	}
	_, err = app.forumComment.CommentPost(forumID, uid, "hello")
	if err != nil {
		t.Fatalf("comment post: %v", err)
	}
	var commentID int
	if err := db.QueryRow(`SELECT id FROM forum_comments WHERE forum_id=?`, forumID).Scan(&commentID); err != nil {
		t.Fatalf("comment id: %v", err)
	}

	req, rr := newRequest(http.MethodGet, "/forum/comment/edit/x/y", nil)
	app.ForumEditCommentGet(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("ForumEditCommentGet bad path status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/comment/edit/"+strings.TrimSpace("1")+"/bad", strings.NewReader("comment=ok"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ForumEditCommentPost(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("ForumEditCommentPost bad comment id status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/comment/remove/bad/path", nil)
	app.ForumRemoveCommentPost(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("ForumRemoveCommentPost invalid ids should early return 200, got %d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/comment/remove/"+strings.TrimSpace("1")+"/"+strings.TrimSpace("1"), nil)
	app.ForumRemoveCommentPost(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("ForumRemoveCommentPost no session ownership status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/comment/1", strings.NewReader("comment="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ForumCommentPost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ForumCommentPost blank comment status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/forum/comment/edit/1/1", strings.NewReader("comment="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ForumEditCommentPost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ForumEditCommentPost without session status=%d", rr.Code)
	}
}

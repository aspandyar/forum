package main

import (
	"net/http"
	"strconv"
	"testing"
)

func TestForumView_NoSessionBadSessionNoRowsAndNoRecord(t *testing.T) {
	app, db := newWebTestApp(t)
	addBaseTemplate(app, "view.tmpl.html")

	userID := seedWebUser(t, app, "viewu", "viewu@example.com", 2)
	forumID, err := app.forums.Insert("view title", "view body", "go", 7, userID, "")
	if err != nil {
		t.Fatalf("insert forum: %v", err)
	}
	if err := app.forums.ChangeForumStatus(forumID, 1); err != nil {
		t.Fatalf("publish forum: %v", err)
	}

	req, rr := newRequest(http.MethodGet, "/forum/view/"+itoa(forumID), nil)
	app.forumView(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("forumView no session status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/view/"+itoa(forumID), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-token"})
	app.forumView(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("forumView invalid session token status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/forum/view/999999", nil)
	app.forumView(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("forumView missing row status=%d", rr.Code)
	}

	_, err = db.Exec(`INSERT INTO forums(title,content,tags,user_id,created,expires,image_path,status) VALUES('expired','x','go',?,datetime('now'),datetime('now','-1 day'),'',1)`, userID)
	if err != nil {
		t.Fatalf("insert expired forum: %v", err)
	}
	var expiredID int
	if err := db.QueryRow(`SELECT id FROM forums WHERE title='expired'`).Scan(&expiredID); err != nil {
		t.Fatalf("expired id: %v", err)
	}
	req, rr = newRequest(http.MethodGet, "/forum/view/"+itoa(expiredID), nil)
	app.forumView(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("forumView expired no record status=%d", rr.Code)
	}
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

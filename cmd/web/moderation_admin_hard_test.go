package main

import (
	"net/http"
	"strconv"
	"testing"
)

func TestModerationAdminHandlers_Matrix(t *testing.T) {
	app, db := newWebTestApp(t)
	addBaseTemplate(app, "notification.tmpl.html")
	addBaseTemplate(app, "report.tmpl.html")
	addBaseTemplate(app, "addTags.tmpl.html")

	adminID := seedWebUser(t, app, "adminx", "adminx@example.com", 4)
	userID := seedWebUser(t, app, "userx", "userx@example.com", 2)
	forumID, err := app.forums.Insert("title", "body", "go", 7, userID, "")
	if err != nil {
		t.Fatalf("insert forum: %v", err)
	}

	_, err = db.Exec(`INSERT INTO forum_notifications(user_name, body, status, forum_link, user_id, user_not_id) VALUES ('u','b','admin', ?, ?, ?)`, forumID, adminID, userID)
	if err != nil {
		t.Fatalf("seed notification: %v", err)
	}
	var notificationID int
	if err := db.QueryRow(`SELECT id FROM forum_notifications ORDER BY id DESC LIMIT 1`).Scan(&notificationID); err != nil {
		t.Fatalf("notification id: %v", err)
	}

	req, rr := newRequest(http.MethodGet, "/user/notification", nil)
	app.userNotification(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("userNotification missing cookie status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/user/notification", nil)
	attachSessionCookie(t, app, req, adminID)
	app.userNotification(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("userNotification admin status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/user/notification/remove/"+strconv.Itoa(notificationID), nil)
	attachSessionCookie(t, app, req, adminID)
	app.userNotificationRemove(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("userNotificationRemove status=%d", rr.Code)
	}

	_, _ = db.Exec(`INSERT INTO forum_notifications(user_name, body, status, forum_link, user_id, user_not_id) VALUES ('u','b','admin', ?, ?, ?)`, forumID, adminID, userID)
	if err := db.QueryRow(`SELECT id FROM forum_notifications ORDER BY id DESC LIMIT 1`).Scan(&notificationID); err != nil {
		t.Fatalf("notification id 2: %v", err)
	}
	req, rr = newRequest(http.MethodGet, "/moderation/accept/"+strconv.Itoa(notificationID)+"/"+strconv.Itoa(userID), nil)
	app.userModerationDone(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("userModerationDone status=%d", rr.Code)
	}

	_, _ = db.Exec(`INSERT INTO forum_notifications(user_name, body, status, forum_link, user_id, user_not_id) VALUES ('u','b','admin', ?, ?, ?)`, forumID, adminID, userID)
	if err := db.QueryRow(`SELECT id FROM forum_notifications ORDER BY id DESC LIMIT 1`).Scan(&notificationID); err != nil {
		t.Fatalf("notification id 3: %v", err)
	}
	req, rr = newRequest(http.MethodGet, "/moderation/denote/"+strconv.Itoa(userID)+"/"+strconv.Itoa(notificationID), nil)
	app.moderDenoteHandler(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("moderDenoteHandler status=%d", rr.Code)
	}

	_, _ = db.Exec(`INSERT INTO forum_notifications(user_name, body, status, forum_link, user_id, user_not_id) VALUES ('u','b','moder', ?, ?, ?)`, forumID, adminID, userID)
	if err := db.QueryRow(`SELECT id FROM forum_notifications ORDER BY id DESC LIMIT 1`).Scan(&notificationID); err != nil {
		t.Fatalf("notification id 4: %v", err)
	}
	req, rr = newRequest(http.MethodGet, "/moderation/forum/"+strconv.Itoa(notificationID)+"/"+strconv.Itoa(forumID), nil)
	app.forumAcceptHandler(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("forumAcceptHandler status=%d", rr.Code)
	}
}

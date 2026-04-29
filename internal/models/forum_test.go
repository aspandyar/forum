package models

import (
	"sort"
	"testing"
)

func TestContainsAny(t *testing.T) {
	if !containsAny([]string{"go", "web"}, []string{"db", "go"}) {
		t.Fatal("expected overlap to return true")
	}
	if containsAny([]string{"a"}, []string{"b"}) {
		t.Fatal("expected non-overlap to return false")
	}
}

func TestForumInsertEditAndRemove(t *testing.T) {
	db := newTestDB(t)
	m := &ForumModel{DB: db}

	u1 := seedUser(t, db, "harry")
	forumID, err := m.Insert("title", "body", "go, web", 7, u1, "/img.png")
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if err := m.ChangeForumStatus(forumID, 1); err != nil {
		t.Fatalf("ChangeForumStatus: %v", err)
	}

	if err := m.Edit("new title", "new body", "go, api", 7, u1, "/new.png", forumID); err != nil {
		t.Fatalf("Edit: %v", err)
	}

	got, err := m.Get(forumID, u1, true)
	if err != nil {
		t.Fatalf("Get after Edit: %v", err)
	}
	if got.Title != "new title" || got.Content != "new body" {
		t.Fatalf("unexpected edited forum: %#v", got)
	}

	liker := seedUser(t, db, "ian")
	commenter := seedUser(t, db, "jane")
	_, _ = db.Exec(`INSERT INTO forum_likes(forum_id, user_id, like_status) VALUES(?, ?, 1)`, forumID, liker)
	commentID := seedComment(t, db, forumID, commenter, "comment")
	_, _ = db.Exec(`INSERT INTO forum_likes(comment_id, user_id, like_status) VALUES(?, ?, 1)`, commentID, liker)
	_, _ = db.Exec(`INSERT INTO forum_notifications(user_name, body, status, forum_link, user_id, user_not_id) VALUES('u','b','moder', ?, 1, 1)`, forumID)

	if err := m.Remove(forumID); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM forums WHERE id = ?`, forumID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected forum row deleted, count=%d", count)
	}
}

func TestForumListingAndFilters(t *testing.T) {
	db := newTestDB(t)
	m := &ForumModel{DB: db}

	u1 := seedUser(t, db, "kate")
	u2 := seedUser(t, db, "liam")

	fGo := seedForum(t, db, u1, "go post", 1, "go, web")
	fDB := seedForum(t, db, u1, "db post", 1, "db, sql")
	_ = seedForum(t, db, u1, "hidden", 0, "go")
	_, _ = db.Exec(
		`INSERT INTO forums(title, content, tags, user_id, created, expires, image_path, status)
		 VALUES('expired', 'x', 'go', ?, datetime('now'), datetime('now', '-1 day'), '', 1)`,
		u1,
	)

	latest, err := m.Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if len(latest) != 2 {
		t.Fatalf("expected 2 latest visible+unexpired forums, got %d", len(latest))
	}

	all, err := m.ShowAll()
	if err != nil {
		t.Fatalf("ShowAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 showAll forums, got %d", len(all))
	}

	filtered, err := m.ShowCategory([]string{"go"})
	if err != nil {
		t.Fatalf("ShowCategory: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != fGo {
		t.Fatalf("expected only go post in filter result, got %+v", filtered)
	}

	userPosts, err := m.ShowAllUserPosts(u1)
	if err != nil {
		t.Fatalf("ShowAllUserPosts: %v", err)
	}
	if len(userPosts) != 3 { // status filter only, so expired forum is still included.
		t.Fatalf("expected 3 user posts, got %d", len(userPosts))
	}

	ids := make([]int, 0, len(userPosts))
	for _, f := range userPosts {
		ids = append(ids, f.ID)
	}
	sort.Ints(ids)
	if !containsInt(ids, fGo) || !containsInt(ids, fDB) {
		t.Fatalf("unexpected user post ids: %v", ids)
	}

	commentID := seedComment(t, db, fDB, u2, "nice")
	_, _ = db.Exec(`INSERT INTO forum_likes(forum_id, user_id, like_status) VALUES(?, ?, 1)`, fGo, u2)
	_, _ = db.Exec(`INSERT INTO forum_likes(comment_id, user_id, like_status) VALUES(?, ?, 1)`, commentID, u2)

	likedForums, err := m.ShowAllUserLikes(u2)
	if err != nil {
		t.Fatalf("ShowAllUserLikes: %v", err)
	}
	if len(likedForums) != 2 {
		t.Fatalf("expected 2 liked forums (direct + comment-liked), got %d", len(likedForums))
	}
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

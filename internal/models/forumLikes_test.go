package models

import (
	"testing"
)

func TestLikeOrDislikeTransitions(t *testing.T) {
	db := newTestDB(t)
	model := &ForumLikesModel{DB: db}

	u1 := seedUser(t, db, "alice")
	forumID := seedForum(t, db, u1, "post", 1, "go, web")

	gotID, err := model.LikeOrDislike(forumID, u1, 1)
	if err != nil {
		t.Fatalf("insert like: %v", err)
	}
	if gotID != forumID {
		t.Fatalf("returned forum id = %d, want %d", gotID, forumID)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM forum_likes WHERE forum_id = ? AND user_id = ?`, forumID, u1).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 like row after insert, got %d", count)
	}

	if _, err := model.LikeOrDislike(forumID, u1, 1); err != nil {
		t.Fatalf("toggle same like should delete row: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM forum_likes WHERE forum_id = ? AND user_id = ?`, forumID, u1).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after toggling same reaction, got %d", count)
	}

	if _, err := model.LikeOrDislike(forumID, u1, 1); err != nil {
		t.Fatalf("reinsert like: %v", err)
	}
	if _, err := model.LikeOrDislike(forumID, u1, -1); err != nil {
		t.Fatalf("switch to dislike: %v", err)
	}
	var status int
	if err := db.QueryRow(`SELECT like_status FROM forum_likes WHERE forum_id = ? AND user_id = ?`, forumID, u1).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != -1 {
		t.Fatalf("expected switched like_status -1, got %d", status)
	}
}

func TestLikeOrDislikeRejectsInvalidStatus(t *testing.T) {
	db := newTestDB(t)
	model := &ForumLikesModel{DB: db}

	u1 := seedUser(t, db, "bob")
	forumID := seedForum(t, db, u1, "post", 1, "go")

	if _, err := model.LikeOrDislike(forumID, u1, 5); err == nil {
		t.Fatal("expected error for invalid like status")
	}
}

func TestLikeOrDislikeCommentTransitions(t *testing.T) {
	db := newTestDB(t)
	model := &ForumLikesModel{DB: db}

	u1 := seedUser(t, db, "chris")
	u2 := seedUser(t, db, "dora")
	forumID := seedForum(t, db, u1, "post", 1, "go")
	commentID := seedComment(t, db, forumID, u2, "nice post")

	gotForumID, err := model.LikeOrDislikeComment(commentID, u1, 1)
	if err != nil {
		t.Fatalf("insert comment like: %v", err)
	}
	if gotForumID != forumID {
		t.Fatalf("returned forumID = %d, want %d", gotForumID, forumID)
	}

	var status int
	if err := db.QueryRow(`SELECT like_status FROM forum_likes WHERE comment_id = ? AND user_id = ?`, commentID, u1).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != 1 {
		t.Fatalf("expected like_status 1, got %d", status)
	}

	if _, err := model.LikeOrDislikeComment(commentID, u1, -1); err != nil {
		t.Fatalf("switch comment reaction: %v", err)
	}
	if err := db.QueryRow(`SELECT like_status FROM forum_likes WHERE comment_id = ? AND user_id = ?`, commentID, u1).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != -1 {
		t.Fatalf("expected like_status -1 after switch, got %d", status)
	}
}

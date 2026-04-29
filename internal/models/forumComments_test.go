package models

import "testing"

func TestForumCommentCRUD(t *testing.T) {
	db := newTestDB(t)
	commentModel := &ForumCommentModel{DB: db}
	forumModel := &ForumModel{DB: db}

	u1 := seedUser(t, db, "eve")
	forumID := seedForum(t, db, u1, "forum", 1, "go")

	gotForumID, err := commentModel.CommentPost(forumID, u1, "first")
	if err != nil {
		t.Fatalf("CommentPost: %v", err)
	}
	if gotForumID != forumID {
		t.Fatalf("CommentPost returned forumID %d, want %d", gotForumID, forumID)
	}

	var commentID int
	if err := db.QueryRow(`SELECT id FROM forum_comments WHERE forum_id = ? AND user_id = ?`, forumID, u1).Scan(&commentID); err != nil {
		t.Fatalf("query inserted comment: %v", err)
	}

	if err := commentModel.EditCommentPost(forumID, u1, "edited", commentID); err != nil {
		t.Fatalf("EditCommentPost: %v", err)
	}

	var body string
	if err := db.QueryRow(`SELECT comment FROM forum_comments WHERE id = ?`, commentID).Scan(&body); err != nil {
		t.Fatalf("query edited comment: %v", err)
	}
	if body != "edited" {
		t.Fatalf("edited comment = %q, want %q", body, "edited")
	}

	userIDFromComment, err := forumModel.GetUserIDFromComment(commentID)
	if err != nil {
		t.Fatalf("GetUserIDFromComment: %v", err)
	}
	if userIDFromComment != u1 {
		t.Fatalf("GetUserIDFromComment = %d, want %d", userIDFromComment, u1)
	}

	if err := commentModel.RemoveCommentPost(commentID); err != nil {
		t.Fatalf("RemoveCommentPost: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM forum_comments WHERE id = ?`, commentID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected comment to be removed, count=%d", count)
	}
}

func TestShowAllUserComments(t *testing.T) {
	db := newTestDB(t)
	commentModel := &ForumCommentModel{DB: db}
	forumModel := &ForumModel{DB: db}

	u1 := seedUser(t, db, "frank")
	u2 := seedUser(t, db, "gina")
	f1 := seedForum(t, db, u1, "one", 1, "go")
	f2 := seedForum(t, db, u1, "two", 1, "web")

	if _, err := commentModel.CommentPost(f1, u2, "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := commentModel.CommentPost(f2, u2, "c2"); err != nil {
		t.Fatal(err)
	}
	if _, err := commentModel.CommentPost(f1, u1, "owner comment"); err != nil {
		t.Fatal(err)
	}

	comments, err := forumModel.ShowAllUserComments(u2)
	if err != nil {
		t.Fatalf("ShowAllUserComments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments for user %d, got %d", u2, len(comments))
	}
}

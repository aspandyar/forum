package models

import "testing"

func TestForumGetEditAndLookupHelpers(t *testing.T) {
	db := newTestDB(t)
	m := &ForumModel{DB: db}
	cm := &ForumCommentModel{DB: db}

	u1 := seedUser(t, db, "owner")
	u2 := seedUser(t, db, "reader")
	forumID := seedForum(t, db, u1, "title", 1, "go, web")
	commentID := seedComment(t, db, forumID, u2, "hello")

	got, err := m.GetEdit(forumID, u1, true, commentID)
	if err != nil {
		t.Fatalf("GetEdit: %v", err)
	}
	if got.ID != forumID {
		t.Fatalf("GetEdit forum id = %d, want %d", got.ID, forumID)
	}
	if got.EditComment.CommentID != commentID {
		t.Fatalf("expected edit comment id %d, got %d", commentID, got.EditComment.CommentID)
	}

	gotUserID, err := m.GetUserIDFromForum(forumID)
	if err != nil {
		t.Fatalf("GetUserIDFromForum: %v", err)
	}
	if gotUserID != u1 {
		t.Fatalf("GetUserIDFromForum = %d, want %d", gotUserID, u1)
	}

	name, err := m.GetUserByUserID(u1)
	if err != nil {
		t.Fatalf("GetUserByUserID: %v", err)
	}
	if name == "" {
		t.Fatal("expected non-empty user name")
	}

	userByForum, err := m.GetUserByForumID(forumID)
	if err != nil {
		t.Fatalf("GetUserByForumID: %v", err)
	}
	if userByForum.ID != u1 {
		t.Fatalf("GetUserByForumID ID = %d, want %d", userByForum.ID, u1)
	}

	if err := cm.CommentPostNotification(forumID, u2, "commented", "reader"); err == nil {
		t.Fatal("expected CommentPostNotification to fail with current schema")
	}

	if err := m.ChangeForumStatus(forumID, 0); err != nil {
		t.Fatalf("ChangeForumStatus hidden: %v", err)
	}
	if _, err := m.GetUserIDFromForum(forumID); err == nil {
		t.Fatal("expected error for hidden forum in GetUserIDFromForum")
	}
}

func TestGetAllTagsAndCommentUserLookup(t *testing.T) {
	db := newTestDB(t)
	m := &ForumModel{DB: db}

	u := seedUser(t, db, "lookup")
	if _, err := db.Exec(`INSERT INTO forum_tags(tags) VALUES ('go'), ('sql')`); err != nil {
		t.Fatalf("seed tags: %v", err)
	}
	tags, err := m.GetAllTags()
	if err != nil {
		t.Fatalf("GetAllTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	name, err := m.GetUserByUserIDInComment(u)
	if err != nil {
		t.Fatalf("GetUserByUserIDInComment: %v", err)
	}
	if name == "" {
		t.Fatal("expected username from comment lookup helper")
	}
}

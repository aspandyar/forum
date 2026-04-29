package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aspandyar/forum/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func newRepoTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "repo_test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "init-up.sql"))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	for _, stmt := range strings.Split(string(schemaBytes), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec schema: %v", err)
		}
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedRepoUser(t *testing.T, db *sql.DB, email string) int {
	t.Helper()
	userModel := &models.UserModel{DB: db}
	if err := userModel.Insert(strings.Split(email, "@")[0], email, "password123", 2); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var id int
	if err := db.QueryRow(`SELECT id FROM users WHERE email=?`, email).Scan(&id); err != nil {
		t.Fatalf("query user id: %v", err)
	}
	return id
}

func TestAuthRepositories(t *testing.T) {
	db := newRepoTestDB(t)
	userRepo := &UserRepository{Model: &models.UserModel{DB: db}}
	sessionRepo := &SessionRepository{Model: &models.SessionModel{DB: db}}

	if err := userRepo.Insert("alice", "alice@example.com", "password123", 2); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	uid, err := userRepo.Authenticate("alice@example.com", "password123")
	if err != nil || uid <= 0 {
		t.Fatalf("Authenticate: uid=%d err=%v", uid, err)
	}
	session, err := sessionRepo.CreateSession(uid)
	if err != nil || session == nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := sessionRepo.InvalidateSession(session.Token); err != nil {
		t.Fatalf("InvalidateSession: %v", err)
	}
}

func TestForumAndLikeAndCommentRepositories(t *testing.T) {
	db := newRepoTestDB(t)
	forumRepo := &ForumRepository{Model: &models.ForumModel{DB: db}}
	likeRepo := &LikeRepository{Model: &models.ForumLikesModel{DB: db}}
	commentRepo := &CommentRepository{
		CommentModel: &models.ForumCommentModel{DB: db},
		ForumModel:   &models.ForumModel{DB: db},
	}

	uid := seedRepoUser(t, db, "bob@example.com")
	forumID, err := forumRepo.Insert("title", "body", "go, web", 7, uid, "")
	if err != nil {
		t.Fatalf("forum insert: %v", err)
	}
	if err := forumRepo.Edit("title2", "body2", "go", 7, uid, "", forumID); err != nil {
		t.Fatalf("forum edit: %v", err)
	}
	if _, err := forumRepo.Latest(); err != nil {
		t.Fatalf("latest: %v", err)
	}
	if _, err := forumRepo.ShowAll(); err != nil {
		t.Fatalf("show all: %v", err)
	}
	if _, err := forumRepo.ShowCategory([]string{"go"}); err != nil {
		t.Fatalf("show category: %v", err)
	}
	if _, err := forumRepo.ShowAllUserPosts(uid); err != nil {
		t.Fatalf("show user posts: %v", err)
	}
	if _, err := forumRepo.ShowAllUserLikes(uid); err != nil {
		t.Fatalf("show user likes: %v", err)
	}

	if _, err := likeRepo.LikeOrDislike(forumID, uid, 1); err != nil {
		t.Fatalf("like forum: %v", err)
	}
	commentForumID, err := commentRepo.CommentPost(forumID, uid, "hello")
	if err != nil || commentForumID != forumID {
		t.Fatalf("comment post: id=%d err=%v", commentForumID, err)
	}

	var commentID int
	if err := db.QueryRow(`SELECT id FROM forum_comments WHERE forum_id=? AND user_id=?`, forumID, uid).Scan(&commentID); err != nil {
		t.Fatalf("query comment id: %v", err)
	}
	if err := commentRepo.EditCommentPost(forumID, uid, "edited", commentID); err != nil {
		t.Fatalf("edit comment: %v", err)
	}
	if _, err := likeRepo.LikeOrDislikeComment(commentID, uid, 1); err != nil {
		t.Fatalf("like comment: %v", err)
	}
	if _, err := commentRepo.ShowAllUserComments(uid); err != nil {
		t.Fatalf("show user comments: %v", err)
	}
	if err := commentRepo.RemoveCommentPost(commentID); err != nil {
		t.Fatalf("remove comment: %v", err)
	}
	if err := forumRepo.Remove(forumID); err != nil {
		t.Fatalf("forum remove: %v", err)
	}
}

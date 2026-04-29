package models

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "forum_test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "init-up.sql"))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	statements := strings.Split(string(schemaBytes), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec schema statement %q: %v", stmt, err)
		}
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func seedUser(t *testing.T, db *sql.DB, name string) int {
	t.Helper()

	email := name + "@example.com"
	hashed := strings.Repeat("a", 60)
	res, err := db.Exec(
		`INSERT INTO users(name, email, hashed_password, created) VALUES(?, ?, ?, datetime('now'))`,
		name,
		email,
		hashed,
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed user id: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO roles(role, user_id) VALUES(?, ?)`, 2, id); err != nil {
		t.Fatalf("seed role: %v", err)
	}

	return int(id)
}

func seedForum(t *testing.T, db *sql.DB, userID int, title string, status int, tags string) int {
	t.Helper()

	if tags == "" {
		tags = "go, web"
	}

	res, err := db.Exec(
		`INSERT INTO forums(title, content, tags, user_id, created, expires, image_path, status)
		 VALUES(?, 'content', ?, ?, datetime('now'), datetime('now', '+5 day'), '', ?)`,
		title,
		tags,
		userID,
		status,
	)
	if err != nil {
		t.Fatalf("seed forum: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed forum id: %v", err)
	}
	return int(id)
}

func seedComment(t *testing.T, db *sql.DB, forumID, userID int, body string) int {
	t.Helper()

	res, err := db.Exec(
		`INSERT INTO forum_comments(forum_id, user_id, comment) VALUES(?, ?, ?)`,
		forumID,
		userID,
		body,
	)
	if err != nil {
		t.Fatalf("seed comment: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed comment id: %v", err)
	}
	return int(id)
}

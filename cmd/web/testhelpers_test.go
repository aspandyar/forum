package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/aspandyar/forum/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func newWebTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "web_test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schemaBytes, err := os.ReadFile(filepath.Join("..", "..", "init-up.sql"))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	for _, stmt := range strings.Split(string(schemaBytes), ";") {
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

func newWebTestApp(t *testing.T) (*application, *sql.DB) {
	t.Helper()

	db := newWebTestDB(t)
	app := &application{
		errorLog: log.New(io.Discard, "", 0),
		infoLog:  log.New(io.Discard, "", 0),
		forums:   &models.ForumModel{DB: db},
		users:    &models.UserModel{DB: db},
		sessions: &models.SessionModel{DB: db},
		forumLike: &models.ForumLikesModel{
			DB: db,
		},
		forumComment: &models.ForumCommentModel{
			DB: db,
		},
		tempalteCache: map[string]*template.Template{},
	}
	return app, db
}

func seedWebUser(t *testing.T, app *application, name, email string, role int) int {
	t.Helper()
	if err := app.users.Insert(name, email, "password123", role); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var id int
	if err := app.users.DB.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&id); err != nil {
		t.Fatalf("query user id: %v", err)
	}
	return id
}

func attachSessionCookie(t *testing.T, app *application, req *http.Request, userID int) {
	t.Helper()
	session, err := app.sessions.CreateSession(userID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: session.Token, Expires: session.Expiry})
}

func newRequest(method, target string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, body)
	rr := httptest.NewRecorder()
	return req, rr
}

func mustTimeWithin(t *testing.T, got, min, max time.Time) {
	t.Helper()
	if got.Before(min) || got.After(max) {
		t.Fatalf("time %v outside expected range [%v, %v]", got, min, max)
	}
}

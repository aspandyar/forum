package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestRoutesBasicEndpoints(t *testing.T) {
	app, _ := newWebTestApp(t)
	h := app.routes()

	req, rr := newRequest(http.MethodGet, "/swagger", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("/swagger status=%d, want 302", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/swagger", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("/swagger POST status=%d, want 405", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/user/logout", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("protected route unauthenticated status=%d, want 303", rr.Code)
	}
}

func TestOpenDBAndReadDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "db.sqlite")
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	sqlPath := filepath.Join(t.TempDir(), "init.sql")
	sqlText := "CREATE TABLE x(id INTEGER); INSERT INTO x(id) VALUES (1);"
	if err := os.WriteFile(sqlPath, []byte(sqlText), 0600); err != nil {
		t.Fatalf("write sql file: %v", err)
	}

	readDB(sqlPath, db)
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM x`).Scan(&count); err != nil {
		t.Fatalf("query readDB-created table: %v", err)
	}
	if count != 1 {
		t.Fatalf("count=%d, want 1", count)
	}
}

func TestNewTemplateCache(t *testing.T) {
	cache, err := newTemplateCache()
	if err != nil {
		t.Fatalf("newTemplateCache: %v", err)
	}
	if cache == nil {
		t.Fatal("expected non-nil template cache map")
	}
}

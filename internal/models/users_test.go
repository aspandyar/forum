package models

import (
	"errors"
	"testing"
)

func TestUserInsertAuthenticateAndRole(t *testing.T) {
	db := newTestDB(t)
	model := &UserModel{DB: db}

	if err := model.Insert("neo", "neo@example.com", "secret123", 3); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	userID, err := model.Authenticate("neo@example.com", "secret123")
	if err != nil {
		t.Fatalf("Authenticate valid: %v", err)
	}
	if userID <= 0 {
		t.Fatalf("expected valid user ID, got %d", userID)
	}

	role, err := model.GetUserRole(userID)
	if err != nil {
		t.Fatalf("GetUserRole: %v", err)
	}
	if role != 3 {
		t.Fatalf("role = %d, want 3", role)
	}
}

func TestUserInsertDuplicateEmail(t *testing.T) {
	db := newTestDB(t)
	model := &UserModel{DB: db}

	if err := model.Insert("dup1", "dup@example.com", "secret123", 2); err != nil {
		t.Fatalf("initial Insert: %v", err)
	}
	err := model.Insert("dup2", "dup@example.com", "secret456", 2)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestUserAuthenticateInvalidCredentials(t *testing.T) {
	db := newTestDB(t)
	model := &UserModel{DB: db}

	if err := model.Insert("auth", "auth@example.com", "secret123", 2); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if _, err := model.Authenticate("missing@example.com", "secret123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("missing email should return ErrInvalidCredentials, got %v", err)
	}
	if _, err := model.Authenticate("auth@example.com", "wrongpass"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password should return ErrInvalidCredentials, got %v", err)
	}
}

func TestUserTagManagementAndAdminRoleInsert(t *testing.T) {
	db := newTestDB(t)
	model := &UserModel{DB: db}

	if err := model.InsertTags("go"); err != nil {
		t.Fatalf("InsertTags: %v", err)
	}
	if err := model.RemoveTag("go"); err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO users(name, email, hashed_password, created) VALUES('adminseed','adminseed@example.com', ?, datetime('now'))`, "hash"); err != nil {
		t.Fatalf("seed admin user row: %v", err)
	}
	if err := model.GetAdminUser(); err != nil {
		t.Fatalf("GetAdminUser: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM roles WHERE role = 4 AND user_id = 1`).Scan(&count); err != nil {
		t.Fatalf("query admin role row: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected admin role row, got count=%d", count)
	}
}

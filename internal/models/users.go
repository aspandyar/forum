package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string, role int) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created) 
	VALUES (?, ?, ?, strftime('%Y-%m-%d %H:%M:%S', 'now'));`

	result, err := m.DB.Exec(stmt, name, email, hashedPassword)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	user_id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	stmt = `INSERT INTO roles (role, user_id)
	VALUES (?, ?);`
	_, err = m.DB.Exec(stmt, role, user_id)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) GetUserRole(user_id int) (int, error) {
	var role int
	stmt := `SELECT role FROM roles
	WHERE user_id = ?`

	row := m.DB.QueryRow(stmt, user_id)

	err := row.Scan(&role)
	if err != nil {
		return -1, err
	}

	return role, nil
}

func (m *UserModel) GetAdminUser() error {
	stmt := `INSERT INTO roles (role, user_id) VALUES (4, 1);`

	_, err := m.DB.Exec(stmt)
	if err != nil {
		return err
	}

	return nil
}

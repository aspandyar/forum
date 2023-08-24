package models

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Token  int64     // Assuming you're using an automatically incrementing INTEGER primary key
	Data   []byte    // BLOB data
	Expiry time.Time // Expiry date and time
}

type SessionModel struct {
	DB *sql.DB
}

func generateSessionID() string {
	return uuid.New().String()
}

func (m *SessionModel) CreateSession(userID int) (string, error) {
	sessionToken := generateSessionID()
	expiration := time.Now().Add(15 * time.Minute)
	_, err := m.DB.Exec("INSERT INTO sessions (token, user_id, expiry) VALUES (?, ?, ?)", sessionToken, userID, expiration)
	if err != nil {
		return "", err
	}
	return sessionToken, nil
}

func (m *SessionModel) GetSession(token string) (int, time.Time, error) {
	var userID int
	var expiry string
	err := m.DB.QueryRow("SELECT user_id, expiry FROM sessions WHERE token = ?", token).Scan(&userID, &expiry)
	if err != nil {
		return 0, time.Time{}, err
	}
	parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999-07:00", expiry)
	return userID, parsedTime, nil
}

func SetSessionCookie(w http.ResponseWriter, token string, expiration time.Time) {
	sessionCookie := &http.Cookie{
		Name:    "session",
		Value:   token,
		Expires: expiration,
		Path:    "/",
	}
	http.SetCookie(w, sessionCookie)
}

func ClearSessionCookie(w http.ResponseWriter) {
	expiredCookie := &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, expiredCookie)
}

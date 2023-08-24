package models

import (
	"database/sql"
	"net/http"
	"net/url"
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

const (
	sessionCookieName = "session_id"
	sessionDuration   = 15 * time.Minute
)

func generateSessionID() string {
	return uuid.New().String()
}

func (m *SessionModel) CreateSessionForUser(userID int) (string, error) {
	stmt := `INSERT INTO sessions (token, user_id, expiry) VALUES (?, ?, ?);`
	sessionID := generateSessionID()
	expiration := time.Now().Add(sessionDuration)

	_, err := m.DB.Exec(stmt, sessionID, userID, expiration.Format(time.RFC3339))
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (m *SessionModel) GetUserIDFromSessionToken(token string) (int, error) {
	var userID int
	var expiryStr string
	err := m.DB.QueryRow("SELECT user_id, expiry FROM sessions WHERE token = ?", token).Scan(&userID, &expiryStr)
	if err != nil {
		return 0, err
	}

	expiry, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil || time.Now().After(expiry) {
		return 0, err
	}

	return userID, err
}

func (m *SessionModel) DeleteSessionForUser(userID int) error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:    sessionCookieName,
		Value:   sessionID,
		Expires: time.Now().Add(sessionDuration),
	}
	http.SetCookie(w, &cookie)
}

func GetSessionCookie(r *http.Request) (*http.Cookie, error) {
	return r.Cookie(sessionCookieName)
}

func SetFlashMessage(w http.ResponseWriter, message string) {
	cookie := http.Cookie{
		Name:  "flash",
		Value: url.QueryEscape(message),
		// You can set other cookie properties like Path, Expires, etc.
	}
	http.SetCookie(w, &cookie)
}

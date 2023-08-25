package models

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Token  string
	UserID int
	Expiry time.Time
}

type SessionModel struct {
	DB *sql.DB
}

func generateSessionID() string {
	return uuid.New().String()
}

func (m *SessionModel) CreateSession(userID int) (*Session, error) {
	sessionToken := generateSessionID()
	expiration := time.Now().Add(15 * time.Minute)

	existingSession, err := m.GetActiveSession(userID)
	if err != nil {
		return nil, err
	}

	if existingSession != nil {
		err = m.InvalidateSession(existingSession.Token)
		if err != nil {
			return nil, err
		}
	}

	_, err = m.DB.Exec("INSERT INTO sessions (token, user_id, expiry) VALUES (?, ?, ?)", sessionToken, userID, expiration)
	if err != nil {
		return nil, err
	}

	session := &Session{
		Token:  sessionToken,
		UserID: userID,
		Expiry: expiration,
	}

	return session, nil
}

func (m *SessionModel) GetActiveSession(userID int) (*Session, error) {
	var token string
	var expiry string

	err := m.DB.QueryRow("SELECT token, expiry FROM sessions WHERE user_id = ? AND expiry > ?", userID, time.Now()).Scan(&token, &expiry)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active session found
		}
		return nil, err
	}

	parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999-07:00", expiry)
	return &Session{Token: token, UserID: userID, Expiry: parsedTime}, nil
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

func (m *SessionModel) InvalidateSession(token string) error {
	_, err := m.DB.Exec("UPDATE sessions SET expiry = ? WHERE token = ?", time.Now(), token)
	if err != nil {
		return err
	}
	return nil
}

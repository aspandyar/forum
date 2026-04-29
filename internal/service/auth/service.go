package auth

import "github.com/aspandyar/forum/internal/models"

type UserRepository interface {
	Insert(name, email, password string, role int) error
	Authenticate(email, password string) (int, error)
}

type SessionRepository interface {
	CreateSession(userID int) (*models.Session, error)
	InvalidateSession(token string) error
}

type Service struct {
	Users    UserRepository
	Sessions SessionRepository
}

package sqlite

import "github.com/aspandyar/forum/internal/models"

type UserRepository struct {
	Model *models.UserModel
}

func (r *UserRepository) Insert(name, email, password string, role int) error {
	return r.Model.Insert(name, email, password, role)
}

func (r *UserRepository) Authenticate(email, password string) (int, error) {
	return r.Model.Authenticate(email, password)
}

type SessionRepository struct {
	Model *models.SessionModel
}

func (r *SessionRepository) CreateSession(userID int) (*models.Session, error) {
	return r.Model.CreateSession(userID)
}

func (r *SessionRepository) InvalidateSession(token string) error {
	return r.Model.InvalidateSession(token)
}

package forum

import "github.com/aspandyar/forum/internal/models"

type Repository interface {
	Insert(title, content, tags string, expires, userID int, imagePath string) (int, error)
	Edit(title, content, tags string, expires, userID int, imagePath string, forumID int) error
	Remove(forumID int) error
	Latest() ([]*models.Forum, error)
	ShowAll() ([]*models.Forum, error)
	ShowCategory(tags []string) ([]*models.Forum, error)
	ShowAllUserPosts(userID int) ([]*models.Forum, error)
	ShowAllUserLikes(userID int) ([]*models.Forum, error)
}

type Service struct {
	Repo Repository
}

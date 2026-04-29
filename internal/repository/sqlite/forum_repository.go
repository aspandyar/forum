package sqlite

import (
	"github.com/aspandyar/forum/internal/models"
)

type ForumRepository struct {
	Model *models.ForumModel
}

func (r *ForumRepository) Insert(title, content, tags string, expires, userID int, imagePath string) (int, error) {
	return r.Model.Insert(title, content, tags, expires, userID, imagePath)
}

func (r *ForumRepository) Edit(title, content, tags string, expires, userID int, imagePath string, forumID int) error {
	return r.Model.Edit(title, content, tags, expires, userID, imagePath, forumID)
}

func (r *ForumRepository) Remove(forumID int) error {
	return r.Model.Remove(forumID)
}

func (r *ForumRepository) Latest() ([]*models.Forum, error) {
	return r.Model.Latest()
}

func (r *ForumRepository) ShowAll() ([]*models.Forum, error) {
	return r.Model.ShowAll()
}

func (r *ForumRepository) ShowCategory(tags []string) ([]*models.Forum, error) {
	return r.Model.ShowCategory(tags)
}

func (r *ForumRepository) ShowAllUserPosts(userID int) ([]*models.Forum, error) {
	return r.Model.ShowAllUserPosts(userID)
}

func (r *ForumRepository) ShowAllUserLikes(userID int) ([]*models.Forum, error) {
	return r.Model.ShowAllUserLikes(userID)
}

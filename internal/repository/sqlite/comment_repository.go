package sqlite

import "github.com/aspandyar/forum/internal/models"

type CommentRepository struct {
	CommentModel *models.ForumCommentModel
	ForumModel   *models.ForumModel
}

func (r *CommentRepository) CommentPost(forumID, userID int, comment string) (int, error) {
	return r.CommentModel.CommentPost(forumID, userID, comment)
}

func (r *CommentRepository) EditCommentPost(forumID, userID int, comment string, commentID int) error {
	return r.CommentModel.EditCommentPost(forumID, userID, comment, commentID)
}

func (r *CommentRepository) RemoveCommentPost(commentID int) error {
	return r.CommentModel.RemoveCommentPost(commentID)
}

func (r *CommentRepository) ShowAllUserComments(userID int) ([]*models.ForumComment, error) {
	return r.ForumModel.ShowAllUserComments(userID)
}

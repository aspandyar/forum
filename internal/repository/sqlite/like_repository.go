package sqlite

import "github.com/aspandyar/forum/internal/models"

type LikeRepository struct {
	Model *models.ForumLikesModel
}

func (r *LikeRepository) LikeOrDislike(forumID, userID, likeStatus int) (int, error) {
	return r.Model.LikeOrDislike(forumID, userID, likeStatus)
}

func (r *LikeRepository) LikeOrDislikeComment(commentID, userID, likeStatus int) (int, error) {
	return r.Model.LikeOrDislikeComment(commentID, userID, likeStatus)
}

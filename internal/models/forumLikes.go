package models

import (
	"database/sql"
)

type ForumLikes struct {
	ID         int
	ForumID    int
	UserID     int
	LikeStatus int
}

type ForumLikesModel struct {
	DB *sql.DB
}

func (m *ForumLikesModel) LikeOrDislike(forumID, userID int, isLike bool) error {
	var likeStatus int
	if isLike {
		likeStatus = 1
	} else {
		likeStatus = -1
	}

	stmt := `INSERT INTO forum_likes (forum_id, user_id, like_status) VALUES (?, ?, ?);`

	_, err := m.DB.Exec(stmt, forumID, userID, likeStatus)
	if err != nil {
		return err
	}

	return nil
}

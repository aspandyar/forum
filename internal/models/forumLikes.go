package models

import (
	"database/sql"
	"errors"
)

type ForumLikes struct {
	ID         int
	ForumID    int
	UserID     int
	LikeStatus int
}

type ForumLikesComment struct {
	ID         int
	CommentID  int
	UserID     int
	LikeStatus int
}

type ForumLikesModel struct {
	DB *sql.DB
}

func (m *ForumLikesModel) LikeOrDislike(forumID, userID, likeStatus int) (int, error) {
	stmt := `SELECT id, forum_id, user_id, like_status
	FROM forum_likes
	WHERE forum_id = ? AND user_id = ?`

	row := m.DB.QueryRow(stmt, forumID, userID)

	f := &ForumLikes{}

	err := row.Scan(&f.ID, &f.ForumID, &f.UserID, &f.LikeStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if f.ID != 0 {
		if f.LikeStatus == likeStatus {
			stmt = `DELETE FROM forum_likes WHERE forum_id = ? AND user_id = ?`
			_, err = m.DB.Exec(stmt, forumID, userID)
			if err != nil {
				return 0, err
			}
		} else {
			stmt = `UPDATE forum_likes SET like_status = ? WHERE forum_id = ? AND user_id = ?`
			_, err = m.DB.Exec(stmt, likeStatus, forumID, userID)
			if err != nil {
				return 0, err
			}
		}
	} else {

		if likeStatus != 1 && likeStatus != -1 {
			return 0, errors.New("no such data")
		}

		stmt = `INSERT INTO forum_likes (forum_id, user_id, like_status) VALUES (?, ?, ?);`

		_, err = m.DB.Exec(stmt, forumID, userID, likeStatus)
		if err != nil {
			return 0, err
		}
	}

	return forumID, nil
}

func (m *ForumLikesModel) LikeOrDislikeComment(commentID, userID, likeStatus int) (int, error) {
	stmt := `SELECT id, comment_id, user_id, like_status
	FROM forum_likes
	WHERE comment_id = ? AND user_id = ?`

	row := m.DB.QueryRow(stmt, commentID, userID)

	f := &ForumLikesComment{}

	err := row.Scan(&f.ID, &f.CommentID, &f.UserID, &f.LikeStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if f.ID != 0 {
		if f.LikeStatus == likeStatus {
			stmt = `DELETE FROM forum_likes WHERE comment_id = ? AND user_id = ?`
			_, err = m.DB.Exec(stmt, commentID, userID)
			if err != nil {
				return 0, err
			}
		} else {
			stmt = `UPDATE forum_likes SET like_status = ? WHERE comment_id = ? AND user_id = ?`
			_, err = m.DB.Exec(stmt, likeStatus, commentID, userID)
			if err != nil {
				return 0, err
			}
		}
	} else {

		if likeStatus != 1 && likeStatus != -1 {
			return 0, errors.New("no such data")
		}

		stmt = `INSERT INTO forum_likes (comment_id, user_id, like_status) VALUES (?, ?, ?);`

		_, err = m.DB.Exec(stmt, commentID, userID, likeStatus)
		if err != nil {
			return 0, err
		}
	}

	stmt = `SELECT forum_id
	FROM forum_comments
	WHERE id = ?`

	row = m.DB.QueryRow(stmt, commentID)

	var forumID int

	err = row.Scan(&forumID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return forumID, nil
}

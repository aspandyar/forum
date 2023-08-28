package models

import (
	"database/sql"
	"errors"
	"fmt"
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

func (m *ForumLikesModel) LikeOrDislike(forumID, userID, likeStatus int) (int, error) {
	stmt := `SELECT id, forum_id, user_id, like_status
	FROM forum_likes
	WHERE forum_id = ? AND user_id = ?`

	row := m.DB.QueryRow(stmt, forumID, userID)

	f := &ForumLikes{}

	err := row.Scan(&f.ID, &f.ForumID, &f.UserID, &f.LikeStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Println(err)
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

package models

import (
	"database/sql"
)

type ForumComment struct {
	ID      int
	ForumID int
	UserID  int
	Comment string
}

type ForumCommentModel struct {
	DB *sql.DB
}

func (m *ForumCommentModel) CommentPost(forumID, userID int, comment string) (int, error) {
	stmt := `INSERT INTO forum_comments (forum_id, user_id, comment) 
	VALUES (?, ?, ?);`

	_, err := m.DB.Exec(stmt, forumID, userID, comment)
	if err != nil {
		return 0, err
	}

	return forumID, nil
}

func (m *ForumCommentModel) EditCommentPost(forumID, userID int, comment string, commentID int) error {
	stmt := `UPDATE forum_comments 
	SET forum_id = ?, user_id = ?, comment = ?
	WHERE id = ? `

	_, err := m.DB.Exec(stmt, forumID, userID, comment, commentID)
	if err != nil {
		return err
	}
	return nil
}

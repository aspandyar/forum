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

func (m *ForumCommentModel) RemoveCommentPost(commentID int) error {
	stmt := `DELETE FROM forum_comments
	WHERE id = ?`

	_, err := m.DB.Exec(stmt, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (m *ForumModel) GetUserIDFromComment(forumCommentID int) (int, error) {
	stmt := `SELECT user_id
	FROM forum_comments 
	WHERE id = ?;`

	row := m.DB.QueryRow(stmt, forumCommentID)

	var userID int

	err := row.Scan(&userID)
	if err != nil || userID <= 0 {
		return 0, err
	}

	return userID, nil
}

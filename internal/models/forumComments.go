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

func (m *ForumCommentModel) CommentPostNotification(forumID, userID int, comment, user_name string) error {
	stmt := `INSERT INTO forum_notifications (user_name, body, status, forum_link, user_id) 
	VALUES (?, ?, ?, ?, ?);`

	_, err := m.DB.Exec(stmt, user_name, comment, "commented", forumID, userID)
	if err != nil {
		return err
	}

	return nil
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

// func (m *ForumCommentModel) EditCommentPostNotification(body string, notID int) error {
// 	stmt := `UPDATE forum_notifications
// 	SET body = ?
// 	WHERE id = ?`

// 	_, err := m.DB.Exec(stmt, body, notID)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// } TODO: edit comment should eddit also notification... but how to get id from notification???

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

func (m *ForumModel) ShowAllUserComments(userID int) ([]*ForumComment, error) {
	stmt := `SELECT forum_id, comment FROM forum_comments
	WHERE user_id = ?`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	forums := []*ForumComment{}

	for rows.Next() {
		f := &ForumComment{}

		err := rows.Scan(&f.ForumID, &f.Comment)
		if err != nil {
			return nil, err
		}

		forums = append(forums, f)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return forums, nil
}

func (m *ForumModel) GetUserByUserIDInComment(userID int) (string, error) {
	stmt := `SELECT name, email FROM users WHERE id = ?`

	row := m.DB.QueryRow(stmt, userID)

	var user User

	err := row.Scan(&user.Name, &user.Email)
	if err != nil {
		return user.Name, err
	}

	return user.Name, nil
}

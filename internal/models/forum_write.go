package models

import (
	"database/sql"
	"errors"
)

func (m *ForumModel) Insert(title, content, tags string, expires, userID int, imagePath string) (int, error) {
	stmt := `INSERT INTO forums (title, content, tags, user_id, created, expires, image_path) 
	VALUES (?, ?, ?, ?, strftime('%Y-%m-%d %H:%M:%S', 'now'), strftime('%Y-%m-%d %H:%M:%S', 'now', '+' || ? || ' day'), ?);`

	result, err := m.DB.Exec(stmt, title, content, tags, userID, expires, imagePath)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *ForumModel) Edit(title, content, tags string, expires, userID int, imagePath string, forumID int) error {
	stmt := `UPDATE forums 
	SET title = ?, content = ?, tags = ?, user_id = ?, expires = strftime('%Y-%m-%d %H:%M:%S', 'now', '+' || ? || ' day'), image_path = ?
	WHERE id = ?`

	_, err := m.DB.Exec(stmt, title, content, tags, userID, expires, imagePath, forumID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		}
		return err
	}

	return nil
}

func (m *ForumModel) Remove(forumID int) error {
	stmt := `DELETE FROM forums WHERE id = ?;`

	_, err := m.DB.Exec(stmt, forumID)
	if err != nil {
		return err
	}

	stmt = `DELETE FROM forum_likes WHERE forum_id = ?;`

	_, err = m.DB.Exec(stmt, forumID)
	if err != nil {
		return err
	}

	stmt = `DELETE FROM forum_comments WHERE forum_id = ?;`

	_, err = m.DB.Exec(stmt, forumID)
	if err != nil {
		return err
	}

	stmt = `DELETE FROM forum_notifications WHERE forum_link = ?;`

	_, err = m.DB.Exec(stmt, forumID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ForumModel) ChangeForumStatus(forumID int, status int) error {
	stmt := `UPDATE forums
	SET status = ? WHERE id = ?`

	_, err := m.DB.Exec(stmt, status, forumID)
	if err != nil {
		return err
	}

	return nil
}

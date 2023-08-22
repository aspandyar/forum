package models

import (
	"database/sql"
	"errors"
	"time"
)

type Forum struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type ForumModel struct {
	DB *sql.DB
}

func (m *ForumModel) Insert(title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO forums (title, content, created, expires) 
	VALUES (?, ?, strftime('%Y-%m-%d %H:%M:%S', 'now'), strftime('%Y-%m-%d %H:%M:%S', 'now', '+' || ? || ' day'));`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *ForumModel) Get(id int) (*Forum, error) {
	stmt := `SELECT id, title, content, created, expires
	FROM forums
	WHERE expires > datetime('now') AND id = ?;`

	row := m.DB.QueryRow(stmt, id)

	f := &Forum{}

	err := row.Scan(&f.ID, &f.Title, &f.Content, &f.Created, &f.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return f, nil
}

func (m *ForumModel) Latest() ([]*Forum, error) {
	stmt := `SELECT id, title, content, created, expires
	FROM forums
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now')
	ORDER BY id DESC
	LIMIT 10;`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	forums := []*Forum{}

	for rows.Next() {
		f := &Forum{}

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Created, &f.Expires)
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

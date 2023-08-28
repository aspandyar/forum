package models

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"time"
)

type Forum struct {
	ID            int
	Title         string
	Content       string
	Reacted       bool
	Liked         bool
	LikesCount    int
	DislikesCount int
	Created       time.Time
	Expires       time.Time
}

type ForumModel struct {
	DB *sql.DB
}

func (m *ForumModel) Init(initSqlFileName string) error {
	file, err := ioutil.ReadFile(initSqlFileName)
	if err != nil {
		log.Fatalf("Can't read SQL file %v", err)
	}
	// Execute all
	_, err = m.DB.Exec(string(file))
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	return nil
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

func (m *ForumModel) Get(id, userId int) (*Forum, error) {
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

	var reacted, liked bool

	stmt = `SELECT id, like_status
	FROM forum_likes
	WHERE forum_id = ? AND user_id = ?`

	fl := &ForumLikes{}

	rowL := m.DB.QueryRow(stmt, id, userId)

	err = rowL.Scan(&fl.ID, &fl.LikeStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			reacted = false
			liked = false
		} else {
			return nil, err
		}
	}

	if fl.ID == 0 {
		reacted = false
		liked = false
	} else if fl.LikeStatus == 1 {
		reacted = true
		liked = true
	} else if fl.LikeStatus == -1 {
		reacted = true
		liked = false
	}

	f.Reacted = reacted
	f.Liked = liked

	stmt = `SELECT f.id, f.title,
		SUM(CASE WHEN l.like_status = 1 THEN 1 ELSE 0 END) AS like_count,
		SUM(CASE WHEN l.like_status = -1 THEN 1 ELSE 0 END) AS dislike_count
	FROM forums f
	LEFT JOIN forum_likes l ON f.id = l.forum_id
	WHERE f.id = ?
	GROUP BY f.id, f.title`

	rows := m.DB.QueryRow(stmt, id)

	fs := &Forum{}

	err = rows.Scan(&fs.ID, &fs.Title, &fs.LikesCount, &fs.DislikesCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	f.LikesCount = fs.LikesCount
	f.DislikesCount = fs.DislikesCount

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

func (m *ForumModel) ShowAll() ([]*Forum, error) {
	stmt := `SELECT id, title, content, created, expires
	FROM forums
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now')
	ORDER BY id DESC;`

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

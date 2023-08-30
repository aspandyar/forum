package models

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Forum struct {
	ID            int
	Title         string
	Content       string
	Tags          string
	TagsOutput    []string
	Reacted       bool
	Liked         bool
	LikesCount    int
	DislikesCount int
	Comment       []userComment
	Created       time.Time
	Expires       time.Time
}

type userComment struct {
	CommentID     int
	User          string
	Comment       string
	Reacted       bool
	Liked         bool
	LikesCount    int
	DislikesCount int
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

func (m *ForumModel) Insert(title, content, tags string, expires, userID int) (int, error) {
	stmt := `INSERT INTO forums (title, content, tags, user_id, created, expires) 
	VALUES (?, ?, ?, ?, strftime('%Y-%m-%d %H:%M:%S', 'now'), strftime('%Y-%m-%d %H:%M:%S', 'now', '+' || ? || ' day'));`

	result, err := m.DB.Exec(stmt, title, content, tags, userID, expires)
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
	stmt := `SELECT id, title, content, tags, created, expires
	FROM forums
	WHERE expires > datetime('now') AND id = ?;`

	row := m.DB.QueryRow(stmt, id)

	f := &Forum{}

	err := row.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	f.TagsOutput = strings.Split(f.Tags, ", ")

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

	stmt = `SELECT f.id,
		SUM(CASE WHEN l.like_status = 1 THEN 1 ELSE 0 END) AS like_count,
		SUM(CASE WHEN l.like_status = -1 THEN 1 ELSE 0 END) AS dislike_count
	FROM forums f
	LEFT JOIN forum_likes l ON f.id = l.forum_id
	WHERE f.id = ?
	GROUP BY f.id`

	rows := m.DB.QueryRow(stmt, id)

	fs := &Forum{}

	err = rows.Scan(&fs.ID, &fs.LikesCount, &fs.DislikesCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	f.LikesCount = fs.LikesCount
	f.DislikesCount = fs.DislikesCount

	stmt = `SELECT u.name, fc.comment, fc.id
    FROM forum_comments fc
    JOIN users u ON fc.user_id = u.id
    WHERE fc.forum_id = ?`

	rowsL, err := m.DB.Query(stmt, id)
	if err != nil {
		return nil, err
	}
	defer rowsL.Close()

	var userComments []userComment

	for rowsL.Next() {
		var userComment userComment

		err := rowsL.Scan(&userComment.User, &userComment.Comment, &userComment.CommentID)
		if err != nil {
			return nil, err
		}

		var reacted, liked bool

		stmt = `SELECT id, like_status
		FROM forum_likes
		WHERE comment_id = ? AND user_id = ?`

		fl := &ForumLikesComment{}

		rowTemp := m.DB.QueryRow(stmt, userComment.CommentID, userId)

		err = rowTemp.Scan(&fl.ID, &fl.LikeStatus)
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

		userComment.Reacted = reacted
		userComment.Liked = liked

		stmt = `SELECT f.id,
		SUM(CASE WHEN l.like_status = 1 THEN 1 ELSE 0 END) AS like_count,
		SUM(CASE WHEN l.like_status = -1 THEN 1 ELSE 0 END) AS dislike_count
		FROM forum_comments f
		LEFT JOIN forum_likes l ON f.id = l.comment_id
		WHERE f.id = ?
		GROUP BY f.id`

		rows := m.DB.QueryRow(stmt, userComment.CommentID)

		fs := &ForumComment{}

		err = rows.Scan(&fs.ID, &userComment.LikesCount, &userComment.DislikesCount)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNoRecord
			} else {
				return nil, err
			}
		}

		userComments = append(userComments, userComment)
	}

	if err := rowsL.Err(); err != nil {
		return nil, err
	}

	f.Comment = userComments

	return f, nil
}

func (m *ForumModel) Latest() ([]*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires
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

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
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
	stmt := `SELECT id, title, content, tags, created, expires
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

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
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

func (m *ForumModel) ShowCategory(tags []string) ([]*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires
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

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
		if err != nil {
			return nil, err
		}

		f.TagsOutput = strings.Split(f.Tags, ", ")

		if containsAny(tags, f.TagsOutput) {
			forums = append(forums, f)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return forums, nil
}

func (m *ForumModel) ShowAllUserPosts(userID int) ([]*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires
	FROM forums 
	WHERE user_id = ?;`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	forums := []*Forum{}

	for rows.Next() {
		f := &Forum{}

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
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

func (m *ForumModel) ShowAllUserLikes(userID int) ([]*Forum, error) {
	stmt := `SELECT f.id, f.title, f.content, f.tags, f.created, f.expires
	FROM forums f
	INNER JOIN (
		SELECT forum_id
		FROM forum_likes
		WHERE user_id = ?
		
		UNION
		
		SELECT fc.forum_id
		FROM forum_comments fc
		JOIN forum_likes fl ON fl.comment_id = fc.id 
		WHERE fl.user_id = ?
	) AS relevant_forums ON f.id = relevant_forums.forum_id
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now');`

	rows, err := m.DB.Query(stmt, userID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	forums := []*Forum{}

	for rows.Next() {
		f := &Forum{}

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
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

func containsAny(source []string, target []string) bool {
	for _, s := range source {
		for _, t := range target {
			if s == t {
				return true
			}
		}
	}
	return false
}

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
	ImagePath     string
	IsOwnForum    bool
	EditComment   userComment
}

type userComment struct {
	CommentID     int
	ForumID       int
	User          string
	Comment       string
	Reacted       bool
	Liked         bool
	LikesCount    int
	DislikesCount int
	IsOwnComment  bool
}

type Notification struct {
	UserID        int
	ForumID       int
	UserCommented string
	Body          string
	Status        int
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
		} else {
			return err
		}
	}

	return nil
}

func (m *ForumModel) Get(id, userId int, isOwnForum bool) (*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires, image_path
	FROM forums
	WHERE expires > datetime('now') AND id = ?;`

	row := m.DB.QueryRow(stmt, id)

	f := &Forum{}

	err := row.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires, &f.ImagePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	f.TagsOutput = strings.Split(f.Tags, ", ")
	f.IsOwnForum = isOwnForum

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

	stmt = `SELECT u.name, fc.comment, fc.id, u.id
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

		var givenUser int
		err := rowsL.Scan(&userComment.User, &userComment.Comment, &userComment.CommentID, &givenUser)
		if err != nil {
			return nil, err
		}

		if givenUser == userId {
			userComment.IsOwnComment = true
			userComment.ForumID = f.ID
		} else {
			userComment.IsOwnComment = false
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

func (m *ForumModel) GetEdit(forumID, userID int, isOwnForum bool, commentID int) (*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires, image_path
	FROM forums
	WHERE expires > datetime('now') AND id = ?;`

	row := m.DB.QueryRow(stmt, forumID)

	f := &Forum{}

	err := row.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires, &f.ImagePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	f.TagsOutput = strings.Split(f.Tags, ", ")
	f.IsOwnForum = isOwnForum

	var reacted, liked bool

	stmt = `SELECT id, like_status
	FROM forum_likes
	WHERE forum_id = ? AND user_id = ?`

	fl := &ForumLikes{}

	rowL := m.DB.QueryRow(stmt, forumID, userID)

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

	rows := m.DB.QueryRow(stmt, forumID)

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

	stmt = `SELECT u.name, fc.comment, fc.id, u.id
    FROM forum_comments fc
    JOIN users u ON fc.user_id = u.id
    WHERE fc.forum_id = ?`

	rowsL, err := m.DB.Query(stmt, forumID)
	if err != nil {
		return nil, err
	}
	defer rowsL.Close()

	var userComments []userComment

	for rowsL.Next() {
		var userComment userComment

		var givenUser int
		err := rowsL.Scan(&userComment.User, &userComment.Comment, &userComment.CommentID, &givenUser)
		if err != nil {
			return nil, err
		}

		if givenUser == userID {
			userComment.IsOwnComment = true
			userComment.ForumID = f.ID
		} else {
			userComment.IsOwnComment = false
		}

		var reacted, liked bool

		stmt = `SELECT id, like_status
		FROM forum_likes
		WHERE comment_id = ? AND user_id = ?`

		fl := &ForumLikesComment{}

		rowTemp := m.DB.QueryRow(stmt, userComment.CommentID, userID)

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

		if userComment.CommentID == commentID {
			f.EditComment = userComment
		} else {
			userComments = append(userComments, userComment)
		}
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

	return nil
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

func (m *ForumModel) ShowUserNotification(userID int) ([]*Notification, error) {
	stmt := `SELECT id, title, content, tags, created, expires
	FROM forums 
	WHERE user_id = ?;`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	AllUserForums := []*Forum{}

	for rows.Next() {
		f := &Forum{}

		err := rows.Scan(&f.ID, &f.Title, &f.Content, &f.Tags, &f.Created, &f.Expires)
		if err != nil {
			return nil, err
		}

		AllUserForums = append(AllUserForums, f)
	}

	allNot := []*Notification{}

	for _, userForums := range AllUserForums {
		stmt = `SELECT comment, user_id
		FROM forum_comments 
		WHERE forum_id = ? AND user_id != ?;`

		rows, err = m.DB.Query(stmt, userForums.ID, userID)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			not := &Notification{}

			var userIDTemp int
			err := rows.Scan(&not.Body, &userIDTemp)
			if err != nil {
				return nil, err
			}

			not.UserCommented, err = m.GetUserByUserID(userIDTemp)
			if err != nil {
				return nil, err
			}

			not.UserID = userID
			not.ForumID = userForums.ID
			not.Status = 0 //commented: 0, liked: 1, disliked: -1

			allNot = append(allNot, not)
		}

		stmt = `SELECT like_status, user_id
		FROM forum_likes
		WHERE forum_id = ? AND user_id != ?`

		rows, err = m.DB.Query(stmt, userForums.ID, userID)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			not := &Notification{}

			var tempUserID int
			err := rows.Scan(&not.Status, &tempUserID)
			if err != nil {
				return nil, err
			}

			not.UserCommented, err = m.GetUserByUserID(tempUserID)
			if err != nil {
				return nil, err
			}

			if not.Status == 1 {
				not.Body = "liked"
			} else if not.Status == -1 {
				not.Body = "disliked"
			} else {
				break
			}

			not.UserID = userID
			not.ForumID = userForums.ID
			//commented: 1, liked: 2, disliked: 3

			allNot = append(allNot, not)
		}

		stmt = `SELECT fl.like_status, fl.user_id
		FROM forum_likes fl
		INNER JOIN forum_comments fc ON fl.comment_id = fc.id
		WHERE fc.user_id != fl.user_id AND fc.user_id = ?;`

		rows, err = m.DB.Query(stmt, userID)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			not := &Notification{}

			var tempUserID int
			err := rows.Scan(&not.Status, &tempUserID)
			if err != nil {
				return nil, err
			}

			not.UserCommented, err = m.GetUserByUserID(tempUserID)
			if err != nil {
				return nil, err
			}

			if not.Status == 1 {
				not.Body = "liked"
			} else if not.Status == -1 {
				not.Body = "disliked"
			} else {
				break
			}

			not.UserID = userID
			not.ForumID = userForums.ID
			//commented: 1, liked: 2, disliked: 3

			allNot = append(allNot, not)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return allNot, nil
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

func (m *ForumModel) GetUserIDFromForum(forumID int) (int, error) {
	stmt := `SELECT user_id
	FROM forums 
	WHERE id = ?;`

	row := m.DB.QueryRow(stmt, forumID)

	var userID int

	err := row.Scan(&userID)
	if err != nil || userID <= 0 {
		return 0, err
	}

	return userID, nil
}

func (m *ForumModel) GetUserByUserID(userID int) (string, error) {
	stmt := `SELECT name, email FROM users WHERE id = ?`

	row := m.DB.QueryRow(stmt, userID)

	var user User

	err := row.Scan(&user.Name, &user.Email)
	if err != nil {
		return user.Name, err
	}

	return user.Name, nil
}

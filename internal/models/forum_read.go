package models

import (
	"database/sql"
	"errors"
	"strings"
)

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
		}
		return nil, err
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
		}
		return nil, err
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

		if givenUser == userId || userId == AdminID {
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
			}
			return nil, err
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
		}
		return nil, err
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
		}
		return nil, err
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
			}
			return nil, err
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

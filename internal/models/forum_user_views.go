package models

func (m *ForumModel) ShowAllUserPosts(userID int) ([]*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires
	FROM forums 
	WHERE user_id = ? AND status = 1;`

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
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now') AND f.status = 1;`

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

func (m *ForumModel) GetUserIDFromForum(forumID int) (int, error) {
	stmt := `SELECT user_id
	FROM forums 
	WHERE id = ? AND status = 1;`

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

func (m *ForumModel) GetUserByForumID(forumID int) (User, error) {
	stmt := `SELECT user_id FROM forums WHERE id = ?`

	row := m.DB.QueryRow(stmt, forumID)
	var user User
	err := row.Scan(&user.ID)
	if err != nil {
		return user, err
	}

	user.Name, err = m.GetUserByUserID(user.ID)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (m *ForumModel) GetAllTags() ([]string, error) {
	stmt := `SELECT tags FROM forum_tags;`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

package models

import "strings"

func (m *ForumModel) Latest() ([]*Forum, error) {
	stmt := `SELECT id, title, content, tags, created, expires
	FROM forums
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now') AND status = 1
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
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now') AND status = 1
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
	WHERE expires > strftime('%Y-%m-%d %H:%M:%S', 'now') AND status = 1
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

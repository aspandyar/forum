package models

type Notification struct {
	ID            int
	UserID        int
	ForumID       int
	UserCommented string
	Body          string
	Status        string
}

func (m *ForumModel) ShowUserNotification(userID int) ([]*Notification, error) {
	stmt := `SELECT id, user_name, body, status, forum_link
	FROM forum_notifications
	WHERE user_id = ?;`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notifications := []*Notification{}

	for rows.Next() {
		notification := &Notification{}

		err := rows.Scan(&notification.ID, &notification.UserCommented, &notification.Body, &notification.Status, &notification.ForumID)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (m *ForumModel) RemoveUserNotification(id int) error {
	stmt := `DELETE FROM forum_notifications WHERE id = ?;`

	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

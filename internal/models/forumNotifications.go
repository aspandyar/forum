package models

type Notification struct {
	UserID        int
	ForumID       int
	UserCommented string
	Body          string
	Status        string
}

func (m *ForumModel) ShowUserNotification(userID int) ([]*Notification, error) { //TODO: write following function, which should output all notification from table, then add several functions for forum_likes and forum_dislikes, remove and edit
	stmt := `SELECT user_name, body, status, forum_link
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

		err := rows.Scan(&notification.UserCommented, &notification.Body, &notification.Status, &notification.ForumID)
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

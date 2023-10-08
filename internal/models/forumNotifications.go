package models

import "errors"

type Notification struct {
	ID              int
	UserID          int
	ForumID         int
	UserCommented   string
	UserCommentedID int
	Body            string
	Status          string
}

const (
	ModeratorRole = 3
	AdminRole     = 4
	AdminStatus   = "admin"
	ModerStatus   = "moder"
)

func (m *ForumModel) AskForModeration(userID int) error {
	var err error

	stmt := `INSERT INTO forum_notifications (user_name, body, status, forum_link, user_id, user_not_id)
	VALUES(?, ?, ?, ?, ?, ?)`

	not := *&Notification{}

	not.Body = "asked for moder"
	not.Status = "admin"
	not.ForumID = 0 // not needed here
	not.UserID = AdminRole
	not.UserCommented, err = m.GetUserByUserID(userID)
	not.UserCommentedID = userID

	_, err = m.DB.Exec(stmt, not.UserCommented, not.Body, not.Status, not.ForumID, not.UserID, not.UserCommentedID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ForumModel) ReportForum(forumID int, body string) error {
	stmt := `INSERT INTO forum_notifications (user_name, body, status, forum_link, user_id, user_not_id)
	VALUES(?, ?, ?, ?, ?, ?)`
	not := *&Notification{}

	var user User
	user, err := m.GetUserByForumID(forumID)
	if err != nil {
		return err
	}

	not.Body = body
	not.Status = "admin"
	not.ForumID = forumID
	not.UserID = AdminRole
	not.UserCommentedID = user.ID
	not.UserCommented, err = m.GetUserByUserID(user.ID)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(stmt, not.UserCommented, not.Body, not.Status, not.ForumID, not.UserID, not.UserCommentedID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ForumModel) AskForNewForum(forumID, userID int, body string) error {
	var err error

	stmt := `INSERT INTO forum_notifications (user_name, body, status, forum_link, user_id, user_not_id)
	VALUES(?, ?, ?, ?, ?, ?)`

	not := *&Notification{}

	not.Body = body
	not.Status = "moder"
	not.ForumID = forumID
	not.UserCommentedID = -1 // not needed  here
	not.UserCommented, err = m.GetUserByUserID(userID)
	not.UserID = userID

	_, err = m.DB.Exec(stmt, not.UserCommented, not.Body, not.Status, not.ForumID, not.UserID, not.UserCommentedID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ForumModel) ShowUserNotification(role int) ([]*Notification, error) {
	stmt := `SELECT id, user_name, body, status, forum_link, user_id, user_not_id
		FROM forum_notifications
		WHERE status = ?;`

	var status string
	if role == AdminRole {
		status = AdminStatus
	} else if role == ModeratorRole {
		status = ModerStatus
	} else {
		return nil, errors.New("how u get here?")
	}

	rows, err := m.DB.Query(stmt, status)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notifications := []*Notification{}

	for rows.Next() {
		notification := &Notification{}

		err := rows.Scan(&notification.ID, &notification.UserCommented, &notification.Body, &notification.Status, &notification.ForumID, &notification.UserID, &notification.UserCommentedID)
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

// func (m *ForumModel) ShowUserNotification(userID int) ([]*Notification, error) {
// 	stmt := `SELECT id, user_name, body, status, forum_link
// 	FROM forum_notifications
// 	WHERE user_id = ?;`

// 	rows, err := m.DB.Query(stmt, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	notifications := []*Notification{}

// 	for rows.Next() {
// 		notification := &Notification{}

// 		err := rows.Scan(&notification.ID, &notification.UserCommented, &notification.Body, &notification.Status, &notification.ForumID)
// 		if err != nil {
// 			return nil, err
// 		}

// 		notifications = append(notifications, notification)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return notifications, nil
// }

func (m *ForumModel) RemoveUserNotification(id int) error {
	stmt := `DELETE FROM forum_notifications WHERE id = ?;`

	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *ForumModel) ChangeUserRole(userID, role int) error {
	stmt := `UPDATE roles
	SET role = ? WHERE user_id = ?`

	_, err := m.DB.Exec(stmt, role, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ForumModel) GetRoleByUserID(userID int) (int, error) {
	stmt := `SELECT role FROM roles WHERE user_id = ?`

	row := m.DB.QueryRow(stmt, userID)

	var role int
	err := row.Scan(&role)
	if err != nil {
		return 0, err
	}

	return role, nil
}

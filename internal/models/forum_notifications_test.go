package models

import "testing"

func TestNotificationFlowsAndRoleQueries(t *testing.T) {
	db := newTestDB(t)
	model := &ForumModel{DB: db}

	adminID := seedUser(t, db, "admin")
	if adminID != 1 {
		t.Fatalf("expected seeded admin user id 1 for these tests, got %d", adminID)
	}
	moderID := seedUser(t, db, "moder")
	userID := seedUser(t, db, "report-target")
	forumID := seedForum(t, db, userID, "pending report", 1, "go")

	if err := model.AskForModeration(moderID); err != nil {
		t.Fatalf("AskForModeration: %v", err)
	}
	if err := model.ReportForum(forumID, moderID, "spam"); err != nil {
		t.Fatalf("ReportForum: %v", err)
	}
	if err := model.AnswerFromAdmin(moderID, "approved"); err != nil {
		t.Fatalf("AnswerFromAdmin: %v", err)
	}
	if err := model.AskForNewForum(forumID, userID, "new forum body"); err != nil {
		t.Fatalf("AskForNewForum: %v", err)
	}

	adminNotifications, err := model.ShowUserNotification(AdminRole)
	if err != nil {
		t.Fatalf("ShowUserNotification admin: %v", err)
	}
	if len(adminNotifications) < 2 {
		t.Fatalf("expected at least 2 admin notifications, got %d", len(adminNotifications))
	}

	moderNotifications, err := model.ShowUserNotification(ModeratorRole)
	if err != nil {
		t.Fatalf("ShowUserNotification moder: %v", err)
	}
	if len(moderNotifications) < 2 {
		t.Fatalf("expected at least 2 moder notifications, got %d", len(moderNotifications))
	}
	if _, err := model.ShowUserNotification(0); err == nil {
		t.Fatal("expected invalid role error")
	}

	if err := model.RemoveUserNotification(adminNotifications[0].ID); err != nil {
		t.Fatalf("RemoveUserNotification: %v", err)
	}
	if err := model.ChangeUserRole(moderID, ModeratorRole); err != nil {
		t.Fatalf("ChangeUserRole: %v", err)
	}
	role, err := model.GetRoleByUserID(moderID)
	if err != nil {
		t.Fatalf("GetRoleByUserID: %v", err)
	}
	if role != ModeratorRole {
		t.Fatalf("role = %d, want %d", role, ModeratorRole)
	}
}

package main

import (
	"net/http"
	"os"
	"time"

	"github.com/aspandyar/forum/internal/validator"
)

type forumCreateForm struct {
	Title     string
	Content   string
	Tags      string
	Expires   int
	ImagePath string
	AllTags   []string
	validator.Validator
}

type userSingupForm struct {
	Name     string
	Email    string
	Password string
	Role     int
	validator.Validator
}

type userLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

type forumLikeForm struct {
	LikeStatus int
	ForumID    int
	UserID     int
	CommentID  int
	validator.Validator
}

type forumCommentForm struct {
	ForumID int
	UserID  int
	Comment string
	validator.Validator
}

type forumReportForm struct {
	reportTypes   string
	reportDetails string
	validator.Validator
}

const (
	adminID = 1

	guestRole     = 1
	userRole      = 2
	moderatorRole = 3
	adminRole     = 4

	visibleStatus   = 1
	invisibleStatus = 0
)

var (
	clientID        = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret    = os.Getenv("GOOGLE_CLIENT_SECRET")
	clientGitID     = os.Getenv("GITHUB_CLIENT_ID")
	clientGitSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI     = func() string {
		if value := os.Getenv("OAUTH_REDIRECT_URI"); value != "" {
			return value
		}
		return "https://localhost:4000/callback"
	}()
)

type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string
}

func (app *application) createAdmin() error {
	adminName := os.Getenv("ADMIN_NAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	app.users.Insert(adminName, adminEmail, adminPassword, adminRole)
	app.users.InsertTags("tag 1")
	app.users.InsertTags("tag 2")
	app.users.InsertTags("tag 3")
	app.users.GetAdminUser()

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	_, expiry, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		return false
	}

	return time.Now().Before(expiry)
}

func (app *application) getRole(r *http.Request) int {
	if !app.isAuthenticated(r) {
		return guestRole
	}

	cookie, _ := r.Cookie("session")
	userID, _, _ := app.sessions.GetSession(cookie.Value)

	role, err := app.users.GetUserRole(userID)
	if err != nil {
		return -1
	}

	return role
}

func (app *application) isOwnForum(userID int, r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	getUserID, expiry, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		return false
	}

	return time.Now().Before(expiry) && (userID == getUserID || getUserID == adminID)
}

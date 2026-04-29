package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aspandyar/forum/internal/models"
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
	// var err error

	adminName := os.Getenv("ADMIN_NAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	app.users.Insert(adminName, adminEmail, adminPassword, adminRole)
	// var sqliteErr sqlite3.Error
	// if errors.Is(err, models.ErrDuplicateEmail) || strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed") {
	// }

	app.users.InsertTags("tag 1")
	app.users.InsertTags("tag 2")
	app.users.InsertTags("tag 3")

	app.users.GetAdminUser()
	// if errors.Is(err, models.ErrDuplicateEmail) || strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed") {
	// 	return nil
	// }

	return nil
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	forums, err := app.forums.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Forums = forums

	app.render(w, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) allForum(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/showAll" {
		app.notFound(w)
		return
	}

	forums, err := app.forums.ShowAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Forums = forums

	app.render(w, http.StatusOK, "allForums.tmpl.html", data)
}

func (app *application) forumAllUserPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/forum/allPosts" {
		app.notFound(w)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	forums, err := app.forums.ShowAllUserPosts(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Forums = forums

	app.render(w, http.StatusOK, "allForums.tmpl.html", data)
}

func (app *application) forumAllUserLikes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/forum/allLikes" {
		app.notFound(w)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	forums, err := app.forums.ShowAllUserLikes(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Forums = forums

	app.render(w, http.StatusOK, "allForums.tmpl.html", data)
}

func (app *application) forumAllUserComments(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/forum/all_comments" {
		app.notFound(w)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	forums, err := app.forums.ShowAllUserComments(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Form = forums

	app.render(w, http.StatusOK, "allComments.tmpl.html", data)
}

func (app *application) userNotification(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	role, err := app.forums.GetRoleByUserID(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if role != adminRole && role != moderatorRole {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	notifications, err := app.forums.ShowUserNotification(role)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)

	data.Form = notifications

	app.render(w, http.StatusOK, "notification.tmpl.html", data)
}

func (app *application) userNotificationRemove(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	adminID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	role, err := app.forums.GetRoleByUserID(adminID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if role != adminRole && role != moderatorRole {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	err = app.forums.RemoveUserNotification(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) forumReportRemoveHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	adminID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	role, err := app.forums.GetRoleByUserID(adminID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if role != adminRole {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[6]
	getUserID, err := strconv.Atoi(idStr)
	if err != nil || getUserID < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[5]
	forumID, err := strconv.Atoi(idStr)
	if err != nil || forumID < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	err = app.forums.ChangeForumStatus(forumID, invisibleStatus)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.AnswerFromAdmin(getUserID, "approved")
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.RemoveUserNotification(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) userModerationDone(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[4]
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	err = app.forums.ChangeUserRole(userID, moderatorRole)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.RemoveUserNotification(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) moderDenoteHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[3]
	moderID, err := strconv.Atoi(idStr)
	if err != nil || moderID < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	err = app.forums.ChangeUserRole(moderID, userRole)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.RemoveUserNotification(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) forumAcceptHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[3]
	notID, err := strconv.Atoi(idStr)
	if err != nil || notID < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[4]
	fourmID, err := strconv.Atoi(idStr)
	if err != nil || fourmID < 1 {
		http.NotFound(w, r)
		return
	}

	err = app.forums.ChangeForumStatus(fourmID, visibleStatus)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.RemoveUserNotification(notID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) addTagsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.addTagsGet(w, r)
	case http.MethodPost:
		app.addTagsPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) addTagsGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "addTags.tmpl.html", data)
}

func (app *application) addTagsPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	add := r.FormValue("add_tag")
	remove := r.FormValue("remove_tag")
	tagsText := r.FormValue("tags_text")

	tags := strings.Split(tagsText, ", ")

	if add != "" && remove != "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if add == "1" {
		for _, tag := range tags {
			err = app.users.InsertTags(tag)
		}
	} else if remove == "-1" {
		for _, tag := range tags {
			err = app.users.RemoveTag(tag)
		}
	} else {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) forumReportHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.ForumReportGet(w, r)
	case http.MethodPost:
		app.ForumReportPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) ForumReportGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "report.tmpl.html", data)
}

func (app *application) ForumReportPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	types := r.PostForm["reportType"]

	form := forumReportForm{
		reportTypes:   strings.Join(types, ", "),
		reportDetails: r.PostForm.Get("reportDetails"),
	}

	form.CheckField(validator.MaxChars(form.reportTypes, 200), "tags", "This field cannot be more than 200 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "report.tmpl.html", data)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[3]
	forumID, err := strconv.Atoi(idStr)
	if err != nil || forumID < 1 {
		http.NotFound(w, r)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	moderID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.ReportForum(forumID, moderID, form.reportTypes+" "+form.reportDetails)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", forumID), http.StatusSeeOther)
}

func (app *application) forumCategory(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/forum/category" {
		app.notFound(w)
		return
	}

	var forum []*models.Forum
	var err error
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		selectedTags := r.Form["tags"]
		customTagsStr := r.PostForm.Get("custom_tags")

		tags := app.processTags(selectedTags, customTagsStr)

		forum, err = app.forums.ShowCategory(tags)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		forum, err = app.forums.ShowAll()
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	data := app.newTemplateData(r)

	data.Forums = forum

	app.render(w, http.StatusOK, "category.tmpl.html", data)
}

func (app *application) forumView(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "view" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	var userID int

	cookie, err := r.Cookie("session")
	if err != nil {
		userID = 0
	} else {
		userID, _, err = app.sessions.GetSession(cookie.Value)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	userFromForum, err := app.forums.GetUserIDFromForum(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.notFound(w)
			return
		}
		app.serverError(w, err)
		return
	}

	isOwn := app.isOwnForum(userFromForum, r)

	forum, err := app.forums.Get(id, userID, isOwn)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)

	data.Form = forum

	app.render(w, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) handleForumCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.ForumCreateGet(w, r)
	case http.MethodPost:
		app.ForumCreatePost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) ForumCreateGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	tags, err := app.forums.GetAllTags()
	if err != nil {
		app.serverError(w, err)
		return
	}

	form := forumCreateForm{
		Expires: 365,
		AllTags: tags,
	}

	data.Form = form

	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) ForumCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxFileSize := int64(20 * 1024 * 1024)
	err = r.ParseMultipartForm(maxFileSize) // check for size (20mb)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var imagePath string
	file, fileHeader, err := r.FormFile("image-upload")
	if err == nil {
		filename := fileHeader.Filename
		extension := filepath.Ext(filename)
		newFilename := strconv.FormatInt(time.Now().UnixNano(), 10) + extension
		imagePath = filepath.Join("/static/images", newFilename)
		imageOut := "/ui" + imagePath

		f, err := os.OpenFile("."+imageOut, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer f.Close()

		fileSize := fileHeader.Size
		if fileSize > maxFileSize {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		_, err = io.Copy(f, file)
		if err != nil {
			app.serverError(w, err)
			return
		}

		defer file.Close()
	}

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadGateway)
		return
	}

	selectedTags := r.Form["tags"]
	customTagsStr := r.PostForm.Get("custom_tags")

	tags := app.processTags(selectedTags, customTagsStr)
	tagsStr := strings.Join(tags, ", ")

	form := forumCreateForm{
		Title:     r.PostForm.Get("title"),
		Content:   r.PostForm.Get("content"),
		Tags:      tagsStr,
		Expires:   expires,
		ImagePath: imagePath,
	}

	form.CheckField(validator.IncorrectInput(form.Tags), "tags", "Incorrect tags formation")
	form.CheckField(validator.MaxChars(form.Tags, 50), "tags", "This field cannot be more than 50 characters long")
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	id, err := app.forums.Insert(form.Title, form.Content, form.Tags, expires, userID, imagePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.AskForNewForum(id, userID, form.Title+"\n"+form.Content)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleForumEdit(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.ForumEditGet(w, r)
	case http.MethodPost:
		app.ForumEditPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) ForumEditGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "edit" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	var userID int

	cookie, err := r.Cookie("session")
	if err != nil {
		userID = 0
	} else {
		userID, _, err = app.sessions.GetSession(cookie.Value)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	userFromForum, err := app.forums.GetUserIDFromForum(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	isOwn := app.isOwnForum(userFromForum, r)

	forum, err := app.forums.Get(id, userID, isOwn)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Forum = forum
	data.Form = forumCreateForm{
		Expires: 365,
	}

	app.render(w, http.StatusOK, "edit.tmpl.html", data)
}

func (app *application) ForumEditPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "edit" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[3]
	forumID, err := strconv.Atoi(idStr)
	if err != nil || forumID < 1 {
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxFileSize := int64(20 * 1024 * 1024)
	err = r.ParseMultipartForm(maxFileSize) // check for size (20mb)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var imagePath string
	file, fileHeader, err := r.FormFile("image-upload")
	if err == nil {
		filename := fileHeader.Filename
		extension := filepath.Ext(filename)
		newFilename := strconv.FormatInt(time.Now().UnixNano(), 10) + extension
		imagePath = filepath.Join("/static/images", newFilename)
		imageOut := "/ui" + imagePath

		f, err := os.OpenFile("."+imageOut, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer f.Close()

		fileSize := fileHeader.Size
		if fileSize > maxFileSize {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		_, err = io.Copy(f, file)
		if err != nil {
			app.serverError(w, err)
			return
		}

		defer file.Close()
	}

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadGateway)
		return
	}

	selectedTags := r.Form["tags"]
	customTagsStr := r.PostForm.Get("custom_tags")

	tags := app.processTags(selectedTags, customTagsStr)
	tagsStr := strings.Join(tags, ", ")

	form := forumCreateForm{
		Title:     r.PostForm.Get("title"),
		Content:   r.PostForm.Get("content"),
		Tags:      tagsStr,
		Expires:   expires,
		ImagePath: imagePath,
	}

	form.CheckField(validator.IncorrectInput(form.Tags), "tags", "Incorrect tags formation")
	form.CheckField(validator.MaxChars(form.Tags, 50), "tags", "This field cannot be more than 50 characters long")
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	userID, _, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.forums.Edit(form.Title, form.Content, form.Tags, expires, userID, imagePath, forumID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", forumID), http.StatusSeeOther)
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
	user_id, _, _ := app.sessions.GetSession(cookie.Value)

	role, err := app.users.GetUserRole(user_id)
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

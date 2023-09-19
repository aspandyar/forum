package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	validator.Validator
}

type userSingupForm struct {
	Name     string
	Email    string
	Password string
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

// Замените на свои данные OAuth 2.0
const (
	clientID        = "30519126384-v31k4ahraui4a59kmev21ju6353ne17p.apps.googleusercontent.com"
	clientGitID     = "d110450fd3d4bae1c7bb"
	clientGitSecret = "460b3d5eba5d619f7463b081a2b211e2083f73cf"

	clientSecret = "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"
	redirectURI  = "http://localhost:4000/callback"
)

// Данные о пользователе
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string
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
	// tagsStr := strings.Join(tags, ", ")

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

	data.Form = forumCreateForm{
		Expires: 365,
	}

	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) ForumCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	maxFileSize := int64(20 * 1024 * 1024)
	err = r.ParseMultipartForm(maxFileSize) //check for size (20mb)
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

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", id), http.StatusSeeOther)
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
	err = r.ParseMultipartForm(maxFileSize) //check for size (20mb)
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

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.userSignupGet(w, r)
	case http.MethodPost:
		app.userSignupPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) userSignupGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSingupForm{}
	app.render(w, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userSingupForm{
		Name:     r.PostForm.Get("name"),
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email or name address is already in use")
			form.AddFieldError("name", "Email or name address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	// Формирование URL для перенаправления пользователя на страницу аутентификации Google
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile", clientID, redirectURI)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (app *application) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Получение кода авторизации из запроса
	code := r.URL.Query().Get("code")

	// Обмен кода авторизации на токен доступа
	tokenURL := "https://accounts.google.com/o/oauth2/token"
	data := fmt.Sprintf("code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code", code, clientID, clientSecret, redirectURI)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResponse map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		http.Error(w, "Failed to decode token response", http.StatusInternalServerError)
		return
	}
	accessToken := tokenResponse["access_token"].(string)

	// Получение данных о пользователе с использованием токена доступа
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, _ := http.NewRequest("GET", userInfoURL, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)

	userInfoResp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer userInfoResp.Body.Close()

	var form userSingupForm
	if err := json.NewDecoder(userInfoResp.Body).Decode(&form); err != nil {
		http.Error(w, "Failed to decode user info response", http.StatusInternalServerError)
		return
	}

	form.Password, err = generateRandomPassword(8)
	if err != nil {
		app.clientError(w, http.StatusInternalServerError)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	// form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	// form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			userID, _ := app.users.Authenticate(form.Email, form.Password)

			session, err := app.sessions.CreateSession(userID)
			if err != nil {
				app.serverError(w, err)
				return
			}
			models.SetSessionCookie(w, session.Token, session.Expiry)

			http.Redirect(w, r, "/forum/create", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
		return
	}

	userID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			http.Redirect(w, r, "/forum/create", http.StatusSeeOther)

		} else {
			app.serverError(w, err)
		}
		return
	}

	session, err := app.sessions.CreateSession(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	models.SetSessionCookie(w, session.Token, session.Expiry)
	http.Redirect(w, r, "/forum/create", http.StatusSeeOther)
}

func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Greate GIT HUB AUTOTEFICATION
func (app *application) loggedinHandler(w http.ResponseWriter, r *http.Request, githubData string) {
	if githubData == "" {
		// Unauthorized users get an unauthorized message
		fmt.Fprintf(w, "UNAUTHORIZED!")
		return
	}

	// Set return type JSON
	w.Header().Set("Content-type", "application/json")

	form := userSingupForm{}
	json.Unmarshal([]byte(githubData), &form)

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	// form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	// form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	err := app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			userID, _ := app.users.Authenticate(form.Email, form.Password)

			session, err := app.sessions.CreateSession(userID)
			if err != nil {
				app.serverError(w, err)
				return
			}
			models.SetSessionCookie(w, session.Token, session.Expiry)

			http.Redirect(w, r, "/forum/create", http.StatusSeeOther)
			return
		} else {
			app.serverError(w, err)
		}
		return
	}

	userID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			http.Redirect(w, r, "/forum/create", http.StatusSeeOther)

		} else {
			app.serverError(w, err)
		}
		return
	}

	session, err := app.sessions.CreateSession(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	models.SetSessionCookie(w, session.Token, session.Expiry)
	http.Redirect(w, r, "/forum/create", http.StatusSeeOther)

}

func (app *application) gitHubLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Create the dynamic redirect URL for login
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		clientGitID,
		"http://localhost:4000/login/github/callback",
	)

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (app *application) gitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	gitHubAccessToken := app.getGitHubAccessToken(code)
	gitHubData := getGitHubData(gitHubAccessToken)
	app.loggedinHandler(w, r, gitHubData)

}
func (app *application) getGitHubAccessToken(code string) string {
	requestBodyMap := map[string]string{
		"client_id":     clientGitID,
		"client_secret": clientGitSecret,
		"code":          code,
	}

	requestJSON, _ := json.Marshal(requestBodyMap)
	req, reqerr := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)

	if reqerr != nil {
		log.Panic("Request creation failed")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Get the response
	resp, resperr := http.DefaultClient.Do(req)

	if resperr != nil {
		log.Panic("Request failed by get the response")
	}

	// Response body converted to stringified JSON
	respbody, _ := ioutil.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	// Return the access token (as the rest of the
	// details are relatively unnecessary for us)
	return ghresp.AccessToken

}
func getGitHubData(accessToken string) string {
	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		log.Panic("API Request creation failed")
	}

	// Set the Authorization header before sending the request
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed by Make the request")
	}

	// Read the response as a byte slice
	respbody, _ := ioutil.ReadAll(resp.Body)

	// Convert byte slice to string and return
	return string(respbody)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.userLoginGet(w, r)
	case http.MethodPost:
		app.userLoginPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) userLoginGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userLoginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	userID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {

			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	session, err := app.sessions.CreateSession(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	models.SetSessionCookie(w, session.Token, session.Expiry)

	http.Redirect(w, r, "/forum/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.sessions.InvalidateSession(cookie.Value)
	if err != nil {
		app.serverError(w, err)
		return
	}

	models.ClearSessionCookie(w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
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

func (app *application) isOwnForum(userID int, r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	getUserID, expiry, err := app.sessions.GetSession(cookie.Value)
	if err != nil {
		return false
	}

	return time.Now().Before(expiry) && userID == getUserID
}

func (app *application) forumIsLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var likeStatus int
	button := r.PostForm.Get("button")
	switch button {
	case "like":
		likeStatus = 1
	case "dislike":
		likeStatus = -1
	default:
		app.clientError(w, http.StatusBadRequest)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "like" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[3]
	forumId, err := strconv.Atoi(idStr)
	if err != nil || forumId < 1 {
		http.NotFound(w, r)
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

	form := forumLikeForm{
		LikeStatus: likeStatus,
		ForumID:    forumId,
		UserID:     userID,
	}

	id, err := app.forumLike.LikeOrDislike(form.ForumID, form.UserID, form.LikeStatus)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", id), http.StatusSeeOther)
}

func (app *application) forumIsLikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var likeStatus int
	button := r.PostForm.Get("button")
	switch button {
	case "like":
		likeStatus = 1
	case "dislike":
		likeStatus = -1
	default:
		app.clientError(w, http.StatusBadRequest)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "likeComment" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[3]
	commentId, err := strconv.Atoi(idStr)
	if err != nil || commentId < 1 {
		http.NotFound(w, r)
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

	form := forumLikeForm{
		LikeStatus: likeStatus,
		CommentID:  commentId,
		UserID:     userID,
	}

	id, err := app.forumLike.LikeOrDislikeComment(form.CommentID, form.UserID, form.LikeStatus)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", id), http.StatusSeeOther)
}

func (app *application) handleForumComment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.forumView(w, r)
	case http.MethodPost:
		app.ForumCommentPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) ForumCommentPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	comment := r.PostForm.Get("comment")

	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[3]
	forumId, err := strconv.Atoi(idStr)
	if err != nil || forumId < 1 {
		http.NotFound(w, r)
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

	form := forumCommentForm{
		ForumID: forumId,
		UserID:  userID,
		Comment: comment,
	}

	form.CheckField(validator.NotBlank(form.Comment), "comment", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "view.tmpl.html", data)
		return
	}

	id, err := app.forumComment.CommentPost(form.ForumID, form.UserID, form.Comment)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", id), http.StatusSeeOther)
}

func (app *application) handleForumEditComment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.ForumEditCommentGet(w, r)
	case http.MethodPost:
		app.ForumEditCommentPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) ForumEditCommentGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 6 || parts[1] != "forum" || parts[2] != "comment" || parts[3] != "edit" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[4]
	forumID, err := strconv.Atoi(idStr)
	if err != nil || forumID < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[5]
	commentID, err := strconv.Atoi(idStr)
	if err != nil || commentID < 1 {
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

	userFromForum, err := app.forums.GetUserIDFromForum(forumID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	isOwn := app.isOwnForum(userFromForum, r)

	forum, err := app.forums.GetEdit(forumID, userID, isOwn, commentID)
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

	app.render(w, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) ForumEditCommentPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	comment := r.PostForm.Get("comment")

	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[4]
	forumId, err := strconv.Atoi(idStr)
	if err != nil || forumId < 1 {
		http.NotFound(w, r)
		return
	}

	idStr = parts[5]
	commentID, err := strconv.Atoi(idStr)
	if err != nil || commentID < 1 {
		http.NotFound(w, r)
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

	form := forumCommentForm{
		ForumID: forumId,
		UserID:  userID,
		Comment: comment,
	}

	form.CheckField(validator.NotBlank(form.Comment), "comment", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "view.tmpl.html", data)
		return
	}

	err = app.forumComment.EditCommentPost(form.ForumID, form.UserID, form.Comment, commentID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", form.ForumID), http.StatusSeeOther)
}

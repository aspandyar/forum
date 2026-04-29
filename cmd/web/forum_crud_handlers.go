package main

import (
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
	form := forumCreateForm{Expires: 365, AllTags: tags}
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
	err = r.ParseMultipartForm(maxFileSize)
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
		if fileHeader.Size > maxFileSize {
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
	tags := app.processTags(r.Form["tags"], r.PostForm.Get("custom_tags"))
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
	role, err := app.forums.GetRoleByUserID(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if role == adminRole {
		err = app.forums.ChangeForumStatus(id, visibleStatus)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		err = app.forums.AskForNewForum(id, userID, form.Title+"\n"+form.Content)
		if err != nil {
			app.serverError(w, err)
			return
		}
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
	id, err := strconv.Atoi(parts[3])
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
	data.Form = forumCreateForm{Expires: 365}
	app.render(w, http.StatusOK, "edit.tmpl.html", data)
}

func (app *application) ForumEditPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "edit" {
		http.NotFound(w, r)
		return
	}
	forumID, err := strconv.Atoi(parts[3])
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
	err = r.ParseMultipartForm(maxFileSize)
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
		if fileHeader.Size > maxFileSize {
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
	tags := app.processTags(r.Form["tags"], r.PostForm.Get("custom_tags"))
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

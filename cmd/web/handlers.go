package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/aspandyar/internal/models"
)

type forumCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
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

func (app *application) forumView(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "view" { //localhost:????/forum/view/id, so len should be 4, parts[1] = forum and parts[2] is view
		http.NotFound(w, r)
		return
	}

	idStr := parts[3] //parts[3] is id
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	forum, err := app.forums.Get(id)
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
		app.handleForumCreateGet(w, r)
	case http.MethodPost:
		app.handleForumCreatePost(w, r)
	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) handleForumCreateGet(w http.ResponseWriter, r *http.Request) {
	// Handle GET request for forum create
	data := app.newTemplateData(r)

	data.Form = forumCreateForm{
		Expires: 365,
	}

	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) handleForumCreatePost(w http.ResponseWriter, r *http.Request) {
	// Handle POST request for forum create
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadGateway)
		return
	}

	form := forumCreateForm{
		Title:       r.PostForm.Get("title"),
		Content:     r.PostForm.Get("content"),
		Expires:     expires,
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Title) == "" {
		form.FieldErrors["title"] = "This field can't be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		form.FieldErrors["title"] = "This field can't be more than 100 characters long"
	}

	if strings.TrimSpace(form.Content) == "" {
		form.FieldErrors["content"] = "This field can't be blank"
	}

	if form.Expires != 1 && form.Expires != 7 && form.Expires != 365 {
		form.FieldErrors["expires"] = "This field must be equal 1, 7, or 365"
	}

	if len(form.FieldErrors) > 0 {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	id, err := app.forums.Insert(form.Title, form.Content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", id), http.StatusSeeOther)
}

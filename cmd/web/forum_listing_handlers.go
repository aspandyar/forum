package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/aspandyar/forum/internal/models"
)

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
	userID, err := app.sessionUserID(r)
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
	userID, err := app.sessionUserID(r)
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
	userID, err := app.sessionUserID(r)
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
	parts := pathParts(r)
	if len(parts) != 4 || parts[1] != "forum" || parts[2] != "view" {
		http.NotFound(w, r)
		return
	}
	id, err := pathInt(parts, 3)
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

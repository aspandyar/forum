package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aspandyar/forum/internal/validator"
)

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
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[4])
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
	parts := strings.Split(r.URL.Path, "/")
	getUserID, err := strconv.Atoi(parts[6])
	if err != nil || getUserID < 1 {
		http.NotFound(w, r)
		return
	}
	forumID, err := strconv.Atoi(parts[5])
	if err != nil || forumID < 1 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[4])
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	if err = app.forums.ChangeForumStatus(forumID, invisibleStatus); err != nil {
		app.serverError(w, err)
		return
	}
	if err = app.forums.AnswerFromAdmin(getUserID, "approved"); err != nil {
		app.serverError(w, err)
		return
	}
	if err = app.forums.RemoveUserNotification(id); err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) userModerationDone(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	userID, err := strconv.Atoi(parts[4])
	if err != nil || userID < 1 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	if err = app.forums.ChangeUserRole(userID, moderatorRole); err != nil {
		app.serverError(w, err)
		return
	}
	if err = app.forums.RemoveUserNotification(id); err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) moderDenoteHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	moderID, err := strconv.Atoi(parts[3])
	if err != nil || moderID < 1 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[4])
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	if err = app.forums.ChangeUserRole(moderID, userRole); err != nil {
		app.serverError(w, err)
		return
	}
	if err = app.forums.RemoveUserNotification(id); err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/user/notification", http.StatusSeeOther)
}

func (app *application) forumAcceptHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	notID, err := strconv.Atoi(parts[3])
	if err != nil || notID < 1 {
		http.NotFound(w, r)
		return
	}
	fourmID, err := strconv.Atoi(parts[4])
	if err != nil || fourmID < 1 {
		http.NotFound(w, r)
		return
	}
	if err = app.forums.ChangeForumStatus(fourmID, visibleStatus); err != nil {
		app.serverError(w, err)
		return
	}
	if err = app.forums.RemoveUserNotification(notID); err != nil {
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
	tags := strings.Split(r.FormValue("tags_text"), ", ")
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
	parts := strings.Split(r.URL.Path, "/")
	forumID, err := strconv.Atoi(parts[3])
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
	http.Redirect(w, r, "/forum/view/"+strconv.Itoa(forumID), http.StatusSeeOther)
}

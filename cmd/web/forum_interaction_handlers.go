package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aspandyar/forum/internal/models"
	"github.com/aspandyar/forum/internal/validator"
)

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

	userID, err := app.sessionUserID(r)
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

	userID, err := app.sessionUserID(r)
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

	userID, err := app.sessionUserID(r)
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

	userID, err := app.sessionUserID(r)
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

func (app *application) ForumRemoveCommentPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	idStr := parts[4]
	forumID, err := strconv.Atoi(idStr)
	if err != nil || forumID < 1 {
		return
	}

	idStr = parts[5]
	commentID, err := strconv.Atoi(idStr)
	if err != nil || commentID < 1 {
		return
	}

	userFromForum, err := app.forums.GetUserIDFromComment(commentID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	isOwn := app.isOwnForum(userFromForum, r)
	if !isOwn {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err = app.forumComment.RemoveCommentPost(commentID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/view/%d", forumID), http.StatusSeeOther)
}

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aspandyar/forum/internal/models"
	githuboauth "github.com/aspandyar/forum/internal/oauth/github"
	googleoauth "github.com/aspandyar/forum/internal/oauth/google"
	"github.com/aspandyar/forum/internal/validator"
)

func (app *application) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile", clientID, redirectURI)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (app *application) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	info, err := googleoauth.FetchUserInfo(clientID, clientSecret, redirectURI, code)
	if err != nil {
		http.Error(w, "Failed to complete Google auth", http.StatusInternalServerError)
		return
	}

	form := userSingupForm{Name: info.Name, Email: info.Email}

	form.Role = userRole
	form.Password, _ = generateRandomPassword(8)
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password, form.Role)
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

func (app *application) loggedinHandler(w http.ResponseWriter, r *http.Request, githubData string) {
	if githubData == "" {
		app.serverError(w, errors.New("github.com: unauthorized"))
		return
	}

	w.Header().Set("Content-type", "application/json")
	form := userSingupForm{}
	json.Unmarshal([]byte(githubData), &form)

	form.Role = userRole
	form.Password, _ = generateRandomPassword(8)
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	err := app.users.Insert(form.Name, form.Email, form.Password, 2)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			userID, err := app.users.Authenticate(form.Email, form.Password)
			if err != nil {
				if errors.Is(err, models.ErrInvalidCredentials) {
					http.Redirect(w, r, "/user/login", http.StatusSeeOther)
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
			return
		}
		app.serverError(w, err)
		return
	}

	userID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
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
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		clientGitID,
		"https://localhost:4000/login/github/callback",
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
	token, err := githuboauth.AccessToken(clientGitID, clientGitSecret, code)
	if err != nil {
		panic(err)
	}
	return token
}

func getGitHubData(accessToken string) string {
	data, _ := githuboauth.UserData(accessToken)
	return data
}

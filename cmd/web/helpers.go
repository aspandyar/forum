package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	renderpkg "github.com/aspandyar/forum/internal/transport/http/render"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, statuc int) {
	http.Error(w, http.StatusText(statuc), statuc)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	err := renderpkg.Render(w, app.tempalteCache, status, page, data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           "",
		IsAuthenticated: app.isAuthenticated(r),
		Role:            app.getRole(r),
	}
}

func (app *application) processTags(selectedTags []string, customTagsStr string) []string {
	customTags := strings.Split(customTagsStr, ",")

	uniqueTags := make(map[string]bool)

	for _, tag := range selectedTags {
		if tag != "" && tag != " " {
			uniqueTags[tag] = true
		}
	}

	for _, tag := range customTags {
		tag = strings.TrimSpace(tag)
		if tag != "" { // Skip empty elements
			uniqueTags[tag] = true
		}
	}

	var tags []string
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}

	return tags
}

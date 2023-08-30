package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
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
	ts, ok := app.tempalteCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	// w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           "",
		IsAuthenticated: app.isAuthenticated(r),
	}
}

func (app *application) processTags(selectedTags []string, customTagsStr string) []string {
	// Split customTagsStr into a slice of tags
	customTags := strings.Split(customTagsStr, ",")

	// Create a map to store unique tags
	uniqueTags := make(map[string]bool)

	// Add selectedTags to the map
	for _, tag := range selectedTags {
		if tag != "" && tag != " " { // Skip empty or space elements
			uniqueTags[tag] = true
		}
	}

	// Add customTags to the map
	for _, tag := range customTags {
		tag = strings.TrimSpace(tag)
		if tag != "" { // Skip empty elements
			uniqueTags[tag] = true
		}
	}

	// Convert uniqueTags map keys back to a slice of strings
	var tags []string
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}

	return tags
}

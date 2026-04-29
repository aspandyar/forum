package render

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/aspandyar/forum/internal/models"
)

type TemplateData struct {
	CurrentYear     int
	Forum           *models.Forum
	Forums          []*models.Forum
	Form            interface{}
	Flash           string
	IsAuthenticated bool
	Role            int
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func Render(w http.ResponseWriter, cache map[string]*template.Template, status int, page string, data *TemplateData) error {
	ts, ok := cache[page]
	if !ok {
		return fmt.Errorf("the template %s does not exist", page)
	}

	buf := new(bytes.Buffer)
	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
		return err
	}

	_ = status
	buf.WriteTo(w)
	return nil
}

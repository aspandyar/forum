package render

import (
	"net/http/httptest"
	"testing"
	"text/template"
	"time"
)

func TestHumanDateFormat(t *testing.T) {
	got := humanDate(time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC))
	if got != "02 Jan 2026 at 15:04" {
		t.Fatalf("humanDate=%q", got)
	}
}

func TestNewTemplateCacheAndRender(t *testing.T) {
	cache, err := NewTemplateCache()
	if err != nil {
		t.Fatalf("NewTemplateCache: %v", err)
	}
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	if len(cache) == 0 {
		return
	}

	var anyPage string
	for k := range cache {
		anyPage = k
		break
	}
	rr := httptest.NewRecorder()
	if err := Render(rr, cache, 200, anyPage, &TemplateData{}); err != nil {
		t.Fatalf("Render with valid page: %v", err)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("expected rendered output")
	}

	rr = httptest.NewRecorder()
	if err := Render(rr, cache, 200, "missing-page.tmpl.html", &TemplateData{}); err == nil {
		t.Fatal("expected missing template error")
	}
}

func TestRender_WithManualTemplates(t *testing.T) {
	okTpl := template.Must(template.New("ok").Parse(`{{define "base"}}render-ok{{end}}`))
	cache := map[string]*template.Template{
		"ok.tmpl.html": okTpl,
	}

	rr := httptest.NewRecorder()
	if err := Render(rr, cache, 200, "ok.tmpl.html", &TemplateData{}); err != nil {
		t.Fatalf("Render ok template err=%v", err)
	}
	if rr.Body.String() != "render-ok" {
		t.Fatalf("unexpected render body %q", rr.Body.String())
	}

	badTpl := template.Must(template.New("bad").Parse(`{{define "not_base"}}x{{end}}`))
	cache["bad.tmpl.html"] = badTpl
	rr = httptest.NewRecorder()
	if err := Render(rr, cache, 200, "bad.tmpl.html", &TemplateData{}); err == nil {
		t.Fatal("expected execute template error for missing base")
	}
}

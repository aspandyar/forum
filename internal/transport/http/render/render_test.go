package render

import (
	"net/http/httptest"
	"testing"
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

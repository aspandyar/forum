package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestClientAndNotFoundHelpers(t *testing.T) {
	app, _ := newWebTestApp(t)

	_, rr := newRequest("GET", "/", nil)
	app.clientError(rr, 418)
	if rr.Code != 418 {
		t.Fatalf("clientError status=%d, want 418", rr.Code)
	}

	_, rr = newRequest("GET", "/", nil)
	app.notFound(rr)
	if rr.Code != 404 {
		t.Fatalf("notFound status=%d, want 404", rr.Code)
	}
}

func TestServerErrorAndRenderBranches(t *testing.T) {
	app, _ := newWebTestApp(t)

	_, rr := newRequest("GET", "/", nil)
	app.serverError(rr, errors.New("boom"))
	if rr.Code != 500 {
		t.Fatalf("serverError status=%d, want 500", rr.Code)
	}

	_, rr = newRequest("GET", "/", nil)
	app.render(rr, 200, "missing.tmpl.html", &templateData{})
	if rr.Code != 500 {
		t.Fatalf("render missing template status=%d, want 500", rr.Code)
	}

	tpl := template.Must(template.New("base").Parse(`{{define "base"}}ok{{end}}`))
	app.tempalteCache["ok.tmpl.html"] = tpl
	_, rr = newRequest("GET", "/", nil)
	app.render(rr, 200, "ok.tmpl.html", &templateData{})
	if !strings.Contains(rr.Body.String(), "ok") {
		t.Fatalf("expected rendered body to contain ok, got %q", rr.Body.String())
	}
}

func TestNewTemplateDataAndLoadEnvFromFile(t *testing.T) {
	app, _ := newWebTestApp(t)
	req, _ := newRequest("GET", "/", nil)
	data := app.newTemplateData(req)
	if data.CurrentYear != time.Now().Year() {
		t.Fatalf("CurrentYear=%d, want %d", data.CurrentYear, time.Now().Year())
	}
	if data.IsAuthenticated {
		t.Fatal("expected unauthenticated in new request")
	}

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("K1=V1\nK2 = V2\nBADLINE\n"), 0600); err != nil {
		t.Fatalf("write temp env: %v", err)
	}
	if err := LoadEnvFromFile(envPath); err != nil {
		t.Fatalf("LoadEnvFromFile: %v", err)
	}
	if os.Getenv("K1") != "V1" || os.Getenv("K2") != "V2" {
		t.Fatalf("expected loaded env vars, got K1=%q K2=%q", os.Getenv("K1"), os.Getenv("K2"))
	}
	if err := LoadEnvFromFile(filepath.Join(t.TempDir(), "missing.env")); err == nil {
		t.Fatal("expected error for missing env file")
	}
}

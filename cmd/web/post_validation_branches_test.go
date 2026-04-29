package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestAuthPostValidationBranches(t *testing.T) {
	app, _ := newWebTestApp(t)
	addBaseTemplate(app, "signup.tmpl.html")
	addBaseTemplate(app, "login.tmpl.html")

	req, rr := newRequest(http.MethodPost, "/user/signup", strings.NewReader("name=&email=bad&password=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.userSignupPost(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("userSignupPost validation status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/user/login", strings.NewReader("email=bad&password="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.userLoginPost(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("userLoginPost validation status=%d", rr.Code)
	}
}

func TestModerationPostValidationBranches(t *testing.T) {
	app, _ := newWebTestApp(t)
	addBaseTemplate(app, "report.tmpl.html")

	req, rr := newRequest(http.MethodPost, "/moderation/report/1", strings.NewReader("reportType="+strings.Repeat("x", 300)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ForumReportPost(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("ForumReportPost validation status=%d", rr.Code)
	}

	req, rr = newRequest(http.MethodPost, "/admin/addTags", strings.NewReader("add_tag=1&remove_tag=-1&tags_text=go"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.addTagsPost(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("addTagsPost conflicting actions status=%d", rr.Code)
	}
}

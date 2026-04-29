package main

import (
	"crypto/tls"
	"net/http"
	"strings"
	"testing"
)

func TestTLSConfigHelpers(t *testing.T) {
	cfg := &tls.Config{}
	ConfigureCipherSuites(cfg, []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256})
	SetMinTLSVersion(cfg, tls.VersionTLS12)
	SetMaxTLSVersion(cfg, tls.VersionTLS13)
	if len(cfg.CipherSuites) != 1 || cfg.MinVersion != tls.VersionTLS12 || cfg.MaxVersion != tls.VersionTLS13 {
		t.Fatalf("unexpected tls config: %+v", cfg)
	}
}

func TestSwaggerHandlers(t *testing.T) {
	app, _ := newWebTestApp(t)

	req, rr := newRequest(http.MethodGet, "/openapi.yaml", nil)
	app.openapiYAML(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("openapi GET status=%d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/yaml") {
		t.Fatalf("unexpected content type: %q", rr.Header().Get("Content-Type"))
	}

	req, rr = newRequest(http.MethodPost, "/openapi.yaml", nil)
	app.openapiYAML(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("openapi POST status=%d, want 405", rr.Code)
	}

	req, rr = newRequest(http.MethodGet, "/swagger/", nil)
	app.swaggerUI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("swagger GET status=%d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("expected swagger html payload, got %q", rr.Body.String())
	}

	req, rr = newRequest(http.MethodGet, "/swagger/nope", nil)
	app.swaggerUI(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("swagger wrong path status=%d, want 404", rr.Code)
	}
}

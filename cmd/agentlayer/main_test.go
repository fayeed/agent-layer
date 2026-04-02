package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServerExposesHealthEndpoint(t *testing.T) {
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected health endpoint to return 200, got %d", recorder.Code)
	}

	if recorder.Body.String() != "ok\n" {
		t.Fatalf("expected health response body, got %q", recorder.Body.String())
	}
}

func TestNewServerRegistersV0RouteShapes(t *testing.T) {
	server := newServer()

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/threads/thread-123/reply"},
		{method: http.MethodPost, path: "/threads/thread-123/escalate"},
		{method: http.MethodGet, path: "/threads/thread-123"},
		{method: http.MethodGet, path: "/threads/thread-123/messages"},
		{method: http.MethodGet, path: "/contacts/contact-123"},
		{method: http.MethodPost, path: "/contacts/contact-123/memory"},
		{method: http.MethodPost, path: "/provider/callbacks/outbound"},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		if recorder.Code == http.StatusNotFound {
			t.Fatalf("expected route %s %s to be registered", tt.method, tt.path)
		}
	}
}

func TestNewServerWiresThreadHandler(t *testing.T) {
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected real thread handler path to return 500 from placeholder service, got %d", recorder.Code)
	}
}

func TestNewServerWiresContactHandler(t *testing.T) {
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/contacts/contact-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected real contact handler path to return 500 from placeholder service, got %d", recorder.Code)
	}
}

func TestNewServerWiresRemainingHandlers(t *testing.T) {
	server := newServer()

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodPost, path: "/threads/thread-123/reply", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/threads/thread-123/escalate", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/threads/thread-123/messages", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/contacts/contact-123/memory", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/provider/callbacks/outbound", body: "{}", want: http.StatusBadRequest},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		if recorder.Code != tt.want {
			t.Fatalf("expected route %s %s to return %d from wired handler, got %d", tt.method, tt.path, tt.want, recorder.Code)
		}
	}
}

func TestNewSMTPServerUsesDefaults(t *testing.T) {
	t.Setenv("AGENTLAYER_SMTP_ADDR", "")
	t.Setenv("AGENTLAYER_SMTP_DOMAIN", "")

	server := newSMTPServer()

	if server.Addr != "localhost:2525" {
		t.Fatalf("expected default smtp addr, got %q", server.Addr)
	}

	if server.Domain != "localhost" {
		t.Fatalf("expected default smtp domain, got %q", server.Domain)
	}
}

func TestNewSMTPServerUsesEnvOverrides(t *testing.T) {
	t.Setenv("AGENTLAYER_SMTP_ADDR", "127.0.0.1:2626")
	t.Setenv("AGENTLAYER_SMTP_DOMAIN", "mail.agentlayer.dev")

	server := newSMTPServer()

	if server.Addr != "127.0.0.1:2626" {
		t.Fatalf("expected configured smtp addr, got %q", server.Addr)
	}

	if server.Domain != "mail.agentlayer.dev" {
		t.Fatalf("expected configured smtp domain, got %q", server.Domain)
	}
}

func TestSMTPAddressHelpers(t *testing.T) {
	t.Setenv("AGENTLAYER_SMTP_ADDR", "0.0.0.0:2526")
	t.Setenv("AGENTLAYER_SMTP_DOMAIN", "smtp.example.com")

	if got := smtpAddress(); got != "0.0.0.0:2526" {
		t.Fatalf("expected smtp addr helper to read env, got %q", got)
	}

	if got := smtpDomain(); got != "smtp.example.com" {
		t.Fatalf("expected smtp domain helper to read env, got %q", got)
	}
}

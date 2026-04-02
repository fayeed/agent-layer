package main

import (
	"net/http"
	"net/http/httptest"
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

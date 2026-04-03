package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBootstrapReadHandlerReturnsBootstrapConfig(t *testing.T) {
	handler := NewBootstrapReadHandler(&bootstrapReadServiceStub{
		result: BootstrapResult{
			OrganizationID: "org-local",
			AgentID:        "agent-local",
			InboxID:        "inbox-local",
			WebhookURL:     "https://example.com/webhook",
			InboxAddress:   "agent@example.com",
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/bootstrap", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", recorder.Code)
	}

	var response bootstrapResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}

	if response.OrganizationID != "org-local" || response.AgentID != "agent-local" || response.InboxID != "inbox-local" {
		t.Fatalf("expected bootstrap ids in response, got %#v", response)
	}
}

type bootstrapReadServiceStub struct {
	result BootstrapResult
	err    error
}

func (s *bootstrapReadServiceStub) GetBootstrap(context.Context) (BootstrapResult, error) {
	return s.result, s.err
}

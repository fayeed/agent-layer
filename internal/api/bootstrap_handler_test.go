package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestBootstrapHandlerBootstrapsLocalRuntime(t *testing.T) {
	service := &bootstrapServiceStub{
		result: BootstrapResult{
			OrganizationID: "org-local",
			AgentID:        "agent-local",
			InboxID:        "inbox-local",
			WebhookURL:     "https://example.com/webhook",
			InboxAddress:   "agent@example.com",
		},
	}
	handler := NewBootstrapHandler(service)

	request := httptest.NewRequest(http.MethodPost, "/bootstrap", bytes.NewBufferString(`{
		"organization_name":"Acme Support",
		"agent_name":"Acme Agent",
		"agent_status":"active",
		"webhook_url":"https://example.com/webhook",
		"webhook_secret":"super-secret",
		"inbox_address":"agent@example.com",
		"inbox_domain":"example.com",
		"inbox_display_name":"Acme Inbox"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d", recorder.Code)
	}

	if service.input.AgentStatus != string(domain.AgentStatusActive) {
		t.Fatalf("expected agent status to be forwarded, got %#v", service.input)
	}

	var response bootstrapResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}

	if response.OrganizationID != "org-local" || response.AgentID != "agent-local" || response.InboxID != "inbox-local" {
		t.Fatalf("expected bootstrap ids in response, got %#v", response)
	}
}

func TestBootstrapHandlerRejectsInvalidAgentStatus(t *testing.T) {
	handler := NewBootstrapHandler(&bootstrapServiceStub{})
	request := httptest.NewRequest(http.MethodPost, "/bootstrap", bytes.NewBufferString(`{
		"agent_status":"sleeping"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid status to return 400, got %d", recorder.Code)
	}
}

type bootstrapServiceStub struct {
	input  BootstrapInput
	result BootstrapResult
	err    error
}

func (s *bootstrapServiceStub) BootstrapLocal(_ context.Context, input BootstrapInput) (BootstrapResult, error) {
	s.input = input
	return s.result, s.err
}

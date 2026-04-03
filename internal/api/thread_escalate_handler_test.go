package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestThreadEscalateHandlerEscalatesThread(t *testing.T) {
	service := &threadEscalationServiceStub{
		thread: domain.Thread{
			ID:             "thread-123",
			OrganizationID: "org-123",
			AgentID:        "agent-123",
			InboxID:        "inbox-123",
			ContactID:      "contact-123",
			State:          domain.ThreadStateEscalated,
		},
	}

	handler := NewThreadEscalateHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/escalate", bytes.NewBufferString(`{
		"reason":"needs human review"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected accepted status, got %d", recorder.Code)
	}

	if service.threadID != "thread-123" {
		t.Fatalf("expected thread id from path, got %q", service.threadID)
	}

	if service.reason != "needs human review" {
		t.Fatalf("expected escalation reason from payload, got %q", service.reason)
	}

	var response threadResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.State != string(domain.ThreadStateEscalated) {
		t.Fatalf("expected escalated thread state in response, got %#v", response)
	}
}

func TestThreadEscalateHandlerRejectsInvalidPayload(t *testing.T) {
	handler := NewThreadEscalateHandler(&threadEscalationServiceStub{})
	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/escalate", bytes.NewBufferString(`{`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestThreadEscalateHandlerReturnsNotFoundForMissingThread(t *testing.T) {
	handler := NewThreadEscalateHandler(&threadEscalationServiceStub{err: domain.ErrNotFound})
	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/escalate", bytes.NewBufferString(`{
		"reason":"needs human review"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewThreadEscalateHandler(&threadEscalationServiceStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodPost, "/threads/thread-123/escalate", bytes.NewBufferString(`{
		"reason":"needs human review"
	}`))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error status, got %d", recorder.Code)
	}
}

type threadEscalationServiceStub struct {
	threadID string
	reason   string
	thread   domain.Thread
	err      error
}

func (s *threadEscalationServiceStub) EscalateThread(_ context.Context, threadID, reason string) (domain.Thread, error) {
	s.threadID = threadID
	s.reason = reason
	return s.thread, s.err
}

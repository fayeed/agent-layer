package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestThreadHandlerGetsThreadByID(t *testing.T) {
	service := &threadServiceStub{
		thread: domain.Thread{
			ID:                "thread-123",
			OrganizationID:    "org-123",
			AgentID:           "agent-123",
			InboxID:           "inbox-123",
			ContactID:         "contact-123",
			SubjectNormalized: "hello world",
			State:             domain.ThreadStateActive,
			LastInboundID:     "message-100",
			LastOutboundID:    "message-101",
			CreatedAt:         time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC),
			UpdatedAt:         time.Date(2026, 4, 3, 9, 5, 0, 0, time.UTC),
		},
	}

	handler := NewThreadHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	if service.threadID != "thread-123" {
		t.Fatalf("expected thread id from path, got %q", service.threadID)
	}

	var response threadResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.ID != "thread-123" {
		t.Fatalf("expected thread id in response, got %#v", response)
	}

	if response.State != string(domain.ThreadStateActive) {
		t.Fatalf("expected thread state in response, got %#v", response)
	}
}

func TestThreadHandlerRejectsInvalidPath(t *testing.T) {
	handler := NewThreadHandler(&threadServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/threads", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestThreadHandlerReturnsNotFoundForMissingThread(t *testing.T) {
	handler := NewThreadHandler(&threadServiceStub{err: errors.New("wrap: " + domain.ErrNotFound.Error())})
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected wrapped generic error to return 500, got %d", recorder.Code)
	}

	handler = NewThreadHandler(&threadServiceStub{err: domain.ErrNotFound})
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}
}

type threadServiceStub struct {
	threadID string
	thread   domain.Thread
	err      error
}

func (s *threadServiceStub) GetThread(_ context.Context, threadID string) (domain.Thread, error) {
	s.threadID = threadID
	return s.thread, s.err
}

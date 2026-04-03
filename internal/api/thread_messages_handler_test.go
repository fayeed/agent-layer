package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestThreadMessagesHandlerGetsMessagesByThreadID(t *testing.T) {
	service := &threadMessagesServiceStub{
		messages: []domain.Message{
			{
				ID:              "message-100",
				ThreadID:        "thread-123",
				Direction:       domain.MessageDirectionInbound,
				Subject:         "Hello World",
				TextBody:        "Inbound message",
				MessageIDHeader: "<message-100@example.com>",
				CreatedAt:       time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC),
			},
			{
				ID:              "message-101",
				ThreadID:        "thread-123",
				Direction:       domain.MessageDirectionOutbound,
				Subject:         "Re: Hello World",
				TextBody:        "Outbound reply",
				MessageIDHeader: "<message-101@example.com>",
				CreatedAt:       time.Date(2026, 4, 3, 9, 5, 0, 0, time.UTC),
			},
		},
	}

	handler := NewThreadMessagesHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123/messages", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	if service.threadID != "thread-123" {
		t.Fatalf("expected thread id from path, got %q", service.threadID)
	}

	var response []messageResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 messages in response, got %#v", response)
	}

	if response[0].ID != "message-100" || response[1].ID != "message-101" {
		t.Fatalf("expected ordered message payloads, got %#v", response)
	}

	if service.limit != 0 {
		t.Fatalf("expected default handler limit to pass through as zero, got %d", service.limit)
	}
}

func TestThreadMessagesHandlerPassesLimitQuery(t *testing.T) {
	service := &threadMessagesServiceStub{}
	handler := NewThreadMessagesHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123/messages?limit=5", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	if service.limit != 5 {
		t.Fatalf("expected query limit to be forwarded, got %d", service.limit)
	}
}

func TestThreadMessagesHandlerRejectsInvalidLimit(t *testing.T) {
	handler := NewThreadMessagesHandler(&threadMessagesServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123/messages?limit=nope", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestThreadMessagesHandlerRejectsInvalidPath(t *testing.T) {
	handler := NewThreadMessagesHandler(&threadMessagesServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

type threadMessagesServiceStub struct {
	threadID string
	limit    int
	messages []domain.Message
	err      error
}

func (s *threadMessagesServiceStub) ListThreadMessages(_ context.Context, threadID string, limit int) ([]domain.Message, error) {
	s.threadID = threadID
	s.limit = limit
	return s.messages, s.err
}

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
	"github.com/agentlayer/agentlayer/internal/outbound"
)

func TestReplyHandlerPostsThreadReply(t *testing.T) {
	service := &replyServiceStub{
		result: outbound.SendReplyResult{
			Thread: domain.Thread{
				ID: "thread-123",
			},
			Message: domain.Message{
				ID:                "message-123",
				ThreadID:          "thread-123",
				Subject:           "Re: Hello World",
				MessageIDHeader:   "<reply-123@agentlayer.local>",
				DeliveryState:     outbound.DeliveryStateSent,
				ProviderMessageID: "ses-123",
			},
		},
	}

	handler := NewReplyHandler(service)

	body := bytes.NewBufferString(`{
		"organization_id":"org-123",
		"agent_id":"agent-123",
		"inbox_id":"inbox-123",
		"contact_id":"contact-123",
		"idempotency_key":"reply-req-123",
		"reply_to_message_id":"message-100",
		"body_text":"Thanks for reaching out.",
		"object_key":"outbound/reply-123.eml"
	}`)

	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", body)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected accepted status, got %d", recorder.Code)
	}

	if service.input.Thread.ID != "thread-123" {
		t.Fatalf("expected thread id from url, got %#v", service.input)
	}

	if service.input.BodyText != "Thanks for reaching out." {
		t.Fatalf("expected body text from request, got %#v", service.input)
	}
	if service.input.IdempotencyKey != "reply-req-123" {
		t.Fatalf("expected idempotency key from request, got %#v", service.input)
	}

	var response replyResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.MessageID != "message-123" {
		t.Fatalf("expected response message id, got %#v", response)
	}

	if response.ProviderMessageID != "ses-123" {
		t.Fatalf("expected provider message id in response, got %#v", response)
	}
}

func TestReplyHandlerUsesIdempotencyHeaderFallback(t *testing.T) {
	service := &replyServiceStub{
		result: outbound.SendReplyResult{
			Message: domain.Message{ID: "message-123", ThreadID: "thread-123"},
		},
	}
	handler := NewReplyHandler(service)

	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(`{
		"reply_to_message_id":"message-100",
		"body_text":"Thanks for reaching out.",
		"object_key":"outbound/reply-123.eml"
	}`))
	request.Header.Set("Idempotency-Key", "reply-header-123")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if service.input.IdempotencyKey != "reply-header-123" {
		t.Fatalf("expected idempotency key from header fallback, got %#v", service.input)
	}
}

func TestReplyHandlerRejectsInvalidPayload(t *testing.T) {
	handler := NewReplyHandler(&replyServiceStub{})

	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(`{`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestReplyHandlerRequiresReplyContextFields(t *testing.T) {
	handler := NewReplyHandler(&replyServiceStub{})

	tests := []struct {
		body string
	}{
		{body: `{"body_text":"Hello","object_key":"outbound/reply.eml"}`},
		{body: `{"reply_to_message_id":"message-100","object_key":"outbound/reply.eml"}`},
		{body: `{"reply_to_message_id":"message-100","body_text":"Hello"}`},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(tt.body))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request status for body %s, got %d", tt.body, recorder.Code)
		}
	}
}

func TestReplyHandlerReturnsNotFoundForMissingReplyContext(t *testing.T) {
	handler := NewReplyHandler(&replyServiceStub{err: domain.ErrNotFound})

	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(`{
		"reply_to_message_id":"message-100",
		"body_text":"Thanks for reaching out.",
		"object_key":"outbound/reply-123.eml"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewReplyHandler(&replyServiceStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(`{
		"reply_to_message_id":"message-100",
		"body_text":"Thanks for reaching out.",
		"object_key":"outbound/reply-123.eml"
	}`))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error status, got %d", recorder.Code)
	}
}

func TestReplyHandlerReturnsConflictForSuppressedRecipient(t *testing.T) {
	handler := NewReplyHandler(&replyServiceStub{err: outbound.ErrRecipientSuppressed})

	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(`{
		"reply_to_message_id":"message-100",
		"body_text":"Thanks for reaching out.",
		"object_key":"outbound/reply-123.eml"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected conflict status for suppressed recipient, got %d", recorder.Code)
	}
}

type replyServiceStub struct {
	input  outbound.SendReplyInput
	result outbound.SendReplyResult
	err    error
}

func (s *replyServiceStub) SendReply(_ context.Context, input outbound.SendReplyInput) (outbound.SendReplyResult, error) {
	s.input = input
	return s.result, s.err
}

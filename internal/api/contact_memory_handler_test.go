package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestContactMemoryHandlerCreatesMemoryEntry(t *testing.T) {
	service := &contactMemoryServiceStub{
		entry: domain.ContactMemoryEntry{
			ID:        "memory-123",
			ContactID: "contact-123",
			ThreadID:  "thread-123",
			Note:      "Prefers email follow-up.",
			Tags:      []string{"preference"},
			CreatedAt: time.Date(2026, 4, 3, 11, 0, 0, 0, time.UTC),
		},
	}

	handler := NewContactMemoryHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/contacts/contact-123/memory", bytes.NewBufferString(`{
		"thread_id":"thread-123",
		"note":"Prefers email follow-up.",
		"tags":["preference"]
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected created status, got %d", recorder.Code)
	}

	if service.contactID != "contact-123" {
		t.Fatalf("expected contact id from path, got %q", service.contactID)
	}

	if service.input.Note != "Prefers email follow-up." {
		t.Fatalf("expected note from request, got %#v", service.input)
	}

	var response contactMemoryResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.ID != "memory-123" {
		t.Fatalf("expected created memory id in response, got %#v", response)
	}

	if response.ThreadID != "thread-123" {
		t.Fatalf("expected thread id in response, got %#v", response)
	}
}

func TestContactMemoryHandlerRejectsInvalidPayload(t *testing.T) {
	handler := NewContactMemoryHandler(&contactMemoryServiceStub{})
	request := httptest.NewRequest(http.MethodPost, "/contacts/contact-123/memory", bytes.NewBufferString(`{`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestContactMemoryHandlerReturnsNotFoundForMissingContact(t *testing.T) {
	handler := NewContactMemoryHandler(&contactMemoryServiceStub{err: domain.ErrNotFound})
	request := httptest.NewRequest(http.MethodPost, "/contacts/contact-123/memory", bytes.NewBufferString(`{
		"thread_id":"thread-123",
		"note":"Prefers email follow-up."
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewContactMemoryHandler(&contactMemoryServiceStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodPost, "/contacts/contact-123/memory", bytes.NewBufferString(`{
		"thread_id":"thread-123",
		"note":"Prefers email follow-up."
	}`))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error status, got %d", recorder.Code)
	}
}

type contactMemoryServiceStub struct {
	contactID string
	input     CreateContactMemoryInput
	entry     domain.ContactMemoryEntry
	err       error
}

func (s *contactMemoryServiceStub) CreateContactMemory(_ context.Context, contactID string, input CreateContactMemoryInput) (domain.ContactMemoryEntry, error) {
	s.contactID = contactID
	s.input = input
	return s.entry, s.err
}

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

func TestContactHandlerGetsContactByID(t *testing.T) {
	service := &contactServiceStub{
		contact: domain.Contact{
			ID:             "contact-123",
			OrganizationID: "org-123",
			EmailAddress:   "sender@example.com",
			DisplayName:    "Sender Example",
			LastSeenAt:     time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC),
			CreatedAt:      time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC),
		},
	}

	handler := NewContactHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/contacts/contact-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	if service.contactID != "contact-123" {
		t.Fatalf("expected contact id from path, got %q", service.contactID)
	}

	var response contactResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.ID != "contact-123" {
		t.Fatalf("expected contact id in response, got %#v", response)
	}

	if response.EmailAddress != "sender@example.com" {
		t.Fatalf("expected email in response, got %#v", response)
	}
}

func TestContactHandlerRejectsInvalidPath(t *testing.T) {
	handler := NewContactHandler(&contactServiceStub{})
	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestContactHandlerReturnsNotFoundForMissingContact(t *testing.T) {
	handler := NewContactHandler(&contactServiceStub{err: domain.ErrNotFound})
	request := httptest.NewRequest(http.MethodGet, "/contacts/contact-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewContactHandler(&contactServiceStub{err: errors.New("boom")})
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error status, got %d", recorder.Code)
	}
}

type contactServiceStub struct {
	contactID string
	contact   domain.Contact
	err       error
}

func (s *contactServiceStub) GetContact(_ context.Context, contactID string) (domain.Contact, error) {
	s.contactID = contactID
	return s.contact, s.err
}

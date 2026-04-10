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

func TestSuppressionHandlerReturnsSuppression(t *testing.T) {
	handler := NewSuppressionHandler(&suppressionServiceStub{
		record: domain.SuppressedAddress{
			ID:             "suppression-123",
			OrganizationID: "org-123",
			EmailAddress:   "sender@example.com",
			Reason:         "complaint",
			Source:         "ses",
			UpdatedAt:      time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/suppressions/suppression-123", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", recorder.Code)
	}

	var response suppressionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}
	if response.ID != "suppression-123" || response.Reason != "complaint" {
		t.Fatalf("expected suppression response, got %#v", response)
	}
}

func TestSuppressionHandlerReturnsNotFound(t *testing.T) {
	handler := NewSuppressionHandler(&suppressionServiceStub{err: domain.ErrNotFound})
	request := httptest.NewRequest(http.MethodGet, "/suppressions/suppression-123", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found response, got %d", recorder.Code)
	}

	handler = NewSuppressionHandler(&suppressionServiceStub{err: errors.New("boom")})
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error response, got %d", recorder.Code)
	}
}

type suppressionServiceStub struct {
	record domain.SuppressedAddress
	err    error
}

func (s *suppressionServiceStub) GetSuppression(context.Context, string) (domain.SuppressedAddress, error) {
	return s.record, s.err
}

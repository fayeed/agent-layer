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

func TestSuppressionsHandlerReturnsSuppressions(t *testing.T) {
	handler := NewSuppressionsHandler(&suppressionsServiceStub{
		records: []domain.SuppressedAddress{
			{
				ID:             "suppression-123",
				OrganizationID: "org-123",
				EmailAddress:   "sender@example.com",
				Reason:         "hard_bounce",
				Source:         "ses",
				UpdatedAt:      time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC),
			},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/suppressions?limit=5", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", recorder.Code)
	}

	var response []suppressionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}
	if len(response) != 1 || response[0].ID != "suppression-123" {
		t.Fatalf("expected suppression list response, got %#v", response)
	}
}

type suppressionsServiceStub struct {
	records []domain.SuppressedAddress
	err     error
}

func (s *suppressionsServiceStub) ListSuppressions(context.Context, int) ([]domain.SuppressedAddress, error) {
	return s.records, s.err
}

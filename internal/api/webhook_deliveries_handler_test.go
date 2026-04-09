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

func TestWebhookDeliveriesHandlerReturnsDeliveries(t *testing.T) {
	service := &webhookDeliveriesServiceStub{
		deliveries: []domain.WebhookDelivery{
			{ID: "delivery-123", EventID: "event-123", NextAttemptAt: time.Date(2026, 4, 9, 20, 0, 0, 0, time.UTC)},
			{ID: "delivery-456", EventID: "event-456"},
		},
	}
	handler := NewWebhookDeliveriesHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", recorder.Code)
	}

	var response []webhookDeliveryResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}

	if len(response) != 2 || response[0].ID != "delivery-123" {
		t.Fatalf("expected webhook deliveries response, got %#v", response)
	}
	if response[0].NextAttemptAt == "" {
		t.Fatalf("expected next attempt timestamp in response, got %#v", response[0])
	}

	if service.limit != 0 {
		t.Fatalf("expected default handler limit to pass through as zero, got %d", service.limit)
	}
}

func TestWebhookDeliveriesHandlerPassesLimitQuery(t *testing.T) {
	service := &webhookDeliveriesServiceStub{}
	handler := NewWebhookDeliveriesHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries?limit=5", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", recorder.Code)
	}

	if service.limit != 5 {
		t.Fatalf("expected query limit to be forwarded, got %d", service.limit)
	}
}

func TestWebhookDeliveriesHandlerRejectsInvalidLimit(t *testing.T) {
	handler := NewWebhookDeliveriesHandler(&webhookDeliveriesServiceStub{})

	request := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries?limit=nope", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit to return 400, got %d", recorder.Code)
	}
}

type webhookDeliveriesServiceStub struct {
	deliveries []domain.WebhookDelivery
	limit      int
	err        error
}

func (s *webhookDeliveriesServiceStub) ListWebhookDeliveries(_ context.Context, limit int) ([]domain.WebhookDelivery, error) {
	s.limit = limit
	return s.deliveries, s.err
}

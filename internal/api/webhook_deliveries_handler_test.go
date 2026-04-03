package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestWebhookDeliveriesHandlerReturnsDeliveries(t *testing.T) {
	handler := NewWebhookDeliveriesHandler(&webhookDeliveriesServiceStub{
		deliveries: []domain.WebhookDelivery{
			{ID: "delivery-123", EventID: "event-123"},
			{ID: "delivery-456", EventID: "event-456"},
		},
	})

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
}

type webhookDeliveriesServiceStub struct {
	deliveries []domain.WebhookDelivery
	err        error
}

func (s *webhookDeliveriesServiceStub) ListWebhookDeliveries(context.Context) ([]domain.WebhookDelivery, error) {
	return s.deliveries, s.err
}

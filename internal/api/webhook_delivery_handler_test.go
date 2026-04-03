package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestWebhookDeliveryHandlerReturnsDelivery(t *testing.T) {
	handler := NewWebhookDeliveryHandler(&webhookDeliveryServiceStub{
		delivery: domain.WebhookDelivery{
			ID:           "delivery-123",
			EventID:      "event-123",
			EventType:    "message.received",
			Status:       "failed",
			AttemptCount: 2,
			ResponseCode: 500,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries/delivery-123", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", recorder.Code)
	}

	var response webhookDeliveryResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}

	if response.ID != "delivery-123" || response.ResponseCode != 500 {
		t.Fatalf("expected webhook delivery response, got %#v", response)
	}
}

type webhookDeliveryServiceStub struct {
	delivery domain.WebhookDelivery
	err      error
}

func (s *webhookDeliveryServiceStub) GetWebhookDelivery(context.Context, string) (domain.WebhookDelivery, error) {
	return s.delivery, s.err
}

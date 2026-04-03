package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestWebhookReplayHandlerReplaysDelivery(t *testing.T) {
	handler := NewWebhookReplayHandler(&webhookReplayServiceStub{
		delivery: domain.WebhookDelivery{
			ID:           "delivery-123",
			EventID:      "event-123",
			EventType:    "message.received",
			Status:       "succeeded",
			AttemptCount: 2,
			ResponseCode: 202,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/webhooks/deliveries/delivery-123/replay", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", recorder.Code)
	}

	var response webhookDeliveryResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}

	if response.ID != "delivery-123" || response.AttemptCount != 2 {
		t.Fatalf("expected replayed delivery response, got %#v", response)
	}
}

type webhookReplayServiceStub struct {
	delivery domain.WebhookDelivery
	err      error
}

func (s *webhookReplayServiceStub) ReplayDelivery(context.Context, string) (domain.WebhookDelivery, error) {
	return s.delivery, s.err
}

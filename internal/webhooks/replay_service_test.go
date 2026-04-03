package webhooks

import (
	"context"
	"errors"
	"testing"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestReplayServiceReplaysStoredWebhookRequest(t *testing.T) {
	recorder := &attemptRecorderStub{
		result: domain.WebhookDelivery{
			ID:           "delivery-123",
			EventID:      "event-123",
			Status:       DeliveryStatusSucceeded,
			AttemptCount: 2,
		},
	}
	dispatcher := &capturingDispatcher{
		result: core.WebhookDispatchResult{
			StatusCode: 202,
			Body:       []byte(`{"ok":true}`),
		},
	}

	service := NewReplayService(
		replayDeliveryGetterStub{
			delivery: domain.WebhookDelivery{
				ID:             "delivery-123",
				EventID:        "event-123",
				RequestURL:     "https://example.com/webhook",
				RequestPayload: []byte(`{"event":"message.received"}`),
				RequestHeaders: map[string]string{"Content-Type": "application/json"},
			},
		},
		dispatcher,
		recorder,
	)

	result, err := service.ReplayDelivery(context.Background(), "delivery-123")
	if err != nil {
		t.Fatalf("expected replay to succeed, got error: %v", err)
	}

	if dispatcher.input.URL != "https://example.com/webhook" {
		t.Fatalf("expected replay url to be reused, got %#v", dispatcher.input)
	}

	if result.Delivery.AttemptCount != 2 {
		t.Fatalf("expected recorded replay result, got %#v", result.Delivery)
	}
}

func TestReplayServiceRequiresStoredRequestSnapshot(t *testing.T) {
	service := NewReplayService(
		replayDeliveryGetterStub{
			delivery: domain.WebhookDelivery{ID: "delivery-123"},
		},
		dispatcherStub{},
		&attemptRecorderStub{},
	)

	_, err := service.ReplayDelivery(context.Background(), "delivery-123")
	if !errors.Is(err, ErrReplayRequestUnavailable) {
		t.Fatalf("expected missing request snapshot error, got %v", err)
	}
}

type replayDeliveryGetterStub struct {
	delivery domain.WebhookDelivery
	err      error
}

func (s replayDeliveryGetterStub) GetWebhookDeliveryByID(context.Context, string) (domain.WebhookDelivery, error) {
	return s.delivery, s.err
}

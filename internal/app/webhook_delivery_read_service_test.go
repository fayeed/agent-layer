package app

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestWebhookDeliveryReadServiceLoadsDeliveryByID(t *testing.T) {
	store := &webhookDeliveryGetterStub{
		delivery: domain.WebhookDelivery{
			ID:           "delivery-123",
			EventID:      "event-123",
			EventType:    "message.received",
			Status:       "failed",
			AttemptCount: 2,
		},
	}

	service := NewWebhookDeliveryReadService(store)
	delivery, err := service.GetWebhookDelivery(context.Background(), "delivery-123")
	if err != nil {
		t.Fatalf("expected webhook delivery read to succeed, got error: %v", err)
	}

	if delivery.ID != "delivery-123" || delivery.AttemptCount != 2 {
		t.Fatalf("expected loaded webhook delivery, got %#v", delivery)
	}
}

type webhookDeliveryGetterStub struct {
	delivery domain.WebhookDelivery
	err      error
}

func (s *webhookDeliveryGetterStub) GetWebhookDeliveryByID(context.Context, string) (domain.WebhookDelivery, error) {
	return s.delivery, s.err
}

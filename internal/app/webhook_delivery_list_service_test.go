package app

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestWebhookDeliveryListServiceLoadsRecentDeliveries(t *testing.T) {
	store := &webhookDeliveryListerStub{
		deliveries: []domain.WebhookDelivery{
			{ID: "delivery-123"},
			{ID: "delivery-456"},
		},
	}

	service := NewWebhookDeliveryListService(store, 10)
	deliveries, err := service.ListWebhookDeliveries(context.Background())
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 2 || deliveries[0].ID != "delivery-123" {
		t.Fatalf("expected loaded webhook deliveries, got %#v", deliveries)
	}

	if store.limit != 10 {
		t.Fatalf("expected configured limit to be passed through, got %d", store.limit)
	}
}

type webhookDeliveryListerStub struct {
	deliveries []domain.WebhookDelivery
	limit      int
	err        error
}

func (s *webhookDeliveryListerStub) ListWebhookDeliveries(_ context.Context, limit int) ([]domain.WebhookDelivery, error) {
	s.limit = limit
	return s.deliveries, s.err
}

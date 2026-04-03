package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type WebhookDeliveryLister interface {
	ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)
}

type WebhookDeliveryListService struct {
	deliveries WebhookDeliveryLister
	limit      int
}

func NewWebhookDeliveryListService(deliveries WebhookDeliveryLister, limit int) WebhookDeliveryListService {
	if limit <= 0 {
		limit = 20
	}
	return WebhookDeliveryListService{
		deliveries: deliveries,
		limit:      limit,
	}
}

func (s WebhookDeliveryListService) ListWebhookDeliveries(ctx context.Context) ([]domain.WebhookDelivery, error) {
	return s.deliveries.ListWebhookDeliveries(ctx, s.limit)
}

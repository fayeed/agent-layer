package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type WebhookDeliveryGetter interface {
	GetWebhookDeliveryByID(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error)
}

type WebhookDeliveryReadService struct {
	deliveries WebhookDeliveryGetter
}

func NewWebhookDeliveryReadService(deliveries WebhookDeliveryGetter) WebhookDeliveryReadService {
	return WebhookDeliveryReadService{deliveries: deliveries}
}

func (s WebhookDeliveryReadService) GetWebhookDelivery(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	return s.deliveries.GetWebhookDeliveryByID(ctx, deliveryID)
}

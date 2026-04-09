package webhooks

import (
	"context"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type DeliveryLister interface {
	ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)
}

type DeliveryReplayer interface {
	ReplayDelivery(ctx context.Context, deliveryID string) (DeliveryResult, error)
}

type RetrySweepResult struct {
	Attempted  int
	Succeeded  int
	Failed     int
	Skipped    int
	Deliveries []domain.WebhookDelivery
}

type RetrySweepService struct {
	deliveries DeliveryLister
	replayer   DeliveryReplayer
	now        func() time.Time
}

func NewRetrySweepService(deliveries DeliveryLister, replayer DeliveryReplayer, now func() time.Time) RetrySweepService {
	if now == nil {
		now = time.Now
	}
	return RetrySweepService{
		deliveries: deliveries,
		replayer:   replayer,
		now:        now,
	}
}

func (s RetrySweepService) RetryDueDeliveries(ctx context.Context, limit int) (RetrySweepResult, error) {
	list, err := s.deliveries.ListWebhookDeliveries(ctx, limit)
	if err != nil {
		return RetrySweepResult{}, err
	}

	result := RetrySweepResult{}
	at := s.now()
	for _, delivery := range list {
		if !isRetryDue(delivery, at) {
			result.Skipped++
			continue
		}

		result.Attempted++
		replayed, err := s.replayer.ReplayDelivery(ctx, delivery.ID)
		if err != nil {
			result.Failed++
			continue
		}

		result.Succeeded++
		result.Deliveries = append(result.Deliveries, replayed.Delivery)
	}

	return result, nil
}

func isRetryDue(delivery domain.WebhookDelivery, at time.Time) bool {
	if delivery.Status != DeliveryStatusRetrying {
		return false
	}
	if delivery.NextAttemptAt.IsZero() {
		return false
	}
	return !delivery.NextAttemptAt.After(at)
}

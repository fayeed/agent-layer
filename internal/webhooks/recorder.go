package webhooks

import (
	"context"
	"net/http"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

const (
	DeliveryStatusSucceeded  = "succeeded"
	DeliveryStatusRetrying   = "retrying"
	DeliveryStatusDeadLetter = "dead_letter"
)

type DeliveryRepository interface {
	Save(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error)
}

type RecordAttemptInput struct {
	Delivery domain.WebhookDelivery
	Response core.WebhookDispatchResult
}

type Recorder struct {
	repository  DeliveryRepository
	maxAttempts int
	backoff     func(attempt int) time.Duration
}

func NewRecorder(repository DeliveryRepository) Recorder {
	return Recorder{
		repository:  repository,
		maxAttempts: 3,
		backoff:     defaultBackoff,
	}
}

func (r Recorder) RecordAttempt(ctx context.Context, input RecordAttemptInput) (domain.WebhookDelivery, error) {
	delivery := input.Delivery
	delivery.AttemptCount++
	delivery.ResponseCode = input.Response.StatusCode
	delivery.ResponseBody = input.Response.Body
	delivery.LastAttemptAt = input.Response.DeliveredAt
	delivery.UpdatedAt = input.Response.DeliveredAt

	if input.Response.StatusCode >= http.StatusOK && input.Response.StatusCode < http.StatusMultipleChoices {
		delivery.Status = DeliveryStatusSucceeded
		delivery.NextAttemptAt = time.Time{}
	} else {
		if delivery.AttemptCount <= r.maxAttempts {
			delivery.Status = DeliveryStatusRetrying
			delivery.NextAttemptAt = input.Response.DeliveredAt.Add(r.backoff(delivery.AttemptCount))
		} else {
			delivery.Status = DeliveryStatusDeadLetter
			delivery.NextAttemptAt = time.Time{}
		}
	}

	return r.repository.Save(ctx, delivery)
}

func defaultBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Minute
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	return delay
}

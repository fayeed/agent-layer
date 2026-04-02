package webhooks

import (
	"context"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

const (
	DeliveryStatusSucceeded = "succeeded"
	DeliveryStatusFailed    = "failed"
)

type DeliveryRepository interface {
	Save(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error)
}

type RecordAttemptInput struct {
	Delivery domain.WebhookDelivery
	Response core.WebhookDispatchResult
}

type Recorder struct {
	repository DeliveryRepository
}

func NewRecorder(repository DeliveryRepository) Recorder {
	return Recorder{repository: repository}
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
	} else {
		delivery.Status = DeliveryStatusFailed
	}

	return r.repository.Save(ctx, delivery)
}

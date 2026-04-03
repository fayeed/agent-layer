package webhooks

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type MessageReceivedDeliveryInterface interface {
	DeliverMessageReceived(ctx context.Context, input DeliverMessageReceivedInput) (DeliverMessageReceivedResult, error)
}

type AttemptRecorderInterface interface {
	RecordAttempt(ctx context.Context, input RecordAttemptInput) (domain.WebhookDelivery, error)
}

type DeliveryResult struct {
	Request  core.WebhookDispatchRequest
	Response core.WebhookDispatchResult
	Delivery domain.WebhookDelivery
}

type DeliveryService struct {
	base     MessageReceivedDeliveryInterface
	recorder AttemptRecorderInterface
}

func NewDeliveryService(base MessageReceivedDeliveryInterface, recorder AttemptRecorderInterface) DeliveryService {
	return DeliveryService{
		base:     base,
		recorder: recorder,
	}
}

func (s DeliveryService) DeliverAndRecordMessageReceived(ctx context.Context, input DeliverMessageReceivedInput) (DeliveryResult, error) {
	delivered, err := s.base.DeliverMessageReceived(ctx, input)
	if err != nil {
		return DeliveryResult{}, err
	}

	delivery := delivered.Request.Delivery
	delivery.RequestURL = input.URL
	delivery.RequestPayload = append([]byte(nil), delivered.Request.Payload...)
	delivery.RequestHeaders = copyHeaders(delivered.Request.Headers)

	recorded, err := s.recorder.RecordAttempt(ctx, RecordAttemptInput{
		Delivery: delivery,
		Response: delivered.Response,
	})
	if err != nil {
		return DeliveryResult{}, err
	}

	return DeliveryResult{
		Request:  delivered.Request,
		Response: delivered.Response,
		Delivery: recorded,
	}, nil
}

func copyHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(headers))
	for key, value := range headers {
		out[key] = value
	}
	return out
}

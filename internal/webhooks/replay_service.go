package webhooks

import (
	"context"
	"errors"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

var ErrReplayRequestUnavailable = errors.New("webhook replay request unavailable")

type DeliveryGetter interface {
	GetWebhookDeliveryByID(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error)
}

type ReplayService struct {
	deliveries DeliveryGetter
	dispatcher DispatcherInterface
	recorder   AttemptRecorderInterface
}

func NewReplayService(deliveries DeliveryGetter, dispatcher DispatcherInterface, recorder AttemptRecorderInterface) ReplayService {
	return ReplayService{
		deliveries: deliveries,
		dispatcher: dispatcher,
		recorder:   recorder,
	}
}

func (s ReplayService) ReplayDelivery(ctx context.Context, deliveryID string) (DeliveryResult, error) {
	delivery, err := s.deliveries.GetWebhookDeliveryByID(ctx, deliveryID)
	if err != nil {
		return DeliveryResult{}, err
	}

	if delivery.RequestURL == "" || len(delivery.RequestPayload) == 0 {
		return DeliveryResult{}, ErrReplayRequestUnavailable
	}

	request := core.WebhookDispatchRequest{
		Delivery: delivery,
		Payload:  append([]byte(nil), delivery.RequestPayload...),
		Headers:  copyHeaders(delivery.RequestHeaders),
	}

	response, err := s.dispatcher.Dispatch(ctx, DispatchInput{
		URL:     delivery.RequestURL,
		Request: request,
	})
	if err != nil {
		return DeliveryResult{}, err
	}

	recorded, err := s.recorder.RecordAttempt(ctx, RecordAttemptInput{
		Delivery: delivery,
		Response: response,
	})
	if err != nil {
		return DeliveryResult{}, err
	}

	return DeliveryResult{
		Request:  request,
		Response: response,
		Delivery: recorded,
	}, nil
}

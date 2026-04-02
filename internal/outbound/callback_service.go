package outbound

import (
	"context"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
)

var ErrProviderMessageNotFound = errors.New("provider message not found")

type ProviderMessageLookup interface {
	FindByProviderMessageID(ctx context.Context, providerMessageID string) (domain.Message, bool, error)
}

type DeliveryStatusRecorder interface {
	RecordStatus(ctx context.Context, input RecordDeliveryStatusInput) (domain.Message, error)
}

type CallbackService struct {
	lookup   ProviderMessageLookup
	recorder DeliveryStatusRecorder
}

func NewCallbackService(lookup ProviderMessageLookup, recorder DeliveryStatusRecorder) CallbackService {
	return CallbackService{
		lookup:   lookup,
		recorder: recorder,
	}
}

func (s CallbackService) ApplyCallback(ctx context.Context, event DeliveryCallbackEvent) (domain.Message, error) {
	message, found, err := s.lookup.FindByProviderMessageID(ctx, event.ProviderMessageID)
	if err != nil {
		return domain.Message{}, err
	}
	if !found {
		return domain.Message{}, ErrProviderMessageNotFound
	}

	return s.recorder.RecordStatus(ctx, RecordDeliveryStatusInput{
		Message:    message,
		Status:     event.Status,
		OccurredAt: event.OccurredAt,
	})
}

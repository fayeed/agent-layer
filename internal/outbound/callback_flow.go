package outbound

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type CallbackApplier interface {
	ApplyCallback(ctx context.Context, event DeliveryCallbackEvent) (domain.Message, error)
}

type SuppressionApplier interface {
	Apply(ctx context.Context, input SuppressionInput) (domain.SuppressedAddress, bool, error)
}

type CallbackFlowInput struct {
	Event   DeliveryCallbackEvent
	Contact domain.Contact
}

type CallbackFlowResult struct {
	Message     domain.Message
	Suppression domain.SuppressedAddress
	Suppressed  bool
}

type CallbackFlow struct {
	callbacks    CallbackApplier
	suppressions SuppressionApplier
}

func NewCallbackFlow(callbacks CallbackApplier, suppressions SuppressionApplier) CallbackFlow {
	return CallbackFlow{
		callbacks:    callbacks,
		suppressions: suppressions,
	}
}

func (f CallbackFlow) Apply(ctx context.Context, input CallbackFlowInput) (CallbackFlowResult, error) {
	message, err := f.callbacks.ApplyCallback(ctx, input.Event)
	if err != nil {
		return CallbackFlowResult{}, err
	}

	suppression, changed, err := f.suppressions.Apply(ctx, SuppressionInput{
		Message:    message,
		Contact:    input.Contact,
		Status:     input.Event.Status,
		OccurredAt: input.Event.OccurredAt,
	})
	if err != nil {
		return CallbackFlowResult{}, err
	}

	return CallbackFlowResult{
		Message:     message,
		Suppression: suppression,
		Suppressed:  changed,
	}, nil
}

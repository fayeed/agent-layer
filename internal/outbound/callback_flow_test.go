package outbound

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestCallbackFlowAppliesStatusAndSuppression(t *testing.T) {
	callbacks := &callbackApplierStub{
		message: domain.Message{
			ID:             "message-123",
			OrganizationID: "org-123",
			ContactID:      "contact-123",
			DeliveryState:  DeliveryStateHardBounce,
		},
	}
	suppressions := &suppressionApplierStub{
		record: domain.SuppressedAddress{
			ID:             "suppression-123",
			OrganizationID: "org-123",
			EmailAddress:   "sender@example.com",
			Reason:         DeliveryStateHardBounce,
		},
		changed: true,
	}

	flow := NewCallbackFlow(callbacks, suppressions)

	result, err := flow.Apply(context.Background(), CallbackFlowInput{
		Event: DeliveryCallbackEvent{
			ProviderMessageID: "ses-123",
			Status:            DeliveryStateHardBounce,
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
	})
	if err != nil {
		t.Fatalf("expected callback flow to succeed, got error: %v", err)
	}

	if callbacks.event.ProviderMessageID != "ses-123" {
		t.Fatalf("expected callback event to be passed through, got %#v", callbacks.event)
	}

	if !result.Suppressed {
		t.Fatal("expected hard bounce to trigger suppression")
	}

	if result.Suppression.ID != "suppression-123" {
		t.Fatalf("expected suppression result, got %#v", result)
	}
}

func TestCallbackFlowSkipsSuppressionWhenNotNeeded(t *testing.T) {
	flow := NewCallbackFlow(
		&callbackApplierStub{
			message: domain.Message{
				ID:            "message-123",
				DeliveryState: DeliveryStateDelivered,
			},
		},
		&suppressionApplierStub{},
	)

	result, err := flow.Apply(context.Background(), CallbackFlowInput{
		Event: DeliveryCallbackEvent{
			ProviderMessageID: "ses-123",
			Status:            DeliveryStateDelivered,
		},
		Contact: domain.Contact{
			EmailAddress: "sender@example.com",
		},
	})
	if err != nil {
		t.Fatalf("expected callback flow to succeed, got error: %v", err)
	}

	if result.Suppressed {
		t.Fatal("expected delivered status to avoid suppression")
	}
}

type callbackApplierStub struct {
	event   DeliveryCallbackEvent
	message domain.Message
	err     error
}

func (s *callbackApplierStub) ApplyCallback(_ context.Context, event DeliveryCallbackEvent) (domain.Message, error) {
	s.event = event
	return s.message, s.err
}

type suppressionApplierStub struct {
	input   SuppressionInput
	record  domain.SuppressedAddress
	changed bool
	err     error
}

func (s *suppressionApplierStub) Apply(_ context.Context, input SuppressionInput) (domain.SuppressedAddress, bool, error) {
	s.input = input
	return s.record, s.changed, s.err
}

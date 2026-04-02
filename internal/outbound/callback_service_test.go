package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestCallbackServiceFindsMessageAndAppliesStatus(t *testing.T) {
	lookup := &providerMessageLookupStub{
		message: domain.Message{
			ID:                "message-123",
			OrganizationID:    "org-123",
			ContactID:         "contact-123",
			ProviderMessageID: "ses-123",
			DeliveryState:     DeliveryStateSent,
		},
		found: true,
	}
	recorder := &deliveryStatusRecorderStub{
		message: domain.Message{
			ID:                "message-123",
			ProviderMessageID: "ses-123",
			DeliveryState:     DeliveryStateDelivered,
		},
	}

	service := NewCallbackService(lookup, recorder)

	result, err := service.ApplyCallback(context.Background(), DeliveryCallbackEvent{
		ProviderMessageID: "ses-123",
		Status:            DeliveryStateDelivered,
		OccurredAt:        time.Date(2026, 4, 3, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected callback application to succeed, got error: %v", err)
	}

	if lookup.providerMessageID != "ses-123" {
		t.Fatalf("expected lookup by provider message id, got %q", lookup.providerMessageID)
	}

	if recorder.input.Status != DeliveryStateDelivered {
		t.Fatalf("expected recorder to receive delivery status, got %#v", recorder.input)
	}

	if result.DeliveryState != DeliveryStateDelivered {
		t.Fatalf("expected updated message to be returned, got %#v", result)
	}
}

func TestCallbackServiceReturnsNotFoundForUnknownProviderMessage(t *testing.T) {
	service := NewCallbackService(&providerMessageLookupStub{}, &deliveryStatusRecorderStub{})

	_, err := service.ApplyCallback(context.Background(), DeliveryCallbackEvent{
		ProviderMessageID: "unknown-provider-id",
		Status:            DeliveryStateDelivered,
		OccurredAt:        time.Date(2026, 4, 3, 13, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected unknown provider message id to fail")
	}
}

type providerMessageLookupStub struct {
	providerMessageID string
	message           domain.Message
	found             bool
	err               error
}

func (s *providerMessageLookupStub) FindByProviderMessageID(_ context.Context, providerMessageID string) (domain.Message, bool, error) {
	s.providerMessageID = providerMessageID
	return s.message, s.found, s.err
}

type deliveryStatusRecorderStub struct {
	input   RecordDeliveryStatusInput
	message domain.Message
	err     error
}

func (s *deliveryStatusRecorderStub) RecordStatus(_ context.Context, input RecordDeliveryStatusInput) (domain.Message, error) {
	s.input = input
	return s.message, s.err
}

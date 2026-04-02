package webhooks

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestDeliveryServiceDeliversAndRecordsAttempt(t *testing.T) {
	base := deliveryServiceStub{
		result: DeliverMessageReceivedResult{
			Request: core.WebhookDispatchRequest{
				Delivery: domain.WebhookDelivery{
					ID:        "delivery-123",
					EventID:   "event-123",
					EventType: "message.received",
				},
			},
			Response: core.WebhookDispatchResult{
				StatusCode: 202,
				Body:       []byte(`{"ok":true}`),
			},
		},
	}
	recorder := &attemptRecorderStub{
		result: domain.WebhookDelivery{
			ID:           "delivery-123",
			EventID:      "event-123",
			EventType:    "message.received",
			Status:       DeliveryStatusSucceeded,
			AttemptCount: 1,
		},
	}

	service := NewDeliveryService(base, recorder)

	result, err := service.DeliverAndRecordMessageReceived(context.Background(), DeliverMessageReceivedInput{
		URL:           "https://example.com/webhook",
		WebhookSecret: "super-secret",
	})
	if err != nil {
		t.Fatalf("expected delivery service to succeed, got error: %v", err)
	}

	if result.Response.StatusCode != 202 {
		t.Fatalf("expected dispatch response, got %#v", result.Response)
	}

	if result.Delivery.Status != DeliveryStatusSucceeded {
		t.Fatalf("expected recorded delivery status, got %#v", result.Delivery)
	}
}

func TestDeliveryServicePassesDispatchResultIntoRecorder(t *testing.T) {
	recorder := &attemptRecorderStub{}
	service := NewDeliveryService(
		deliveryServiceStub{
			result: DeliverMessageReceivedResult{
				Request: core.WebhookDispatchRequest{
					Delivery: domain.WebhookDelivery{
						ID:        "delivery-123",
						EventID:   "event-123",
						EventType: "message.received",
					},
				},
				Response: core.WebhookDispatchResult{
					StatusCode: 500,
					Body:       []byte(`{"error":"boom"}`),
				},
			},
		},
		recorder,
	)

	_, err := service.DeliverAndRecordMessageReceived(context.Background(), DeliverMessageReceivedInput{})
	if err != nil {
		t.Fatalf("expected delivery service to succeed, got error: %v", err)
	}

	if recorder.input.Delivery.EventID != "event-123" {
		t.Fatalf("expected recorder to receive request delivery, got %#v", recorder.input)
	}

	if recorder.input.Response.StatusCode != 500 {
		t.Fatalf("expected recorder to receive dispatch response, got %#v", recorder.input.Response)
	}
}

type deliveryServiceStub struct {
	result DeliverMessageReceivedResult
	err    error
}

func (s deliveryServiceStub) DeliverMessageReceived(context.Context, DeliverMessageReceivedInput) (DeliverMessageReceivedResult, error) {
	return s.result, s.err
}

type attemptRecorderStub struct {
	input  RecordAttemptInput
	result domain.WebhookDelivery
	err    error
}

func (s *attemptRecorderStub) RecordAttempt(_ context.Context, input RecordAttemptInput) (domain.WebhookDelivery, error) {
	s.input = input
	return s.result, s.err
}

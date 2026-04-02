package webhooks

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestServiceDeliversMessageReceivedWebhook(t *testing.T) {
	service := NewService(
		builderStub{
			request: core.WebhookDispatchRequest{
				Delivery: domain.WebhookDelivery{
					ID:        "delivery-123",
					EventID:   "event-123",
					EventType: "message.received",
				},
				Payload: []byte(`{"event":"message.received"}`),
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		signerStub{
			request: core.WebhookDispatchRequest{
				Delivery: domain.WebhookDelivery{
					ID:        "delivery-123",
					EventID:   "event-123",
					EventType: "message.received",
				},
				Payload: []byte(`{"event":"message.received"}`),
				Headers: map[string]string{
					"Content-Type":           "application/json",
					HeaderSignature:          "sha256=test",
					HeaderSignatureTimestamp: "2026-04-03T04:00:00Z",
				},
			},
		},
		dispatcherStub{
			result: core.WebhookDispatchResult{
				StatusCode: 202,
				Body:       []byte(`{"ok":true}`),
			},
		},
	)

	result, err := service.DeliverMessageReceived(context.Background(), DeliverMessageReceivedInput{
		URL:           "https://example.com/webhook",
		WebhookSecret: "super-secret",
		BuildInput: BuildMessageReceivedInput{
			Organization: domain.Organization{ID: "org-123"},
			Agent:        domain.Agent{ID: "agent-123"},
			Inbox:        domain.Inbox{ID: "inbox-123"},
			Delivery: domain.WebhookDelivery{
				ID:        "delivery-123",
				EventID:   "event-123",
				EventType: "message.received",
			},
			Handled: inbound.HandleResult{
				Message: domain.Message{ID: "message-123"},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected service to succeed, got error: %v", err)
	}

	if result.Request.Delivery.EventID != "event-123" {
		t.Fatalf("expected signed request to be returned, got %#v", result.Request)
	}

	if result.Response.StatusCode != 202 {
		t.Fatalf("expected dispatch response to be returned, got %#v", result.Response)
	}
}

func TestServicePassesSignedRequestIntoDispatcher(t *testing.T) {
	dispatcher := &capturingDispatcher{}

	service := NewService(
		builderStub{
			request: core.WebhookDispatchRequest{
				Delivery: domain.WebhookDelivery{
					ID:        "delivery-123",
					EventID:   "event-123",
					EventType: "message.received",
				},
				Payload: []byte(`{"event":"message.received"}`),
			},
		},
		signerStub{
			request: core.WebhookDispatchRequest{
				Delivery: domain.WebhookDelivery{
					ID:        "delivery-123",
					EventID:   "event-123",
					EventType: "message.received",
				},
				Payload: []byte(`{"event":"message.received"}`),
				Headers: map[string]string{
					HeaderSignature:          "sha256=test",
					HeaderSignatureTimestamp: "2026-04-03T04:00:00Z",
				},
			},
		},
		dispatcher,
	)

	_, err := service.DeliverMessageReceived(context.Background(), DeliverMessageReceivedInput{
		URL:           "https://example.com/webhook",
		WebhookSecret: "super-secret",
		BuildInput: BuildMessageReceivedInput{
			Delivery: domain.WebhookDelivery{
				ID:        "delivery-123",
				EventID:   "event-123",
				EventType: "message.received",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected service to succeed, got error: %v", err)
	}

	if dispatcher.input.URL != "https://example.com/webhook" {
		t.Fatalf("expected dispatcher url, got %#v", dispatcher.input)
	}

	if dispatcher.input.Request.Headers[HeaderSignature] != "sha256=test" {
		t.Fatalf("expected signed headers to be forwarded, got %#v", dispatcher.input.Request.Headers)
	}
}

type builderStub struct {
	request core.WebhookDispatchRequest
	err     error
}

func (s builderStub) Build(context.Context, BuildMessageReceivedInput) (core.WebhookDispatchRequest, error) {
	return s.request, s.err
}

type signerStub struct {
	request core.WebhookDispatchRequest
	err     error
}

func (s signerStub) Sign(core.WebhookDispatchRequest, string) (core.WebhookDispatchRequest, error) {
	return s.request, s.err
}

type dispatcherStub struct {
	result core.WebhookDispatchResult
	err    error
}

func (s dispatcherStub) Dispatch(context.Context, DispatchInput) (core.WebhookDispatchResult, error) {
	return s.result, s.err
}

type capturingDispatcher struct {
	input  DispatchInput
	result core.WebhookDispatchResult
	err    error
}

func (s *capturingDispatcher) Dispatch(_ context.Context, input DispatchInput) (core.WebhookDispatchResult, error) {
	s.input = input
	return s.result, s.err
}

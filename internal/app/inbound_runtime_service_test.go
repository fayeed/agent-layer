package app

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/webhooks"
)

func TestInboundRuntimeServiceSkipsWebhookWhenURLMissing(t *testing.T) {
	handled := inbound.HandleResult{
		Thread:  domain.Thread{ID: "thread-123"},
		Contact: domain.Contact{ID: "contact-123"},
	}
	inboundHandler := inboundHandlerStub{result: handled}
	deliveries := &messageReceivedDeliveryServiceStub{}

	service := NewInboundRuntimeService(
		inboundHandler,
		threadMessagesGetterStub{},
		contactMemoryListerStub{},
		deliveries,
		func() time.Time { return time.Date(2026, 4, 3, 20, 0, 0, 0, time.UTC) },
		InboundRuntimeConfig{
			Agent: domain.Agent{Status: domain.AgentStatusActive},
		},
	)

	result, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{})
	if err != nil {
		t.Fatalf("expected inbound runtime service to succeed, got error: %v", err)
	}

	if result.Thread.ID != "thread-123" {
		t.Fatalf("expected handled result to be returned, got %#v", result)
	}

	if deliveries.calls != 0 {
		t.Fatalf("expected webhook delivery to be skipped, got %d calls", deliveries.calls)
	}
}

func TestInboundRuntimeServiceSkipsWebhookWhenAgentPaused(t *testing.T) {
	deliveries := &messageReceivedDeliveryServiceStub{}

	service := NewInboundRuntimeService(
		inboundHandlerStub{result: inbound.HandleResult{}},
		threadMessagesGetterStub{},
		contactMemoryListerStub{},
		deliveries,
		time.Now,
		InboundRuntimeConfig{
			WebhookURL: "https://example.com/webhook",
			Agent:      domain.Agent{Status: domain.AgentStatusPaused},
		},
	)

	if _, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{}); err != nil {
		t.Fatalf("expected paused agent handling to still succeed, got error: %v", err)
	}

	if deliveries.calls != 0 {
		t.Fatalf("expected paused agent to skip webhook delivery, got %d calls", deliveries.calls)
	}
}

func TestInboundRuntimeServiceSkipsWebhookForDuplicateInboundMessage(t *testing.T) {
	deliveries := &messageReceivedDeliveryServiceStub{}

	service := NewInboundRuntimeService(
		inboundHandlerStub{result: inbound.HandleResult{
			Duplicate: true,
			Thread:    domain.Thread{ID: "thread-123"},
			Contact:   domain.Contact{ID: "contact-123"},
			Message:   domain.Message{ID: "message-123"},
		}},
		threadMessagesGetterStub{},
		contactMemoryListerStub{},
		deliveries,
		time.Now,
		InboundRuntimeConfig{
			WebhookURL: "https://example.com/webhook",
			Agent:      domain.Agent{Status: domain.AgentStatusActive},
		},
	)

	result, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{})
	if err != nil {
		t.Fatalf("expected duplicate handling to succeed, got error: %v", err)
	}

	if !result.Duplicate {
		t.Fatalf("expected duplicate flag to be preserved, got %#v", result)
	}

	if deliveries.calls != 0 {
		t.Fatalf("expected duplicate inbound message to skip webhook delivery, got %d calls", deliveries.calls)
	}
}

func TestInboundRuntimeServiceDeliversMessageReceivedWebhook(t *testing.T) {
	now := time.Date(2026, 4, 3, 20, 15, 0, 0, time.UTC)
	handled := inbound.HandleResult{
		Thread:  domain.Thread{ID: "thread-123"},
		Contact: domain.Contact{ID: "contact-123"},
		Message: domain.Message{ID: "message-123"},
	}
	deliveries := &messageReceivedDeliveryServiceStub{}

	service := NewInboundRuntimeService(
		inboundHandlerStub{result: handled},
		threadMessagesGetterStub{
			result: []domain.Message{{ID: "message-100", ThreadID: "thread-123"}},
		},
		contactMemoryListerStub{
			result: []domain.ContactMemoryEntry{{ID: "memory-100", ContactID: "contact-123"}},
		},
		deliveries,
		func() time.Time { return now },
		InboundRuntimeConfig{
			Organization: domain.Organization{ID: "org-123", Name: "Acme"},
			Agent: domain.Agent{
				ID:     "agent-123",
				Name:   "Support Agent",
				Status: domain.AgentStatusActive,
			},
			Inbox: domain.Inbox{
				ID:           "inbox-123",
				EmailAddress: "agent@example.com",
			},
			WebhookURL:    "https://example.com/webhook",
			WebhookSecret: "super-secret",
			HistoryLimit:  5,
			MemoryLimit:   3,
		},
	)

	if _, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{}); err != nil {
		t.Fatalf("expected runtime service to deliver webhook, got error: %v", err)
	}

	if deliveries.calls != 1 {
		t.Fatalf("expected one webhook delivery, got %d", deliveries.calls)
	}

	if deliveries.input.URL != "https://example.com/webhook" {
		t.Fatalf("expected webhook url to be forwarded, got %#v", deliveries.input)
	}

	if deliveries.input.BuildInput.Delivery.EventType != "message.received" {
		t.Fatalf("expected message.received delivery, got %#v", deliveries.input.BuildInput.Delivery)
	}

	if deliveries.input.BuildInput.Delivery.OrganizationID != "org-123" {
		t.Fatalf("expected delivery org id, got %#v", deliveries.input.BuildInput.Delivery)
	}

	if len(deliveries.input.BuildInput.ThreadMessages) != 1 {
		t.Fatalf("expected thread history to be included, got %#v", deliveries.input.BuildInput.ThreadMessages)
	}

	if len(deliveries.input.BuildInput.Memory) != 1 {
		t.Fatalf("expected memory entries to be included, got %#v", deliveries.input.BuildInput.Memory)
	}
}

type inboundHandlerStub struct {
	result inbound.HandleResult
	err    error
}

func (s inboundHandlerStub) HandleStoredMessage(context.Context, core.StoredInboundMessage) (inbound.HandleResult, error) {
	return s.result, s.err
}

type threadMessagesGetterStub struct {
	result []domain.Message
	err    error
}

func (s threadMessagesGetterStub) ListByThreadID(context.Context, string, int) ([]domain.Message, error) {
	return s.result, s.err
}

type contactMemoryListerStub struct {
	result []domain.ContactMemoryEntry
	err    error
}

func (s contactMemoryListerStub) ListMemoryByContactID(context.Context, string, int) ([]domain.ContactMemoryEntry, error) {
	return s.result, s.err
}

type messageReceivedDeliveryServiceStub struct {
	input webhooks.DeliverMessageReceivedInput
	calls int
	err   error
}

func (s *messageReceivedDeliveryServiceStub) DeliverAndRecordMessageReceived(_ context.Context, input webhooks.DeliverMessageReceivedInput) (webhooks.DeliveryResult, error) {
	s.calls++
	s.input = input
	return webhooks.DeliveryResult{}, s.err
}

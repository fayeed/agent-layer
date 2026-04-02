package webhooks

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestBuilderBuildsMessageReceivedRequest(t *testing.T) {
	now := time.Date(2026, 4, 3, 1, 0, 0, 0, time.UTC)

	builder := NewMessageReceivedBuilder()
	request, err := builder.Build(context.Background(), BuildMessageReceivedInput{
		Organization: domain.Organization{
			ID:   "org-123",
			Name: "Acme",
		},
		Agent: domain.Agent{
			ID:   "agent-123",
			Name: "Support Agent",
		},
		Inbox: domain.Inbox{
			ID:           "inbox-123",
			EmailAddress: "agent@example.com",
		},
		Delivery: domain.WebhookDelivery{
			ID:        "delivery-123",
			EventID:   "event-123",
			EventType: "message.received",
			CreatedAt: now,
		},
		Handled: inbound.HandleResult{
			ParsedMessage: core.ParsedMessage{
				Subject:           "Re: Hello World",
				SubjectNormalized: "hello world",
			},
			Contact: domain.Contact{
				ID:           "contact-123",
				EmailAddress: "sender@example.com",
				DisplayName:  "Sender Example",
			},
			Thread: domain.Thread{
				ID:                "thread-123",
				SubjectNormalized: "hello world",
				State:             domain.ThreadStateActive,
			},
			Message: domain.Message{
				ID:              "message-123",
				ThreadID:        "thread-123",
				Direction:       domain.MessageDirectionInbound,
				Subject:         "Re: Hello World",
				TextBody:        "Plain body.",
				MessageIDHeader: "<message-123@example.com>",
				CreatedAt:       now,
			},
			ThreadMatchStrategy: "in_reply_to",
			ThreadCreated:       false,
		},
		ThreadMessages: []domain.Message{
			{
				ID:        "message-100",
				ThreadID:  "thread-123",
				Direction: domain.MessageDirectionOutbound,
				Subject:   "Hello World",
				TextBody:  "Previous message.",
				CreatedAt: now.Add(-1 * time.Hour),
			},
		},
		Memory: []domain.ContactMemoryEntry{
			{
				ID:        "memory-123",
				ContactID: "contact-123",
				Note:      "Prefers email follow-up.",
				Tags:      []string{"preference"},
				CreatedAt: now.Add(-2 * time.Hour),
			},
		},
	})
	if err != nil {
		t.Fatalf("expected build to succeed, got error: %v", err)
	}

	if request.Delivery.EventType != "message.received" {
		t.Fatalf("expected delivery event type to be preserved, got %q", request.Delivery.EventType)
	}

	if request.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected json content type header, got %#v", request.Headers)
	}

	var payload MessageReceivedPayload
	if err := json.Unmarshal(request.Payload, &payload); err != nil {
		t.Fatalf("expected payload to be valid json, got error: %v", err)
	}

	if payload.EventID != "event-123" {
		t.Fatalf("expected event id in payload, got %q", payload.EventID)
	}

	if payload.Organization.ID != "org-123" || payload.Agent.ID != "agent-123" || payload.Inbox.ID != "inbox-123" {
		t.Fatalf("expected organization, agent, and inbox identifiers in payload, got %#v", payload)
	}

	if payload.Message.ID != "message-123" {
		t.Fatalf("expected handled message in payload, got %#v", payload.Message)
	}

	if payload.Contact.ID != "contact-123" {
		t.Fatalf("expected contact in payload, got %#v", payload.Contact)
	}

	if payload.Thread.ID != "thread-123" {
		t.Fatalf("expected thread in payload, got %#v", payload.Thread)
	}

	if len(payload.ThreadMessages) != 1 || payload.ThreadMessages[0].ID != "message-100" {
		t.Fatalf("expected thread messages in payload, got %#v", payload.ThreadMessages)
	}

	if len(payload.Memory) != 1 || payload.Memory[0].ID != "memory-123" {
		t.Fatalf("expected memory in payload, got %#v", payload.Memory)
	}
}

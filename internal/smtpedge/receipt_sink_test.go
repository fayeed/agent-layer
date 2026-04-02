package smtpedge

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestReceiptSinkEnqueuesStoredInboundMessage(t *testing.T) {
	handler := &storedMessageHandlerStub{}
	sink := NewReceiptSink(handler)
	receivedAt := time.Date(2026, 4, 3, 16, 0, 0, 0, time.UTC)

	err := sink.Enqueue(context.Background(), inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-session-123",
		OrganizationID:      "org-123",
		AgentID:             "agent-123",
		InboxID:             "inbox-123",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@example.com"},
		RawMessageObjectKey: "raw/inbound-123.eml",
		ReceivedAt:          receivedAt,
	})
	if err != nil {
		t.Fatalf("expected enqueue to succeed, got error: %v", err)
	}

	if handler.message.Receipt.SMTPTransactionID != "smtp-session-123" {
		t.Fatalf("expected smtp transaction id to be forwarded, got %#v", handler.message)
	}

	if handler.message.Receipt.InboxID != "inbox-123" {
		t.Fatalf("expected inbox id to be forwarded, got %#v", handler.message)
	}

	if handler.message.Receipt.RawMessageObjectKey != "raw/inbound-123.eml" {
		t.Fatalf("expected raw object key to be forwarded, got %#v", handler.message)
	}

	if !handler.message.Receipt.ReceivedAt.Equal(receivedAt) {
		t.Fatalf("expected received time %v, got %#v", receivedAt, handler.message)
	}
}

type storedMessageHandlerStub struct {
	message core.StoredInboundMessage
	err     error
}

func (s *storedMessageHandlerStub) HandleStoredMessage(_ context.Context, message core.StoredInboundMessage) (inbound.HandleResult, error) {
	s.message = message
	return inbound.HandleResult{}, s.err
}

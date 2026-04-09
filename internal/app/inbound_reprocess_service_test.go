package app

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestInboundReprocessServiceLoadsStoredReceiptAndRehandlesMessage(t *testing.T) {
	handler := &reprocessInboundHandlerStub{
		result: inbound.HandleResult{Duplicate: true},
	}

	service := NewInboundReprocessService(inboundReceiptGetterStub{
		receipt: inbound.DurableReceiptRequest{
			SMTPTransactionID:   "smtp-session-123",
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			EnvelopeSender:      "sender@example.com",
			EnvelopeRecipients:  []string{"agent@example.com"},
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 9, 14, 0, 0, 0, time.UTC),
		},
	}, handler)

	result, err := service.ReprocessByObjectKey(context.Background(), "raw/test-message.eml")
	if err != nil {
		t.Fatalf("expected reprocess to succeed, got error: %v", err)
	}

	if !result.Duplicate {
		t.Fatalf("expected handler result to be returned, got %#v", result)
	}

	if handler.message.Receipt.RawMessageObjectKey != "raw/test-message.eml" {
		t.Fatalf("expected stored receipt to be forwarded, got %#v", handler.message)
	}
}

type inboundReceiptGetterStub struct {
	receipt inbound.DurableReceiptRequest
	err     error
}

func (s inboundReceiptGetterStub) GetInboundReceiptByObjectKey(context.Context, string) (inbound.DurableReceiptRequest, error) {
	return s.receipt, s.err
}

type reprocessInboundHandlerStub struct {
	message core.StoredInboundMessage
	result  inbound.HandleResult
	err     error
}

func (s *reprocessInboundHandlerStub) HandleStoredMessage(_ context.Context, message core.StoredInboundMessage) (inbound.HandleResult, error) {
	s.message = message
	return s.result, s.err
}

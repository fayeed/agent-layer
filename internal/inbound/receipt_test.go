package inbound

import (
	"testing"
	"time"
)

func TestDurableReceiptRequestCapturesInboundEnvelope(t *testing.T) {
	now := time.Date(2026, 4, 2, 19, 0, 0, 0, time.UTC)

	request := DurableReceiptRequest{
		SMTPTransactionID:   "smtp-session-123",
		OrganizationID:      "org-123",
		AgentID:             "agent-123",
		InboxID:             "inbox-123",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@example.com"},
		RawMessageObjectKey: "raw/2026/04/02/message.eml",
		ReceivedAt:          now,
	}

	if request.SMTPTransactionID != "smtp-session-123" {
		t.Fatalf("expected smtp transaction id to be captured, got %q", request.SMTPTransactionID)
	}

	if request.EnvelopeSender != "sender@example.com" {
		t.Fatalf("expected envelope sender to be captured, got %q", request.EnvelopeSender)
	}

	if len(request.EnvelopeRecipients) != 1 || request.EnvelopeRecipients[0] != "agent@example.com" {
		t.Fatalf("expected envelope recipients to be captured, got %#v", request.EnvelopeRecipients)
	}

	if request.RawMessageObjectKey != "raw/2026/04/02/message.eml" {
		t.Fatalf("expected raw object key to be captured, got %q", request.RawMessageObjectKey)
	}

	if !request.ReceivedAt.Equal(now) {
		t.Fatalf("expected receipt time %v, got %v", now, request.ReceivedAt)
	}
}

func TestDurableReceiptRequestValidate(t *testing.T) {
	valid := DurableReceiptRequest{
		SMTPTransactionID:   "smtp-session-123",
		OrganizationID:      "org-123",
		AgentID:             "agent-123",
		InboxID:             "inbox-123",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@example.com"},
		RawMessageObjectKey: "raw/2026/04/02/message.eml",
		ReceivedAt:          time.Date(2026, 4, 2, 19, 0, 0, 0, time.UTC),
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid durable receipt request, got error: %v", err)
	}

	invalid := valid
	invalid.RawMessageObjectKey = ""

	if err := invalid.Validate(); err == nil {
		t.Fatal("expected missing raw object key to fail validation")
	}
}

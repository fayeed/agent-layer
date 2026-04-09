package smtpedge

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestSessionAcceptsKnownRecipientAndEmitsReceipt(t *testing.T) {
	lookup := &inboxLookupStub{
		inbox: domain.Inbox{
			ID:             "inbox-123",
			OrganizationID: "org-123",
			AgentID:        "agent-123",
			EmailAddress:   "agent@example.com",
		},
		found: true,
	}
	store := &rawMessageStoreStub{}
	sink := &receiptSinkStub{}

	session := NewSession(lookup, store, sink, func() time.Time {
		return time.Date(2026, 4, 3, 15, 0, 0, 0, time.UTC)
	}, func(now time.Time, inbox domain.Inbox) string {
		if !now.Equal(time.Date(2026, 4, 3, 15, 0, 0, 0, time.UTC)) {
			t.Fatalf("expected object key generator to receive receipt time, got %v", now)
		}
		if inbox.ID != "inbox-123" {
			t.Fatalf("expected object key generator to receive accepted inbox, got %#v", inbox)
		}
		return "raw/inbound-123.eml"
	}, "smtp-session-123")

	if err := session.Mail(context.Background(), "sender@example.com"); err != nil {
		t.Fatalf("expected MAIL FROM to succeed, got error: %v", err)
	}

	if err := session.Rcpt(context.Background(), "agent@example.com"); err != nil {
		t.Fatalf("expected RCPT TO to succeed, got error: %v", err)
	}

	if err := session.Data(context.Background(), bytes.NewBufferString("raw mime body")); err != nil {
		t.Fatalf("expected DATA to succeed, got error: %v", err)
	}

	if lookup.emailAddress != "agent@example.com" {
		t.Fatalf("expected inbox lookup by recipient, got %q", lookup.emailAddress)
	}

	if store.objectKey != "raw/inbound-123.eml" {
		t.Fatalf("expected raw message object key, got %q", store.objectKey)
	}

	if string(store.data) != "raw mime body" {
		t.Fatalf("expected raw message bytes to be stored, got %q", string(store.data))
	}

	if sink.receipt.InboxID != "inbox-123" {
		t.Fatalf("expected receipt to reference inbox, got %#v", sink.receipt)
	}

	if sink.receipt.EnvelopeSender != "sender@example.com" {
		t.Fatalf("expected sender in receipt, got %#v", sink.receipt)
	}
}

func TestSessionRejectsUnknownRecipient(t *testing.T) {
	session := NewSession(&inboxLookupStub{}, &rawMessageStoreStub{}, &receiptSinkStub{}, time.Now, func(time.Time, domain.Inbox) string {
		return "raw/unused.eml"
	}, "smtp-session-123")

	if err := session.Rcpt(context.Background(), "missing@example.com"); err == nil {
		t.Fatal("expected unknown recipient to fail")
	}
}

func TestSessionRejectsDataWithoutRecipient(t *testing.T) {
	session := NewSession(&inboxLookupStub{}, &rawMessageStoreStub{}, &receiptSinkStub{}, time.Now, func(time.Time, domain.Inbox) string {
		return "raw/unused.eml"
	}, "smtp-session-123")

	if err := session.Mail(context.Background(), "sender@example.com"); err != nil {
		t.Fatalf("expected MAIL FROM to succeed, got error: %v", err)
	}

	if err := session.Data(context.Background(), bytes.NewBufferString("raw mime body")); err == nil {
		t.Fatal("expected DATA without accepted recipient to fail")
	}
}

func TestSessionUsesDefaultInboxAwareObjectKeyGenerator(t *testing.T) {
	lookup := &inboxLookupStub{
		inbox: domain.Inbox{
			ID:             "Inbox Local/Primary",
			OrganizationID: "org-123",
			AgentID:        "agent-123",
			EmailAddress:   "agent@example.com",
		},
		found: true,
	}
	store := &rawMessageStoreStub{}
	sink := &receiptSinkStub{}
	receivedAt := time.Date(2026, 4, 9, 12, 34, 56, 0, time.UTC)

	session := NewSession(lookup, store, sink, func() time.Time { return receivedAt }, nil, "smtp-session-123")

	if err := session.Mail(context.Background(), "sender@example.com"); err != nil {
		t.Fatalf("expected MAIL FROM to succeed, got error: %v", err)
	}

	if err := session.Rcpt(context.Background(), "agent@example.com"); err != nil {
		t.Fatalf("expected RCPT TO to succeed, got error: %v", err)
	}

	if err := session.Data(context.Background(), bytes.NewBufferString("raw mime body")); err != nil {
		t.Fatalf("expected DATA to succeed, got error: %v", err)
	}

	if store.objectKey == "" {
		t.Fatal("expected default object key generator to produce a key")
	}

	if sink.receipt.RawMessageObjectKey != store.objectKey {
		t.Fatalf("expected stored object key to match receipt, got store=%q receipt=%q", store.objectKey, sink.receipt.RawMessageObjectKey)
	}

	if sink.receipt.ReceivedAt != receivedAt {
		t.Fatalf("expected receipt timestamp to use session clock, got %v", sink.receipt.ReceivedAt)
	}
}

type inboxLookupStub struct {
	emailAddress string
	inbox        domain.Inbox
	found        bool
	err          error
}

func (s *inboxLookupStub) FindByEmailAddress(_ context.Context, emailAddress string) (domain.Inbox, bool, error) {
	s.emailAddress = emailAddress
	return s.inbox, s.found, s.err
}

type rawMessageStoreStub struct {
	objectKey string
	data      []byte
	err       error
}

func (s *rawMessageStoreStub) Put(_ context.Context, objectKey string, data []byte) error {
	s.objectKey = objectKey
	s.data = append([]byte(nil), data...)
	return s.err
}

type receiptSinkStub struct {
	receipt inbound.DurableReceiptRequest
	err     error
}

func (s *receiptSinkStub) Enqueue(_ context.Context, receipt inbound.DurableReceiptRequest) error {
	s.receipt = receipt
	return s.err
}

var _ io.Reader = bytes.NewBuffer(nil)

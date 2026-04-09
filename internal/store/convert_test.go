package store

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestCoreDomainRoundTripsThroughStoreModels(t *testing.T) {
	now := time.Date(2026, 4, 9, 16, 0, 0, 0, time.UTC)

	organization := domain.Organization{
		ID:        "org-123",
		Name:      "Acme Support",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if got := OrganizationFromModel(OrganizationToModel(organization)); got != organization {
		t.Fatalf("expected organization round trip, got %#v", got)
	}

	agent := domain.Agent{
		ID:             "agent-123",
		OrganizationID: "org-123",
		Name:           "Acme Agent",
		Status:         domain.AgentStatusActive,
		WebhookURL:     "https://example.com/webhook",
		WebhookSecret:  "super-secret",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if got := AgentFromModel(AgentToModel(agent)); got != agent {
		t.Fatalf("expected agent round trip, got %#v", got)
	}

	inbox := domain.Inbox{
		ID:               "inbox-123",
		OrganizationID:   "org-123",
		AgentID:          "agent-123",
		EmailAddress:     "agent@example.com",
		Domain:           "example.com",
		DisplayName:      "Acme Inbox",
		OutboundIdentity: "acme-support",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if got := InboxFromModel(InboxToModel(inbox)); got != inbox {
		t.Fatalf("expected inbox round trip, got %#v", got)
	}

	contact := domain.Contact{
		ID:             "contact-123",
		OrganizationID: "org-123",
		EmailAddress:   "sender@example.com",
		DisplayName:    "Sender Example",
		LastSeenAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if got := ContactFromModel(ContactToModel(contact)); got != contact {
		t.Fatalf("expected contact round trip, got %#v", got)
	}

	thread := domain.Thread{
		ID:                "thread-123",
		OrganizationID:    "org-123",
		AgentID:           "agent-123",
		InboxID:           "inbox-123",
		ContactID:         "contact-123",
		SubjectNormalized: "hello world",
		State:             domain.ThreadStateActive,
		LastInboundID:     "message-1",
		LastOutboundID:    "message-2",
		LastActivityAt:    now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if got := ThreadFromModel(ThreadToModel(thread)); got != thread {
		t.Fatalf("expected thread round trip, got %#v", got)
	}

	message := domain.Message{
		ID:                "message-123",
		OrganizationID:    "org-123",
		ThreadID:          "thread-123",
		InboxID:           "inbox-123",
		ContactID:         "contact-123",
		Direction:         domain.MessageDirectionInbound,
		Subject:           "Hello World",
		SubjectNormalized: "hello world",
		MessageIDHeader:   "<message-123@example.com>",
		InReplyTo:         "<message-100@example.com>",
		References:        []string{"<message-1@example.com>", "<message-100@example.com>"},
		TextBody:          "Plain body.",
		HTMLBody:          "<p>HTML body.</p>",
		RawMIMEObjectKey:  "raw/test-message.eml",
		ProviderMessageID: "provider-123",
		DeliveryState:     "sent",
		SentAt:            now,
		DeliveredAt:       now,
		BouncedAt:         now,
		CreatedAt:         now,
	}
	gotMessage := MessageFromModel(MessageToModel(message))
	if !equalStringSlices(gotMessage.References, message.References) {
		t.Fatalf("expected message references round trip, got %#v", gotMessage)
	}
	if gotMessage.ID != message.ID ||
		gotMessage.OrganizationID != message.OrganizationID ||
		gotMessage.ThreadID != message.ThreadID ||
		gotMessage.InboxID != message.InboxID ||
		gotMessage.ContactID != message.ContactID ||
		gotMessage.Direction != message.Direction ||
		gotMessage.Subject != message.Subject ||
		gotMessage.SubjectNormalized != message.SubjectNormalized ||
		gotMessage.MessageIDHeader != message.MessageIDHeader ||
		gotMessage.InReplyTo != message.InReplyTo ||
		gotMessage.TextBody != message.TextBody ||
		gotMessage.HTMLBody != message.HTMLBody ||
		gotMessage.RawMIMEObjectKey != message.RawMIMEObjectKey ||
		gotMessage.ProviderMessageID != message.ProviderMessageID ||
		gotMessage.DeliveryState != message.DeliveryState ||
		!gotMessage.SentAt.Equal(message.SentAt) ||
		!gotMessage.DeliveredAt.Equal(message.DeliveredAt) ||
		!gotMessage.BouncedAt.Equal(message.BouncedAt) ||
		!gotMessage.CreatedAt.Equal(message.CreatedAt) {
		t.Fatalf("expected message round trip, got %#v", gotMessage)
	}
}

func TestAuxiliaryRuntimeDataRoundTripsThroughStoreModels(t *testing.T) {
	now := time.Date(2026, 4, 9, 16, 30, 0, 0, time.UTC)

	memoryEntry := domain.ContactMemoryEntry{
		ID:        "memory-123",
		ContactID: "contact-123",
		ThreadID:  "thread-123",
		Note:      "Prefers email.",
		Tags:      []string{"preference", "email"},
		CreatedAt: now,
	}
	gotMemory := ContactMemoryFromModel(ContactMemoryToModel(memoryEntry, "org-123"))
	if !equalStringSlices(gotMemory.Tags, memoryEntry.Tags) {
		t.Fatalf("expected contact memory tags round trip, got %#v", gotMemory)
	}
	if gotMemory.ID != memoryEntry.ID ||
		gotMemory.ContactID != memoryEntry.ContactID ||
		gotMemory.ThreadID != memoryEntry.ThreadID ||
		gotMemory.Note != memoryEntry.Note ||
		!gotMemory.CreatedAt.Equal(memoryEntry.CreatedAt) {
		t.Fatalf("expected contact memory round trip, got %#v", gotMemory)
	}

	receipt := inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-session-123",
		OrganizationID:      "org-123",
		AgentID:             "agent-123",
		InboxID:             "inbox-123",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@example.com"},
		RawMessageObjectKey: "raw/test-message.eml",
		ReceivedAt:          now,
	}
	gotReceipt := InboundReceiptFromModel(InboundReceiptToModel(receipt))
	if !equalStringSlices(gotReceipt.EnvelopeRecipients, receipt.EnvelopeRecipients) {
		t.Fatalf("expected inbound receipt recipients round trip, got %#v", gotReceipt)
	}
	if gotReceipt.SMTPTransactionID != receipt.SMTPTransactionID ||
		gotReceipt.OrganizationID != receipt.OrganizationID ||
		gotReceipt.AgentID != receipt.AgentID ||
		gotReceipt.InboxID != receipt.InboxID ||
		gotReceipt.EnvelopeSender != receipt.EnvelopeSender ||
		gotReceipt.RawMessageObjectKey != receipt.RawMessageObjectKey ||
		!gotReceipt.ReceivedAt.Equal(receipt.ReceivedAt) {
		t.Fatalf("expected inbound receipt round trip, got %#v", gotReceipt)
	}

	delivery := domain.WebhookDelivery{
		ID:             "delivery-123",
		OrganizationID: "org-123",
		AgentID:        "agent-123",
		EventType:      "message.received",
		EventID:        "event-123",
		RequestURL:     "https://example.com/webhook",
		RequestPayload: []byte(`{"event":"message.received"}`),
		RequestHeaders: map[string]string{
			"X-AgentLayer-Signature": "abc123",
		},
		Status:        "succeeded",
		AttemptCount:  1,
		ResponseCode:  202,
		ResponseBody:  []byte(`{"ok":true}`),
		LastAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	model, err := WebhookDeliveryToModel(delivery)
	if err != nil {
		t.Fatalf("expected webhook delivery to marshal, got error: %v", err)
	}
	var headers map[string]string
	if err := json.Unmarshal(model.RequestHeaders, &headers); err != nil {
		t.Fatalf("expected webhook headers json, got error: %v", err)
	}
	if headers["X-AgentLayer-Signature"] != "abc123" {
		t.Fatalf("expected marshaled webhook headers, got %#v", headers)
	}

	gotDelivery, err := WebhookDeliveryFromModel(model)
	if err != nil {
		t.Fatalf("expected webhook delivery to unmarshal, got error: %v", err)
	}
	if string(gotDelivery.RequestPayload) != string(delivery.RequestPayload) ||
		gotDelivery.RequestHeaders["X-AgentLayer-Signature"] != "abc123" ||
		string(gotDelivery.ResponseBody) != string(delivery.ResponseBody) {
		t.Fatalf("expected webhook delivery round trip, got %#v", gotDelivery)
	}

	suppression := domain.SuppressedAddress{
		ID:             "suppression-123",
		OrganizationID: "org-123",
		EmailAddress:   "sender@example.com",
		Reason:         "complaint",
		Source:         "ses",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if got := SuppressedAddressFromModel(SuppressedAddressToModel(suppression)); got != suppression {
		t.Fatalf("expected suppression round trip, got %#v", got)
	}

	config := domain.ProviderConfig{
		ID:             "provider-123",
		OrganizationID: "org-123",
		ProviderType:   "ses",
		IsDefault:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if got := ProviderConfigFromModel(ProviderConfigToModel(config, []byte(`{"region":"us-east-1"}`))); got != config {
		t.Fatalf("expected provider config round trip, got %#v", got)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

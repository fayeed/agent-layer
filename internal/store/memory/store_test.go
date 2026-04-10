package memory

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestStoreSupportsSeededInboxAndRawMessageAccess(t *testing.T) {
	store := NewStore()
	store.SeedInbox(domain.Inbox{
		ID:           "inbox-123",
		EmailAddress: "agent@example.com",
	})

	if err := store.Put(context.Background(), "raw/test.eml", []byte("mime")); err != nil {
		t.Fatalf("expected raw put to succeed, got error: %v", err)
	}

	raw, err := store.Get(context.Background(), "raw/test.eml")
	if err != nil {
		t.Fatalf("expected raw get to succeed, got error: %v", err)
	}

	if string(raw) != "mime" {
		t.Fatalf("expected stored raw bytes, got %q", string(raw))
	}

	inbox, found, err := store.FindByEmailAddress(context.Background(), "agent@example.com")
	if err != nil || !found {
		t.Fatalf("expected inbox lookup to succeed, got found=%v err=%v", found, err)
	}

	if inbox.ID != "inbox-123" {
		t.Fatalf("expected seeded inbox, got %#v", inbox)
	}

	receipt := inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-session-123",
		OrganizationID:      "org-123",
		AgentID:             "agent-123",
		InboxID:             "inbox-123",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@example.com"},
		RawMessageObjectKey: "raw/test.eml",
		ReceivedAt:          time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
	if err := store.SaveInboundReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("expected inbound receipt save to succeed, got error: %v", err)
	}

	loadedReceipt, err := store.GetInboundReceiptByObjectKey(context.Background(), "raw/test.eml")
	if err != nil {
		t.Fatalf("expected inbound receipt lookup to succeed, got error: %v", err)
	}

	if loadedReceipt.SMTPTransactionID != "smtp-session-123" {
		t.Fatalf("expected stored inbound receipt, got %#v", loadedReceipt)
	}
}

func TestStoreListsInboundReceiptsByMostRecentFirst(t *testing.T) {
	store := NewStore()

	_ = store.SaveInboundReceipt(context.Background(), inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-older",
		RawMessageObjectKey: "raw/older.eml",
		ReceivedAt:          time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC),
	})
	_ = store.SaveInboundReceipt(context.Background(), inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-newer",
		RawMessageObjectKey: "raw/newer.eml",
		ReceivedAt:          time.Date(2026, 4, 9, 11, 0, 0, 0, time.UTC),
	})

	receipts, err := store.ListInboundReceipts(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected inbound receipt list to succeed, got error: %v", err)
	}

	if len(receipts) != 2 || receipts[0].RawMessageObjectKey != "raw/newer.eml" || receipts[1].RawMessageObjectKey != "raw/older.eml" {
		t.Fatalf("expected inbound receipts ordered by recency, got %#v", receipts)
	}
}

func TestStoreSupportsBootstrapConfigState(t *testing.T) {
	store := NewStore()

	organization, err := store.SaveOrganization(context.Background(), domain.Organization{
		ID:   "org-local",
		Name: "AgentLayer Local",
	})
	if err != nil {
		t.Fatalf("expected organization save to succeed, got error: %v", err)
	}

	agent, err := store.SaveAgent(context.Background(), domain.Agent{
		ID:             "agent-local",
		OrganizationID: "org-local",
		Name:           "Local Agent",
	})
	if err != nil {
		t.Fatalf("expected agent save to succeed, got error: %v", err)
	}

	inbox, err := store.SaveInbox(context.Background(), domain.Inbox{
		ID:             "inbox-local",
		OrganizationID: "org-local",
		AgentID:        "agent-local",
		EmailAddress:   "agent@example.com",
	})
	if err != nil {
		t.Fatalf("expected inbox save to succeed, got error: %v", err)
	}

	if organization.ID != "org-local" || agent.ID != "agent-local" || inbox.ID != "inbox-local" {
		t.Fatalf("expected saved bootstrap config, got org=%#v agent=%#v inbox=%#v", organization, agent, inbox)
	}

	gotOrg, err := store.GetOrganizationByID(context.Background(), "org-local")
	if err != nil || gotOrg.Name != "AgentLayer Local" {
		t.Fatalf("expected organization get to succeed, got org=%#v err=%v", gotOrg, err)
	}

	gotAgent, err := store.GetAgentByID(context.Background(), "agent-local")
	if err != nil || gotAgent.OrganizationID != "org-local" {
		t.Fatalf("expected agent get to succeed, got agent=%#v err=%v", gotAgent, err)
	}

	gotInbox, err := store.GetInboxByID(context.Background(), "inbox-local")
	if err != nil || gotInbox.EmailAddress != "agent@example.com" {
		t.Fatalf("expected inbox get to succeed, got inbox=%#v err=%v", gotInbox, err)
	}
}

func TestStoreSupportsThreadContactAndMessageState(t *testing.T) {
	store := NewStore()

	_, err := store.UpsertByEmail(context.Background(), domain.Contact{
		ID:           "contact-123",
		EmailAddress: "sender@example.com",
	})
	if err != nil {
		t.Fatalf("expected contact upsert to succeed, got error: %v", err)
	}

	_, err = store.Save(context.Background(), domain.Thread{
		ID:        "thread-123",
		ContactID: "contact-123",
	})
	if err != nil {
		t.Fatalf("expected thread save to succeed, got error: %v", err)
	}

	_, err = store.Create(context.Background(), domain.Message{
		ID:              "message-123",
		ThreadID:        "thread-123",
		InboxID:         "inbox-123",
		MessageIDHeader: "<message-123@example.com>",
	})
	if err != nil {
		t.Fatalf("expected message create to succeed, got error: %v", err)
	}

	thread, err := store.GetByID(context.Background(), "thread-123")
	if err != nil {
		t.Fatalf("expected thread get to succeed, got error: %v", err)
	}

	if thread.ID != "thread-123" {
		t.Fatalf("expected stored thread, got %#v", thread)
	}

	contact, err := store.GetContactByID(context.Background(), "contact-123")
	if err != nil {
		t.Fatalf("expected contact get to succeed, got error: %v", err)
	}

	if contact.EmailAddress != "sender@example.com" {
		t.Fatalf("expected stored contact, got %#v", contact)
	}

	messages, err := store.ListByThreadID(context.Background(), "thread-123", 10)
	if err != nil {
		t.Fatalf("expected message list to succeed, got error: %v", err)
	}

	if len(messages) != 1 || messages[0].ID != "message-123" {
		t.Fatalf("expected stored messages, got %#v", messages)
	}

	message, err := store.GetMessageByID(context.Background(), "message-123")
	if err != nil {
		t.Fatalf("expected message get to succeed, got error: %v", err)
	}

	if message.MessageIDHeader != "<message-123@example.com>" {
		t.Fatalf("expected stored message lookup, got %#v", message)
	}

	inboundMessage, found, err := store.FindInboundByHeader(context.Background(), "inbox-123", "<message-123@example.com>")
	if err != nil || !found {
		t.Fatalf("expected inbound message lookup to succeed, got found=%v err=%v", found, err)
	}

	if inboundMessage.ID != "message-123" {
		t.Fatalf("expected inbound message lookup result, got %#v", inboundMessage)
	}
}

func TestStoreSupportsMemoryAndWebhookDeliveryState(t *testing.T) {
	store := NewStore()

	entry, err := store.CreateMemory(context.Background(), domain.ContactMemoryEntry{
		ID:        "memory-123",
		ContactID: "contact-123",
		Note:      "Prefers email.",
	})
	if err != nil {
		t.Fatalf("expected memory create to succeed, got error: %v", err)
	}

	if entry.ID != "memory-123" {
		t.Fatalf("expected created memory entry, got %#v", entry)
	}

	memories, err := store.ListMemoryByContactID(context.Background(), "contact-123", 10)
	if err != nil {
		t.Fatalf("expected memory list to succeed, got error: %v", err)
	}

	if len(memories) != 1 || memories[0].ID != "memory-123" {
		t.Fatalf("expected stored memory entries, got %#v", memories)
	}

	delivery, err := store.SaveWebhookDelivery(context.Background(), domain.WebhookDelivery{
		ID:        "delivery-123",
		EventID:   "event-123",
		EventType: "message.received",
		Status:    "succeeded",
	})
	if err != nil {
		t.Fatalf("expected webhook delivery save to succeed, got error: %v", err)
	}

	if delivery.ID != "delivery-123" {
		t.Fatalf("expected saved webhook delivery, got %#v", delivery)
	}

	loaded, err := store.GetWebhookDeliveryByID(context.Background(), "delivery-123")
	if err != nil {
		t.Fatalf("expected webhook delivery get to succeed, got error: %v", err)
	}

	if loaded.EventID != "event-123" {
		t.Fatalf("expected stored webhook delivery, got %#v", loaded)
	}
}

func TestStoreChecksSuppressedAddresses(t *testing.T) {
	store := NewStore()

	_, err := store.SaveSuppression(context.Background(), domain.SuppressedAddress{
		ID:             "suppression-123",
		OrganizationID: "org-123",
		EmailAddress:   "sender@example.com",
		Reason:         "hard_bounce",
	})
	if err != nil {
		t.Fatalf("expected suppression save to succeed, got error: %v", err)
	}

	suppressed, err := store.IsSuppressed(context.Background(), "org-123", "sender@example.com")
	if err != nil {
		t.Fatalf("expected suppression check to succeed, got error: %v", err)
	}
	if !suppressed {
		t.Fatal("expected suppression check to report true")
	}
}

func TestStoreGetsAndListsSuppressions(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)

	_, _ = store.SaveSuppression(context.Background(), domain.SuppressedAddress{
		ID:             "suppression-older",
		OrganizationID: "org-123",
		EmailAddress:   "older@example.com",
		UpdatedAt:      now,
	})
	_, _ = store.SaveSuppression(context.Background(), domain.SuppressedAddress{
		ID:             "suppression-newer",
		OrganizationID: "org-123",
		EmailAddress:   "newer@example.com",
		UpdatedAt:      now.Add(time.Minute),
	})

	record, err := store.GetSuppressionByID(context.Background(), "suppression-newer")
	if err != nil {
		t.Fatalf("expected suppression get to succeed, got error: %v", err)
	}
	if record.EmailAddress != "newer@example.com" {
		t.Fatalf("expected suppression record, got %#v", record)
	}

	list, err := store.ListSuppressions(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected suppression list to succeed, got error: %v", err)
	}
	if len(list) != 1 || list[0].ID != "suppression-newer" {
		t.Fatalf("expected suppression list result, got %#v", list)
	}
}

func TestStoreSupportsReplySubmissionLookup(t *testing.T) {
	store := NewStore()

	_, err := store.Create(context.Background(), domain.Message{
		ID:       "message-123",
		ThreadID: "thread-123",
	})
	if err != nil {
		t.Fatalf("expected message create to succeed, got error: %v", err)
	}

	if err := store.SaveReplySubmission(context.Background(), "reply:key", "message-123"); err != nil {
		t.Fatalf("expected reply submission save to succeed, got error: %v", err)
	}

	message, found, err := store.FindReplyBySubmissionKey(context.Background(), "reply:key")
	if err != nil || !found {
		t.Fatalf("expected reply submission lookup to succeed, got found=%v err=%v", found, err)
	}

	if message.ID != "message-123" {
		t.Fatalf("expected stored reply submission message, got %#v", message)
	}
}

func TestStoreListsWebhookDeliveriesByMostRecentUpdate(t *testing.T) {
	store := NewStore()

	_, _ = store.SaveWebhookDelivery(context.Background(), domain.WebhookDelivery{
		ID:        "delivery-older",
		EventID:   "event-older",
		UpdatedAt: time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC),
	})
	_, _ = store.SaveWebhookDelivery(context.Background(), domain.WebhookDelivery{
		ID:        "delivery-newer",
		EventID:   "event-newer",
		UpdatedAt: time.Date(2026, 4, 3, 11, 0, 0, 0, time.UTC),
	})

	deliveries, err := store.ListWebhookDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 2 || deliveries[0].ID != "delivery-newer" || deliveries[1].ID != "delivery-older" {
		t.Fatalf("expected deliveries ordered by recency, got %#v", deliveries)
	}
}

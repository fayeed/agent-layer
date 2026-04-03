package memory

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
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
}

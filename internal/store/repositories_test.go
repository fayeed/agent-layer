package store

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestRepositoryInterfacesCompile(t *testing.T) {
	ctx := context.Background()

	var raw RawMessageStore = rawMessageStoreStub{}
	if err := raw.Put(ctx, "raw/message.eml", []byte("mime")); err != nil {
		t.Fatalf("unexpected raw message store error: %v", err)
	}

	var inboxes InboxRepository = inboxRepositoryStub{}
	if _, err := inboxes.GetByEmailAddress(ctx, "agent@example.com"); err != nil {
		t.Fatalf("unexpected inbox repository error: %v", err)
	}

	var contacts ContactRepository = contactRepositoryStub{}
	if _, err := contacts.UpsertByEmail(ctx, domain.Contact{EmailAddress: "sender@example.com"}); err != nil {
		t.Fatalf("unexpected contact repository error: %v", err)
	}

	var threads ThreadRepository = threadRepositoryStub{}
	if _, err := threads.Save(ctx, domain.Thread{ID: "thread-123"}); err != nil {
		t.Fatalf("unexpected thread repository error: %v", err)
	}

	var messages MessageRepository = messageRepositoryStub{}
	if _, err := messages.Create(ctx, domain.Message{ID: "message-123"}); err != nil {
		t.Fatalf("unexpected message repository create error: %v", err)
	}
	if _, err := messages.ListByThreadID(ctx, "thread-123", 10); err != nil {
		t.Fatalf("unexpected message repository list error: %v", err)
	}

	var webhooks WebhookDeliveryRepository = webhookDeliveryRepositoryStub{}
	if _, err := webhooks.Create(ctx, domain.WebhookDelivery{ID: "delivery-123"}); err != nil {
		t.Fatalf("unexpected webhook delivery repository error: %v", err)
	}

	var providers ProviderConfigRepository = providerConfigRepositoryStub{}
	if _, err := providers.GetDefaultByOrganizationID(ctx, "org-123"); err != nil {
		t.Fatalf("unexpected provider config repository error: %v", err)
	}

	var suppressions SuppressedAddressRepository = suppressedAddressRepositoryStub{}
	if _, err := suppressions.IsSuppressed(ctx, "org-123", "sender@example.com"); err != nil {
		t.Fatalf("unexpected suppression repository error: %v", err)
	}
}

type rawMessageStoreStub struct{}

func (rawMessageStoreStub) Put(context.Context, string, []byte) error {
	return nil
}

type inboxRepositoryStub struct{}

func (inboxRepositoryStub) GetByEmailAddress(context.Context, string) (domain.Inbox, error) {
	return domain.Inbox{}, nil
}

type contactRepositoryStub struct{}

func (contactRepositoryStub) UpsertByEmail(context.Context, domain.Contact) (domain.Contact, error) {
	return domain.Contact{}, nil
}

type threadRepositoryStub struct{}

func (threadRepositoryStub) Save(context.Context, domain.Thread) (domain.Thread, error) {
	return domain.Thread{}, nil
}

type messageRepositoryStub struct{}

func (messageRepositoryStub) Create(context.Context, domain.Message) (domain.Message, error) {
	return domain.Message{}, nil
}

func (messageRepositoryStub) ListByThreadID(context.Context, string, int) ([]domain.Message, error) {
	return nil, nil
}

type webhookDeliveryRepositoryStub struct{}

func (webhookDeliveryRepositoryStub) Create(context.Context, domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	return domain.WebhookDelivery{}, nil
}

type providerConfigRepositoryStub struct{}

func (providerConfigRepositoryStub) GetDefaultByOrganizationID(context.Context, string) (domain.ProviderConfig, error) {
	return domain.ProviderConfig{}, nil
}

type suppressedAddressRepositoryStub struct{}

func (suppressedAddressRepositoryStub) IsSuppressed(context.Context, string, string) (bool, error) {
	return false, nil
}

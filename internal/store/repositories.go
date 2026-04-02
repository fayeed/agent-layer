package store

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type RawMessageStore interface {
	Put(ctx context.Context, objectKey string, data []byte) error
}

type InboxRepository interface {
	GetByEmailAddress(ctx context.Context, emailAddress string) (domain.Inbox, error)
}

type ContactRepository interface {
	UpsertByEmail(ctx context.Context, contact domain.Contact) (domain.Contact, error)
}

type ThreadRepository interface {
	Save(ctx context.Context, thread domain.Thread) (domain.Thread, error)
}

type MessageRepository interface {
	Create(ctx context.Context, message domain.Message) (domain.Message, error)
	ListByThreadID(ctx context.Context, threadID string, limit int) ([]domain.Message, error)
}

type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error)
}

type ProviderConfigRepository interface {
	GetDefaultByOrganizationID(ctx context.Context, organizationID string) (domain.ProviderConfig, error)
}

type SuppressedAddressRepository interface {
	IsSuppressed(ctx context.Context, organizationID, emailAddress string) (bool, error)
}

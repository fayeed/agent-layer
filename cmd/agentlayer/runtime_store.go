package main

import (
	"context"
	"database/sql"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/store/blobfs"
	storepg "github.com/agentlayer/agentlayer/internal/store/postgres"
)

type appStore interface {
	SaveOrganization(ctx context.Context, organization domain.Organization) (domain.Organization, error)
	GetOrganizationByID(ctx context.Context, organizationID string) (domain.Organization, error)
	SaveAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error)
	GetAgentByID(ctx context.Context, agentID string) (domain.Agent, error)
	SaveInbox(ctx context.Context, inbox domain.Inbox) (domain.Inbox, error)
	GetInboxByID(ctx context.Context, inboxID string) (domain.Inbox, error)
	FindByEmailAddress(ctx context.Context, emailAddress string) (domain.Inbox, bool, error)
	Put(ctx context.Context, objectKey string, data []byte) error
	Get(ctx context.Context, objectKey string) ([]byte, error)
	SaveInboundReceipt(ctx context.Context, receipt inbound.DurableReceiptRequest) error
	GetInboundReceiptByObjectKey(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error)
	ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error)
	FindByEmail(ctx context.Context, organizationID, emailAddress string) (domain.Contact, bool, error)
	UpsertByEmail(ctx context.Context, contact domain.Contact) (domain.Contact, error)
	GetContactByID(ctx context.Context, contactID string) (domain.Contact, error)
	GetByID(ctx context.Context, threadID string) (domain.Thread, error)
	Save(ctx context.Context, thread domain.Thread) (domain.Thread, error)
	FindByMessageID(ctx context.Context, messageID string) (domain.Thread, bool, error)
	FindMostRecentBySubject(ctx context.Context, organizationID, inboxID, contactID, subjectNormalized string) (domain.Thread, bool, error)
	Create(ctx context.Context, message domain.Message) (domain.Message, error)
	SaveMessage(ctx context.Context, message domain.Message) (domain.Message, error)
	ListByThreadID(ctx context.Context, threadID string, limit int) ([]domain.Message, error)
	FindByProviderMessageID(ctx context.Context, providerMessageID string) (domain.Message, bool, error)
	FindInboundByHeader(ctx context.Context, inboxID, messageIDHeader string) (domain.Message, bool, error)
	GetMessageByID(ctx context.Context, messageID string) (domain.Message, error)
	CreateMemory(ctx context.Context, entry domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error)
	ListMemoryByContactID(ctx context.Context, contactID string, limit int) ([]domain.ContactMemoryEntry, error)
	SaveSuppression(ctx context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error)
	SaveWebhookDelivery(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error)
	GetWebhookDeliveryByID(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error)
	ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)
}

type postgresRuntimeStore struct {
	db           *sql.DB
	raw          blobfs.Store
	bootstrap    storepg.BootstrapStore
	receipts     storepg.InboundReceiptStore
	contacts     storepg.ContactStore
	threads      storepg.ThreadStore
	messages     storepg.MessageStore
	reads        storepg.ReadStore
	memories     storepg.ContactMemoryStore
	webhooks     storepg.WebhookDeliveryStore
	suppressions storepg.SuppressionStore
}

func newPostgresRuntimeStore(db *sql.DB, raw blobfs.Store) *postgresRuntimeStore {
	return &postgresRuntimeStore{
		db:           db,
		raw:          raw,
		bootstrap:    storepg.NewBootstrapStore(db),
		receipts:     storepg.NewInboundReceiptStore(db),
		contacts:     storepg.NewContactStore(db),
		threads:      storepg.NewThreadStore(db),
		messages:     storepg.NewMessageStore(db),
		reads:        storepg.NewReadStore(db),
		memories:     storepg.NewContactMemoryStore(db),
		webhooks:     storepg.NewWebhookDeliveryStore(db),
		suppressions: storepg.NewSuppressionStore(db),
	}
}

func (s *postgresRuntimeStore) SaveOrganization(ctx context.Context, organization domain.Organization) (domain.Organization, error) {
	return s.bootstrap.SaveOrganization(ctx, organization)
}
func (s *postgresRuntimeStore) GetOrganizationByID(ctx context.Context, organizationID string) (domain.Organization, error) {
	return s.bootstrap.GetOrganizationByID(ctx, organizationID)
}
func (s *postgresRuntimeStore) SaveAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error) {
	return s.bootstrap.SaveAgent(ctx, agent)
}
func (s *postgresRuntimeStore) GetAgentByID(ctx context.Context, agentID string) (domain.Agent, error) {
	return s.bootstrap.GetAgentByID(ctx, agentID)
}
func (s *postgresRuntimeStore) SaveInbox(ctx context.Context, inbox domain.Inbox) (domain.Inbox, error) {
	return s.bootstrap.SaveInbox(ctx, inbox)
}
func (s *postgresRuntimeStore) GetInboxByID(ctx context.Context, inboxID string) (domain.Inbox, error) {
	return s.bootstrap.GetInboxByID(ctx, inboxID)
}
func (s *postgresRuntimeStore) FindByEmailAddress(ctx context.Context, emailAddress string) (domain.Inbox, bool, error) {
	return s.bootstrap.FindByEmailAddress(ctx, emailAddress)
}
func (s *postgresRuntimeStore) Put(ctx context.Context, objectKey string, data []byte) error {
	return s.raw.Put(ctx, objectKey, data)
}
func (s *postgresRuntimeStore) Get(ctx context.Context, objectKey string) ([]byte, error) {
	return s.raw.Get(ctx, objectKey)
}
func (s *postgresRuntimeStore) SaveInboundReceipt(ctx context.Context, receipt inbound.DurableReceiptRequest) error {
	return s.receipts.SaveInboundReceipt(ctx, receipt)
}
func (s *postgresRuntimeStore) GetInboundReceiptByObjectKey(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error) {
	return s.receipts.GetInboundReceiptByObjectKey(ctx, objectKey)
}
func (s *postgresRuntimeStore) ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error) {
	return s.receipts.ListInboundReceipts(ctx, limit)
}
func (s *postgresRuntimeStore) FindByEmail(ctx context.Context, organizationID, emailAddress string) (domain.Contact, bool, error) {
	return s.contacts.FindByEmail(ctx, organizationID, emailAddress)
}
func (s *postgresRuntimeStore) UpsertByEmail(ctx context.Context, contact domain.Contact) (domain.Contact, error) {
	return s.contacts.UpsertByEmail(ctx, contact)
}
func (s *postgresRuntimeStore) GetContactByID(ctx context.Context, contactID string) (domain.Contact, error) {
	return s.contacts.GetContactByID(ctx, contactID)
}
func (s *postgresRuntimeStore) GetByID(ctx context.Context, threadID string) (domain.Thread, error) {
	return s.threads.GetByID(ctx, threadID)
}
func (s *postgresRuntimeStore) Save(ctx context.Context, thread domain.Thread) (domain.Thread, error) {
	return s.threads.Save(ctx, thread)
}
func (s *postgresRuntimeStore) FindByMessageID(ctx context.Context, messageID string) (domain.Thread, bool, error) {
	return s.messages.FindByMessageID(ctx, messageID)
}
func (s *postgresRuntimeStore) FindMostRecentBySubject(ctx context.Context, organizationID, inboxID, contactID, subjectNormalized string) (domain.Thread, bool, error) {
	return s.threads.FindMostRecentBySubject(ctx, organizationID, inboxID, contactID, subjectNormalized)
}
func (s *postgresRuntimeStore) Create(ctx context.Context, message domain.Message) (domain.Message, error) {
	return s.messages.Create(ctx, message)
}
func (s *postgresRuntimeStore) SaveMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	return s.messages.Save(ctx, message)
}
func (s *postgresRuntimeStore) ListByThreadID(ctx context.Context, threadID string, limit int) ([]domain.Message, error) {
	return s.reads.ListByThreadID(ctx, threadID, limit)
}
func (s *postgresRuntimeStore) FindByProviderMessageID(ctx context.Context, providerMessageID string) (domain.Message, bool, error) {
	return s.messages.FindByProviderMessageID(ctx, providerMessageID)
}
func (s *postgresRuntimeStore) FindInboundByHeader(ctx context.Context, inboxID, messageIDHeader string) (domain.Message, bool, error) {
	return s.messages.FindInboundByHeader(ctx, inboxID, messageIDHeader)
}
func (s *postgresRuntimeStore) GetMessageByID(ctx context.Context, messageID string) (domain.Message, error) {
	return s.messages.GetMessageByID(ctx, messageID)
}
func (s *postgresRuntimeStore) CreateMemory(ctx context.Context, entry domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error) {
	contact, err := s.contacts.GetContactByID(ctx, entry.ContactID)
	if err != nil {
		return domain.ContactMemoryEntry{}, err
	}
	return s.memories.CreateMemory(ctx, entry, contact.OrganizationID)
}
func (s *postgresRuntimeStore) ListMemoryByContactID(ctx context.Context, contactID string, limit int) ([]domain.ContactMemoryEntry, error) {
	return s.memories.ListMemoryByContactID(ctx, contactID, limit)
}
func (s *postgresRuntimeStore) SaveSuppression(ctx context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	return s.suppressions.Save(ctx, record)
}
func (s *postgresRuntimeStore) SaveWebhookDelivery(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	return s.webhooks.SaveWebhookDelivery(ctx, delivery)
}
func (s *postgresRuntimeStore) GetWebhookDeliveryByID(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	return s.webhooks.GetWebhookDeliveryByID(ctx, deliveryID)
}
func (s *postgresRuntimeStore) ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	return s.webhooks.ListWebhookDeliveries(ctx, limit)
}
func (s *postgresRuntimeStore) Close() error {
	return s.db.Close()
}

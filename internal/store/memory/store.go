package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/platform/idempotency"
)

var (
	ErrThreadNotFound  = errors.New("thread not found")
	ErrContactNotFound = errors.New("contact not found")
	ErrMessageNotFound = errors.New("message not found")
	ErrConfigNotFound  = errors.New("config not found")
)

type Store struct {
	mu sync.RWMutex

	organizationsByID     map[string]domain.Organization
	agentsByID            map[string]domain.Agent
	rawMessages           map[string][]byte
	inboxesByID           map[string]domain.Inbox
	inboxesByEmail        map[string]domain.Inbox
	inboundReceiptsByKey  map[string]inbound.DurableReceiptRequest
	contactsByID          map[string]domain.Contact
	contactsByEmail       map[string]domain.Contact
	threadsByID           map[string]domain.Thread
	messagesByID          map[string]domain.Message
	messagesByInboundKey  map[string]string
	messagesByProviderID  map[string]domain.Message
	messagesByThreadID    map[string][]string
	memoriesByID          map[string]domain.ContactMemoryEntry
	memoriesByContactID   map[string][]string
	suppressionsByID      map[string]domain.SuppressedAddress
	webhookDeliveriesByID map[string]domain.WebhookDelivery
}

func NewStore() *Store {
	return &Store{
		organizationsByID:     make(map[string]domain.Organization),
		agentsByID:            make(map[string]domain.Agent),
		rawMessages:           make(map[string][]byte),
		inboxesByID:           make(map[string]domain.Inbox),
		inboxesByEmail:        make(map[string]domain.Inbox),
		inboundReceiptsByKey:  make(map[string]inbound.DurableReceiptRequest),
		contactsByID:          make(map[string]domain.Contact),
		contactsByEmail:       make(map[string]domain.Contact),
		threadsByID:           make(map[string]domain.Thread),
		messagesByID:          make(map[string]domain.Message),
		messagesByInboundKey:  make(map[string]string),
		messagesByProviderID:  make(map[string]domain.Message),
		messagesByThreadID:    make(map[string][]string),
		memoriesByID:          make(map[string]domain.ContactMemoryEntry),
		memoriesByContactID:   make(map[string][]string),
		suppressionsByID:      make(map[string]domain.SuppressedAddress),
		webhookDeliveriesByID: make(map[string]domain.WebhookDelivery),
	}
}

func (s *Store) SeedInbox(inbox domain.Inbox) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inboxesByID[inbox.ID] = inbox
	s.inboxesByEmail[inbox.EmailAddress] = inbox
}

func (s *Store) SaveOrganization(_ context.Context, organization domain.Organization) (domain.Organization, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.organizationsByID[organization.ID] = organization
	return organization, nil
}

func (s *Store) GetOrganizationByID(_ context.Context, organizationID string) (domain.Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	organization, ok := s.organizationsByID[organizationID]
	if !ok {
		return domain.Organization{}, fmt.Errorf("%w: organization", domain.ErrNotFound)
	}
	return organization, nil
}

func (s *Store) SaveAgent(_ context.Context, agent domain.Agent) (domain.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentsByID[agent.ID] = agent
	return agent, nil
}

func (s *Store) GetAgentByID(_ context.Context, agentID string) (domain.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, ok := s.agentsByID[agentID]
	if !ok {
		return domain.Agent{}, fmt.Errorf("%w: agent", domain.ErrNotFound)
	}
	return agent, nil
}

func (s *Store) SaveInbox(_ context.Context, inbox domain.Inbox) (domain.Inbox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inboxesByID[inbox.ID] = inbox
	s.inboxesByEmail[inbox.EmailAddress] = inbox
	return inbox, nil
}

func (s *Store) GetInboxByID(_ context.Context, inboxID string) (domain.Inbox, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inbox, ok := s.inboxesByID[inboxID]
	if !ok {
		return domain.Inbox{}, fmt.Errorf("%w: inbox", domain.ErrNotFound)
	}
	return inbox, nil
}

func (s *Store) Put(_ context.Context, objectKey string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rawMessages[objectKey] = append([]byte(nil), data...)
	return nil
}

func (s *Store) Get(_ context.Context, objectKey string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.rawMessages[objectKey]
	if !ok {
		return nil, fmt.Errorf("%w: raw message", domain.ErrNotFound)
	}
	return append([]byte(nil), data...), nil
}

func (s *Store) SaveInboundReceipt(_ context.Context, receipt inbound.DurableReceiptRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inboundReceiptsByKey[receipt.RawMessageObjectKey] = receipt
	return nil
}

func (s *Store) GetInboundReceiptByObjectKey(_ context.Context, objectKey string) (inbound.DurableReceiptRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	receipt, ok := s.inboundReceiptsByKey[objectKey]
	if !ok {
		return inbound.DurableReceiptRequest{}, fmt.Errorf("%w: inbound receipt", domain.ErrNotFound)
	}
	return receipt, nil
}

func (s *Store) FindByEmailAddress(_ context.Context, emailAddress string) (domain.Inbox, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inbox, ok := s.inboxesByEmail[emailAddress]
	return inbox, ok, nil
}

func (s *Store) FindByEmail(_ context.Context, _, emailAddress string) (domain.Contact, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	contact, ok := s.contactsByEmail[emailAddress]
	return contact, ok, nil
}

func (s *Store) UpsertByEmail(_ context.Context, contact domain.Contact) (domain.Contact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contactsByID[contact.ID] = contact
	s.contactsByEmail[contact.EmailAddress] = contact
	return contact, nil
}

func (s *Store) GetByID(_ context.Context, threadID string) (domain.Thread, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	thread, ok := s.threadsByID[threadID]
	if !ok {
		return domain.Thread{}, fmt.Errorf("%w: thread", domain.ErrNotFound)
	}
	return thread, nil
}

func (s *Store) GetContactByID(_ context.Context, contactID string) (domain.Contact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	contact, ok := s.contactsByID[contactID]
	if !ok {
		return domain.Contact{}, fmt.Errorf("%w: contact", domain.ErrNotFound)
	}
	return contact, nil
}

func (s *Store) Save(_ context.Context, thread domain.Thread) (domain.Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.threadsByID[thread.ID]; ok {
		if thread.OrganizationID == "" {
			thread.OrganizationID = existing.OrganizationID
		}
		if thread.AgentID == "" {
			thread.AgentID = existing.AgentID
		}
		if thread.InboxID == "" {
			thread.InboxID = existing.InboxID
		}
		if thread.ContactID == "" {
			thread.ContactID = existing.ContactID
		}
		if thread.SubjectNormalized == "" {
			thread.SubjectNormalized = existing.SubjectNormalized
		}
		if thread.State == "" {
			thread.State = existing.State
		}
	}
	s.threadsByID[thread.ID] = thread
	return thread, nil
}

func (s *Store) FindByMessageID(_ context.Context, messageID string) (domain.Thread, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, message := range s.messagesByID {
		if message.MessageIDHeader == messageID {
			thread, ok := s.threadsByID[message.ThreadID]
			return thread, ok, nil
		}
	}
	return domain.Thread{}, false, nil
}

func (s *Store) FindMostRecentBySubject(_ context.Context, organizationID, inboxID, contactID, subjectNormalized string) (domain.Thread, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		best      domain.Thread
		bestFound bool
	)

	for _, thread := range s.threadsByID {
		if thread.OrganizationID != organizationID || thread.InboxID != inboxID || thread.ContactID != contactID {
			continue
		}
		if thread.SubjectNormalized != subjectNormalized {
			continue
		}

		if !bestFound || thread.LastActivityAt.After(best.LastActivityAt) {
			best = thread
			bestFound = true
		}
	}

	return best, bestFound, nil
}

func (s *Store) Create(_ context.Context, message domain.Message) (domain.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesByID[message.ID] = message
	if message.InboxID != "" && message.MessageIDHeader != "" {
		s.messagesByInboundKey[idempotency.InboundMessageKey(message.InboxID, message.MessageIDHeader)] = message.ID
	}
	if message.ProviderMessageID != "" {
		s.messagesByProviderID[message.ProviderMessageID] = message
	}
	if message.ThreadID != "" {
		s.messagesByThreadID[message.ThreadID] = append(s.messagesByThreadID[message.ThreadID], message.ID)
	}
	return message, nil
}

func (s *Store) SaveMessage(_ context.Context, message domain.Message) (domain.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesByID[message.ID] = message
	if message.InboxID != "" && message.MessageIDHeader != "" {
		s.messagesByInboundKey[idempotency.InboundMessageKey(message.InboxID, message.MessageIDHeader)] = message.ID
	}
	if message.ProviderMessageID != "" {
		s.messagesByProviderID[message.ProviderMessageID] = message
	}
	return message, nil
}

func (s *Store) FindInboundByHeader(_ context.Context, inboxID, messageIDHeader string) (domain.Message, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messageID, ok := s.messagesByInboundKey[idempotency.InboundMessageKey(inboxID, messageIDHeader)]
	if !ok {
		return domain.Message{}, false, nil
	}

	message, ok := s.messagesByID[messageID]
	return message, ok, nil
}

func (s *Store) ListByThreadID(_ context.Context, threadID string, limit int) ([]domain.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.messagesByThreadID[threadID]
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}
	out := make([]domain.Message, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.messagesByID[id])
	}
	return out, nil
}

func (s *Store) FindByProviderMessageID(_ context.Context, providerMessageID string) (domain.Message, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	message, ok := s.messagesByProviderID[providerMessageID]
	return message, ok, nil
}

func (s *Store) GetMessageByID(_ context.Context, messageID string) (domain.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	message, ok := s.messagesByID[messageID]
	if !ok {
		return domain.Message{}, fmt.Errorf("%w: message", domain.ErrNotFound)
	}
	return message, nil
}

func (s *Store) CreateMemory(_ context.Context, entry domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memoriesByID[entry.ID] = entry
	if entry.ContactID != "" {
		s.memoriesByContactID[entry.ContactID] = append(s.memoriesByContactID[entry.ContactID], entry.ID)
	}
	return entry, nil
}

func (s *Store) ListMemoryByContactID(_ context.Context, contactID string, limit int) ([]domain.ContactMemoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.memoriesByContactID[contactID]
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}
	out := make([]domain.ContactMemoryEntry, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.memoriesByID[id])
	}
	return out, nil
}

func (s *Store) SaveSuppression(_ context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suppressionsByID[record.ID] = record
	return record, nil
}

func (s *Store) SaveWebhookDelivery(_ context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhookDeliveriesByID[delivery.ID] = delivery
	return delivery, nil
}

func (s *Store) GetWebhookDeliveryByID(_ context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	delivery, ok := s.webhookDeliveriesByID[deliveryID]
	if !ok {
		return domain.WebhookDelivery{}, fmt.Errorf("%w: webhook delivery", domain.ErrNotFound)
	}
	return delivery, nil
}

func (s *Store) ListWebhookDeliveries(_ context.Context, limit int) ([]domain.WebhookDelivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.WebhookDelivery, 0, len(s.webhookDeliveriesByID))
	for _, delivery := range s.webhookDeliveriesByID {
		out = append(out, delivery)
	}

	sort.Slice(out, func(i, j int) bool {
		if !out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].UpdatedAt.After(out[j].UpdatedAt)
		}
		return out[i].ID > out[j].ID
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}

	return out, nil
}

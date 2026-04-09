package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/webhooks"
)

type InboundHandler interface {
	HandleStoredMessage(ctx context.Context, stored core.StoredInboundMessage) (inbound.HandleResult, error)
}

type ContactMemoryLister interface {
	ListMemoryByContactID(ctx context.Context, contactID string, limit int) ([]domain.ContactMemoryEntry, error)
}

type MessageReceivedDeliveryService interface {
	DeliverAndRecordMessageReceived(ctx context.Context, input webhooks.DeliverMessageReceivedInput) (webhooks.DeliveryResult, error)
}

type InboundRuntimeConfig struct {
	Organization  domain.Organization
	Agent         domain.Agent
	Inbox         domain.Inbox
	WebhookURL    string
	WebhookSecret string
	HistoryLimit  int
	MemoryLimit   int
}

type InboundRuntimeService struct {
	inbound  InboundHandler
	messages ThreadMessagesGetter
	memories ContactMemoryLister
	webhooks MessageReceivedDeliveryService
	now      func() time.Time
	config   InboundRuntimeConfig
}

func NewInboundRuntimeService(
	inboundHandler InboundHandler,
	messages ThreadMessagesGetter,
	memories ContactMemoryLister,
	webhookService MessageReceivedDeliveryService,
	now func() time.Time,
	config InboundRuntimeConfig,
) InboundRuntimeService {
	if now == nil {
		now = time.Now
	}
	if config.HistoryLimit <= 0 {
		config.HistoryLimit = 20
	}
	if config.MemoryLimit <= 0 {
		config.MemoryLimit = 10
	}

	return InboundRuntimeService{
		inbound:  inboundHandler,
		messages: messages,
		memories: memories,
		webhooks: webhookService,
		now:      now,
		config:   config,
	}
}

func (s InboundRuntimeService) HandleStoredMessage(ctx context.Context, stored core.StoredInboundMessage) (inbound.HandleResult, error) {
	handled, err := s.inbound.HandleStoredMessage(ctx, stored)
	if err != nil {
		return inbound.HandleResult{}, err
	}

	if handled.Duplicate {
		return handled, nil
	}

	if s.config.WebhookURL == "" || s.config.Agent.Status != domain.AgentStatusActive || s.webhooks == nil {
		return handled, nil
	}

	threadMessages, err := s.messages.ListByThreadID(ctx, handled.Thread.ID, s.config.HistoryLimit)
	if err != nil {
		return inbound.HandleResult{}, err
	}

	memoryEntries, err := s.memories.ListMemoryByContactID(ctx, handled.Contact.ID, s.config.MemoryLimit)
	if err != nil {
		return inbound.HandleResult{}, err
	}

	now := s.now().UTC()
	_, err = s.webhooks.DeliverAndRecordMessageReceived(ctx, webhooks.DeliverMessageReceivedInput{
		URL:           s.config.WebhookURL,
		WebhookSecret: s.config.WebhookSecret,
		BuildInput: webhooks.BuildMessageReceivedInput{
			Organization:   s.config.Organization,
			Agent:          s.config.Agent,
			Inbox:          s.config.Inbox,
			Delivery:       newMessageReceivedDelivery(s.config.Organization.ID, s.config.Agent.ID, now),
			Handled:        handled,
			ThreadMessages: threadMessages,
			Memory:         memoryEntries,
		},
	})
	if err != nil {
		return inbound.HandleResult{}, err
	}

	return handled, nil
}

func newMessageReceivedDelivery(organizationID, agentID string, now time.Time) domain.WebhookDelivery {
	return domain.WebhookDelivery{
		ID:             "delivery-" + randomHexID(),
		OrganizationID: organizationID,
		AgentID:        agentID,
		EventID:        "event-" + randomHexID(),
		EventType:      "message.received",
		Status:         "pending",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func randomHexID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "generated"
	}
	return hex.EncodeToString(buf[:])
}

package core

import (
	"context"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type InboundReceipt struct {
	SMTPTransactionID   string
	OrganizationID      string
	AgentID             string
	InboxID             string
	EnvelopeSender      string
	EnvelopeRecipients  []string
	RawMessageObjectKey string
	ReceivedAt          time.Time
}

type StoredInboundMessage struct {
	Receipt      InboundReceipt
	RawSizeBytes int64
}

type ParsedAddress struct {
	Email       string
	DisplayName string
}

type ParsedAttachment struct {
	FileName           string
	ContentType        string
	ObjectKey          string
	SizeBytes          int64
	ContentID          string
	ContentDisposition string
}

type ParsedMessage struct {
	MessageIDHeader   string
	InReplyTo         string
	References        []string
	Subject           string
	SubjectNormalized string
	TextBody          string
	HTMLBody          string
	From              ParsedAddress
	ReplyTo           []ParsedAddress
	To                []ParsedAddress
	CC                []ParsedAddress
	Attachments       []ParsedAttachment
	RawHeaders        map[string][]string
}

type ThreadResolutionInput struct {
	OrganizationID string
	AgentID        string
	InboxID        string
	ContactID      string
	ParsedMessage  ParsedMessage
	ReceivedAt     time.Time
}

type ThreadResolutionResult struct {
	Thread    domain.Thread
	MatchedBy string
	Created   bool
}

type ContactResolutionInput struct {
	OrganizationID string
	ParsedMessage  ParsedMessage
	ReceivedAt     time.Time
}

type ContactResolutionResult struct {
	Contact domain.Contact
	Created bool
}

type WebhookDispatchRequest struct {
	Delivery domain.WebhookDelivery
	Payload  []byte
	Headers  map[string]string
}

type WebhookDispatchResult struct {
	StatusCode  int
	Body        []byte
	DeliveredAt time.Time
}

type OutboundSendRequest struct {
	Organization domain.Organization
	Agent        domain.Agent
	Inbox        domain.Inbox
	Thread       domain.Thread
	Contact      domain.Contact
	Message      domain.Message
	RawMIME      []byte
}

type SendResult struct {
	ProviderMessageID string
	AcceptedAt        time.Time
}

type DeliveryStatus struct {
	ProviderMessageID string
	State             string
	UpdatedAt         time.Time
}

type ProviderHealth struct {
	ProviderName string
	Healthy      bool
	CheckedAt    time.Time
	Details      string
}

type InboundTransport interface {
	Receive(ctx context.Context, receipt InboundReceipt) error
}

type MessageParser interface {
	Parse(ctx context.Context, message StoredInboundMessage) (ParsedMessage, error)
}

type ThreadResolver interface {
	Resolve(ctx context.Context, input ThreadResolutionInput) (ThreadResolutionResult, error)
}

type ContactResolver interface {
	Resolve(ctx context.Context, input ContactResolutionInput) (ContactResolutionResult, error)
}

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, request WebhookDispatchRequest) (WebhookDispatchResult, error)
}

type EmailProvider interface {
	Send(ctx context.Context, request OutboundSendRequest) (SendResult, error)
	GetDeliveryStatus(ctx context.Context, providerMessageID string) (DeliveryStatus, error)
	HealthCheck(ctx context.Context) (ProviderHealth, error)
}

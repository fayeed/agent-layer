package domain

import "time"

type ThreadState string

const (
	ThreadStateActive    ThreadState = "ACTIVE"
	ThreadStateWaiting   ThreadState = "WAITING"
	ThreadStateEscalated ThreadState = "ESCALATED"
	ThreadStateResolved  ThreadState = "RESOLVED"
	ThreadStateDormant   ThreadState = "DORMANT"
)

func (s ThreadState) IsValid() bool {
	switch s {
	case ThreadStateActive, ThreadStateWaiting, ThreadStateEscalated, ThreadStateResolved, ThreadStateDormant:
		return true
	default:
		return false
	}
}

type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"
	MessageDirectionOutbound MessageDirection = "outbound"
)

func (d MessageDirection) IsValid() bool {
	switch d {
	case MessageDirectionInbound, MessageDirectionOutbound:
		return true
	default:
		return false
	}
}

type AgentStatus string

const (
	AgentStatusActive   AgentStatus = "active"
	AgentStatusPaused   AgentStatus = "paused"
	AgentStatusArchived AgentStatus = "archived"
)

func (s AgentStatus) IsValid() bool {
	switch s {
	case AgentStatusActive, AgentStatusPaused, AgentStatusArchived:
		return true
	default:
		return false
	}
}

type Organization struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Agent struct {
	ID             string
	OrganizationID string
	Name           string
	Status         AgentStatus
	WebhookURL     string
	WebhookSecret  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Inbox struct {
	ID               string
	OrganizationID   string
	AgentID          string
	EmailAddress     string
	Domain           string
	DisplayName      string
	OutboundIdentity string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Contact struct {
	ID             string
	OrganizationID string
	EmailAddress   string
	DisplayName    string
	LastSeenAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Thread struct {
	ID                string
	OrganizationID    string
	AgentID           string
	InboxID           string
	ContactID         string
	SubjectNormalized string
	State             ThreadState
	LastInboundID     string
	LastOutboundID    string
	LastActivityAt    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Message struct {
	ID                string
	OrganizationID    string
	ThreadID          string
	InboxID           string
	ContactID         string
	Direction         MessageDirection
	Subject           string
	SubjectNormalized string
	MessageIDHeader   string
	InReplyTo         string
	References        []string
	TextBody          string
	HTMLBody          string
	RawMIMEObjectKey  string
	CreatedAt         time.Time
}

type ProviderConfig struct {
	ID             string
	OrganizationID string
	ProviderType   string
	IsDefault      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type WebhookDelivery struct {
	ID             string
	OrganizationID string
	AgentID        string
	EventType      string
	EventID        string
	Status         string
	AttemptCount   int
	LastAttemptAt  time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

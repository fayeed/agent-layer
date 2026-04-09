package store

import "time"

type OrganizationModel struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AgentModel struct {
	ID             string
	OrganizationID string
	Name           string
	Status         string
	WebhookURL     string
	WebhookSecret  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type InboxModel struct {
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

type ContactModel struct {
	ID             string
	OrganizationID string
	EmailAddress   string
	DisplayName    string
	LastSeenAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ThreadModel struct {
	ID                string
	OrganizationID    string
	AgentID           string
	InboxID           string
	ContactID         string
	SubjectNormalized string
	State             string
	LastInboundID     string
	LastOutboundID    string
	LastActivityAt    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type MessageModel struct {
	ID                string
	OrganizationID    string
	ThreadID          string
	InboxID           string
	ContactID         string
	Direction         string
	Subject           string
	SubjectNormalized string
	MessageIDHeader   string
	InReplyTo         string
	References        []string
	TextBody          string
	HTMLBody          string
	RawMIMEObjectKey  string

	// Provider and delivery fields remain mutable after create even though the
	// message record itself is otherwise treated as immutable.
	ProviderMessageID string
	DeliveryState     string
	SentAt            time.Time
	DeliveredAt       time.Time
	BouncedAt         time.Time
	CreatedAt         time.Time
}

type MessageAttachmentModel struct {
	ID          string
	MessageID   string
	FileName    string
	ContentType string
	ObjectKey   string
	SizeBytes   int64
	ContentID   string
	Disposition string
	CreatedAt   time.Time
}

type ContactMemoryModel struct {
	ID             string
	OrganizationID string
	ContactID      string
	ThreadID       string
	Note           string
	Tags           []string
	CreatedAt      time.Time
}

type InboundReceiptModel struct {
	RawMessageObjectKey string
	SMTPTransactionID   string
	OrganizationID      string
	AgentID             string
	InboxID             string
	EnvelopeSender      string
	EnvelopeRecipients  []string
	ReceivedAt          time.Time
	CreatedAt           time.Time
}

type WebhookDeliveryModel struct {
	ID             string
	OrganizationID string
	AgentID        string
	EventType      string
	EventID        string
	RequestURL     string
	RequestPayload []byte
	RequestHeaders []byte
	Status         string
	AttemptCount   int
	LastAttemptAt  time.Time
	ResponseCode   int
	ResponseBody   []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SuppressedAddressModel struct {
	ID             string
	OrganizationID string
	EmailAddress   string
	Reason         string
	Source         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProviderConfigModel struct {
	ID             string
	OrganizationID string
	ProviderType   string
	IsDefault      bool
	ConfigJSON     []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AuditLogModel struct {
	ID             string
	OrganizationID string
	EventType      string
	ActorType      string
	ActorID        string
	ResourceType   string
	ResourceID     string
	PayloadJSON    []byte
	CreatedAt      time.Time
}

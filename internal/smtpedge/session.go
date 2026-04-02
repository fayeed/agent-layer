package smtpedge

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

type Clock func() time.Time
type ObjectKeyGenerator func() string

type InboxLookup interface {
	FindByEmailAddress(ctx context.Context, emailAddress string) (domain.Inbox, bool, error)
}

type RawMessageStore interface {
	Put(ctx context.Context, objectKey string, data []byte) error
}

type ReceiptSink interface {
	Enqueue(ctx context.Context, receipt inbound.DurableReceiptRequest) error
}

type Session struct {
	lookup      InboxLookup
	store       RawMessageStore
	sink        ReceiptSink
	now         Clock
	nextKey     ObjectKeyGenerator
	sessionID   string
	sender      string
	recipient   string
	targetInbox domain.Inbox
}

func NewSession(
	lookup InboxLookup,
	store RawMessageStore,
	sink ReceiptSink,
	now Clock,
	nextKey ObjectKeyGenerator,
	sessionID string,
) Session {
	if now == nil {
		now = time.Now
	}
	if nextKey == nil {
		nextKey = func() string { return "raw/generated.eml" }
	}

	return Session{
		lookup:    lookup,
		store:     store,
		sink:      sink,
		now:       now,
		nextKey:   nextKey,
		sessionID: sessionID,
	}
}

func (s *Session) Mail(_ context.Context, from string) error {
	if from == "" {
		return errors.New("mail from is required")
	}
	s.sender = from
	return nil
}

func (s *Session) Rcpt(ctx context.Context, to string) error {
	if to == "" {
		return errors.New("rcpt to is required")
	}

	inbox, found, err := s.lookup.FindByEmailAddress(ctx, to)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("recipient inbox not found")
	}

	s.recipient = to
	s.targetInbox = inbox
	return nil
}

func (s *Session) Data(ctx context.Context, reader io.Reader) error {
	if s.sender == "" {
		return errors.New("mail from must be set before data")
	}
	if s.recipient == "" {
		return errors.New("at least one recipient must be accepted before data")
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	objectKey := s.nextKey()
	if err := s.store.Put(ctx, objectKey, raw); err != nil {
		return err
	}

	return s.sink.Enqueue(ctx, inbound.DurableReceiptRequest{
		SMTPTransactionID:   s.sessionID,
		OrganizationID:      s.targetInbox.OrganizationID,
		AgentID:             s.targetInbox.AgentID,
		InboxID:             s.targetInbox.ID,
		EnvelopeSender:      s.sender,
		EnvelopeRecipients:  []string{s.recipient},
		RawMessageObjectKey: objectKey,
		ReceivedAt:          s.now().UTC(),
	})
}

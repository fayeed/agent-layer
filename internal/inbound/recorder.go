package inbound

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type ContactRepository interface {
	UpsertByEmail(ctx context.Context, contact domain.Contact) (domain.Contact, error)
}

type ThreadRepository interface {
	Save(ctx context.Context, thread domain.Thread) (domain.Thread, error)
}

type MessageRepository interface {
	Create(ctx context.Context, message domain.Message) (domain.Message, error)
}

type RecordResult struct {
	Contact domain.Contact
	Thread  domain.Thread
	Message domain.Message
}

type Recorder struct {
	contacts ContactRepository
	threads  ThreadRepository
	messages MessageRepository
}

func NewRecorder(contacts ContactRepository, threads ThreadRepository, messages MessageRepository) Recorder {
	return Recorder{
		contacts: contacts,
		threads:  threads,
		messages: messages,
	}
}

func (r Recorder) Record(ctx context.Context, stored core.StoredInboundMessage, processed ProcessResult) (RecordResult, error) {
	contact, err := r.contacts.UpsertByEmail(ctx, processed.Contact)
	if err != nil {
		return RecordResult{}, err
	}

	thread, err := r.threads.Save(ctx, processed.Thread)
	if err != nil {
		return RecordResult{}, err
	}

	message, err := r.messages.Create(ctx, domain.Message{
		ID:                newMessageID(),
		OrganizationID:    stored.Receipt.OrganizationID,
		ThreadID:          thread.ID,
		InboxID:           stored.Receipt.InboxID,
		ContactID:         contact.ID,
		Direction:         domain.MessageDirectionInbound,
		Subject:           processed.ParsedMessage.Subject,
		SubjectNormalized: processed.ParsedMessage.SubjectNormalized,
		MessageIDHeader:   processed.ParsedMessage.MessageIDHeader,
		InReplyTo:         processed.ParsedMessage.InReplyTo,
		References:        processed.ParsedMessage.References,
		TextBody:          processed.ParsedMessage.TextBody,
		HTMLBody:          processed.ParsedMessage.HTMLBody,
		RawMIMEObjectKey:  stored.Receipt.RawMessageObjectKey,
		CreatedAt:         stored.Receipt.ReceivedAt,
	})
	if err != nil {
		return RecordResult{}, err
	}

	thread.LastInboundID = message.ID
	thread.LastActivityAt = stored.Receipt.ReceivedAt
	thread.UpdatedAt = stored.Receipt.ReceivedAt

	thread, err = r.threads.Save(ctx, thread)
	if err != nil {
		return RecordResult{}, err
	}

	return RecordResult{
		Contact: contact,
		Thread:  thread,
		Message: message,
	}, nil
}

func newMessageID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "message-generated"
	}
	return "message-" + hex.EncodeToString(buf[:])
}

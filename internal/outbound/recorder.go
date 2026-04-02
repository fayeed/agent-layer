package outbound

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

const DeliveryStateQueued = "queued"

type MessageRepository interface {
	Create(ctx context.Context, message domain.Message) (domain.Message, error)
}

type ThreadRepository interface {
	Save(ctx context.Context, thread domain.Thread) (domain.Thread, error)
}

type RecordQueuedReplyInput struct {
	Organization domain.Organization
	Agent        domain.Agent
	Inbox        domain.Inbox
	Thread       domain.Thread
	Contact      domain.Contact
	Metadata     ReplyMetadata
	RawMIME      string
	ObjectKey    string
	BodyText     string
	QueuedAt     time.Time
}

type RecordQueuedReplyResult struct {
	Thread  domain.Thread
	Message domain.Message
}

type Recorder struct {
	messages MessageRepository
	threads  ThreadRepository
}

func NewRecorder(messages MessageRepository) Recorder {
	return Recorder{
		messages: messages,
	}
}

func NewRecorderWithThreads(messages MessageRepository, threads ThreadRepository) Recorder {
	return Recorder{
		messages: messages,
		threads:  threads,
	}
}

func (r Recorder) RecordQueuedReply(ctx context.Context, input RecordQueuedReplyInput) (RecordQueuedReplyResult, error) {
	message, err := r.messages.Create(ctx, domain.Message{
		ID:               newOutboundMessageID(),
		OrganizationID:   input.Organization.ID,
		ThreadID:         input.Thread.ID,
		InboxID:          input.Inbox.ID,
		ContactID:        input.Contact.ID,
		Direction:        domain.MessageDirectionOutbound,
		Subject:          input.Metadata.Subject,
		MessageIDHeader:  input.Metadata.MessageIDHeader,
		InReplyTo:        input.Metadata.InReplyTo,
		References:       input.Metadata.References,
		TextBody:         input.BodyText,
		RawMIMEObjectKey: input.ObjectKey,
		DeliveryState:    DeliveryStateQueued,
		CreatedAt:        input.QueuedAt,
	})
	if err != nil {
		return RecordQueuedReplyResult{}, err
	}

	thread := input.Thread
	if r.threads != nil {
		thread.LastOutboundID = message.ID
		thread.LastActivityAt = input.QueuedAt
		thread.UpdatedAt = input.QueuedAt

		thread, err = r.threads.Save(ctx, thread)
		if err != nil {
			return RecordQueuedReplyResult{}, err
		}
	}

	return RecordQueuedReplyResult{
		Thread:  thread,
		Message: message,
	}, nil
}

func newOutboundMessageID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "outbound-message-generated"
	}
	return "message-" + hex.EncodeToString(buf[:])
}

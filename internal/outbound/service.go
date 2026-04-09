package outbound

import (
	"context"
	"errors"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type Clock func() time.Time

type ReplyAssembler interface {
	AssembleReply(input ReplyAssemblyInput) (string, ReplyMetadata, error)
}

type QueueRecorder interface {
	RecordQueuedReply(ctx context.Context, input RecordQueuedReplyInput) (RecordQueuedReplyResult, error)
}

type SenderInterface interface {
	SendQueuedReply(ctx context.Context, input SendQueuedReplyInput) (core.SendResult, error)
}

type StatusRecorderInterface interface {
	RecordSent(ctx context.Context, input RecordSentInput) (domain.Message, error)
}

type SuppressionChecker interface {
	IsSuppressed(ctx context.Context, organizationID, emailAddress string) (bool, error)
}

var ErrRecipientSuppressed = errors.New("recipient is suppressed")

type SendReplyInput struct {
	Organization   domain.Organization
	Agent          domain.Agent
	Inbox          domain.Inbox
	Thread         domain.Thread
	ReplyToMessage domain.Message
	Contact        domain.Contact
	BodyText       string
	ObjectKey      string
	IdempotencyKey string
}

type SendReplyResult struct {
	Thread     domain.Thread
	Message    domain.Message
	SendResult core.SendResult
}

type Service struct {
	assembler      ReplyAssembler
	queueRecorder  QueueRecorder
	sender         SenderInterface
	statusRecorder StatusRecorderInterface
	suppressions   SuppressionChecker
	now            Clock
}

func NewService(
	assembler ReplyAssembler,
	queueRecorder QueueRecorder,
	sender SenderInterface,
	statusRecorder StatusRecorderInterface,
	suppressions SuppressionChecker,
	now Clock,
) Service {
	if now == nil {
		now = time.Now
	}

	return Service{
		assembler:      assembler,
		queueRecorder:  queueRecorder,
		sender:         sender,
		statusRecorder: statusRecorder,
		suppressions:   suppressions,
		now:            now,
	}
}

func (s Service) SendReply(ctx context.Context, input SendReplyInput) (SendReplyResult, error) {
	if s.suppressions != nil {
		suppressed, err := s.suppressions.IsSuppressed(ctx, input.Organization.ID, input.Contact.EmailAddress)
		if err != nil {
			return SendReplyResult{}, err
		}
		if suppressed {
			return SendReplyResult{}, ErrRecipientSuppressed
		}
	}

	rawMIME, metadata, err := s.assembler.AssembleReply(ReplyAssemblyInput{
		Organization:   input.Organization,
		Agent:          input.Agent,
		Inbox:          input.Inbox,
		Thread:         input.Thread,
		ReplyToMessage: input.ReplyToMessage,
		Contact:        input.Contact,
		BodyText:       input.BodyText,
	})
	if err != nil {
		return SendReplyResult{}, err
	}

	queued, err := s.queueRecorder.RecordQueuedReply(ctx, RecordQueuedReplyInput{
		Organization: input.Organization,
		Agent:        input.Agent,
		Inbox:        input.Inbox,
		Thread:       input.Thread,
		Contact:      input.Contact,
		Metadata:     metadata,
		RawMIME:      rawMIME,
		ObjectKey:    input.ObjectKey,
		BodyText:     input.BodyText,
		QueuedAt:     s.now().UTC(),
	})
	if err != nil {
		return SendReplyResult{}, err
	}

	sendResult, err := s.sender.SendQueuedReply(ctx, SendQueuedReplyInput{
		Organization: input.Organization,
		Agent:        input.Agent,
		Inbox:        input.Inbox,
		Thread:       queued.Thread,
		Contact:      input.Contact,
		Message:      queued.Message,
		RawMIME:      []byte(rawMIME),
	})
	if err != nil {
		return SendReplyResult{}, err
	}

	message, err := s.statusRecorder.RecordSent(ctx, RecordSentInput{
		Message:    queued.Message,
		SendResult: sendResult,
	})
	if err != nil {
		return SendReplyResult{}, err
	}

	return SendReplyResult{
		Thread:     queued.Thread,
		Message:    message,
		SendResult: sendResult,
	}, nil
}

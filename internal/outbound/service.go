package outbound

import (
	"context"
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

type SendReplyInput struct {
	Organization   domain.Organization
	Agent          domain.Agent
	Inbox          domain.Inbox
	Thread         domain.Thread
	ReplyToMessage domain.Message
	Contact        domain.Contact
	BodyText       string
	ObjectKey      string
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
	now            Clock
}

func NewService(
	assembler ReplyAssembler,
	queueRecorder QueueRecorder,
	sender SenderInterface,
	statusRecorder StatusRecorderInterface,
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
		now:            now,
	}
}

func (s Service) SendReply(ctx context.Context, input SendReplyInput) (SendReplyResult, error) {
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
		Message:      queued.Message,
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

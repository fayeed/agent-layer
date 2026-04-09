package inbound

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type ProcessorInterface interface {
	Process(ctx context.Context, message core.StoredInboundMessage) (ProcessResult, error)
}

type RecorderInterface interface {
	Record(ctx context.Context, stored core.StoredInboundMessage, processed ProcessResult) (RecordResult, error)
}

type DuplicateLookupInterface interface {
	FindInboundByHeader(ctx context.Context, inboxID, messageIDHeader string) (domain.Message, bool, error)
	GetByID(ctx context.Context, threadID string) (domain.Thread, error)
	GetContactByID(ctx context.Context, contactID string) (domain.Contact, error)
}

type HandleResult struct {
	ParsedMessage       core.ParsedMessage
	Contact             domain.Contact
	Thread              domain.Thread
	Message             domain.Message
	ThreadMatchStrategy string
	ThreadCreated       bool
	Duplicate           bool
}

type Service struct {
	processor  ProcessorInterface
	recorder   RecorderInterface
	duplicates DuplicateLookupInterface
}

func NewService(processor ProcessorInterface, recorder RecorderInterface) Service {
	return Service{
		processor: processor,
		recorder:  recorder,
	}
}

func NewServiceWithDuplicateLookup(processor ProcessorInterface, recorder RecorderInterface, duplicates DuplicateLookupInterface) Service {
	return Service{
		processor:  processor,
		recorder:   recorder,
		duplicates: duplicates,
	}
}

func (s Service) HandleStoredMessage(ctx context.Context, stored core.StoredInboundMessage) (HandleResult, error) {
	processed, err := s.processor.Process(ctx, stored)
	if err != nil {
		return HandleResult{}, err
	}

	if s.duplicates != nil && processed.ParsedMessage.MessageIDHeader != "" {
		existing, found, err := s.duplicates.FindInboundByHeader(ctx, stored.Receipt.InboxID, processed.ParsedMessage.MessageIDHeader)
		if err != nil {
			return HandleResult{}, err
		}
		if found {
			thread, err := s.duplicates.GetByID(ctx, existing.ThreadID)
			if err != nil {
				return HandleResult{}, err
			}

			contact, err := s.duplicates.GetContactByID(ctx, existing.ContactID)
			if err != nil {
				return HandleResult{}, err
			}

			return HandleResult{
				ParsedMessage:       processed.ParsedMessage,
				Contact:             contact,
				Thread:              thread,
				Message:             existing,
				ThreadMatchStrategy: "duplicate",
				ThreadCreated:       false,
				Duplicate:           true,
			}, nil
		}
	}

	recorded, err := s.recorder.Record(ctx, stored, processed)
	if err != nil {
		return HandleResult{}, err
	}

	return HandleResult{
		ParsedMessage:       processed.ParsedMessage,
		Contact:             recorded.Contact,
		Thread:              recorded.Thread,
		Message:             recorded.Message,
		ThreadMatchStrategy: processed.ThreadMatchStrategy,
		ThreadCreated:       processed.ThreadCreated,
	}, nil
}

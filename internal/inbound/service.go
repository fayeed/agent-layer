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

type HandleResult struct {
	ParsedMessage       core.ParsedMessage
	Contact             domain.Contact
	Thread              domain.Thread
	Message             domain.Message
	ThreadMatchStrategy string
	ThreadCreated       bool
}

type Service struct {
	processor ProcessorInterface
	recorder  RecorderInterface
}

func NewService(processor ProcessorInterface, recorder RecorderInterface) Service {
	return Service{
		processor: processor,
		recorder:  recorder,
	}
}

func (s Service) HandleStoredMessage(ctx context.Context, stored core.StoredInboundMessage) (HandleResult, error) {
	processed, err := s.processor.Process(ctx, stored)
	if err != nil {
		return HandleResult{}, err
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

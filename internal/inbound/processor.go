package inbound

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type MessageParser interface {
	Parse(ctx context.Context, message core.StoredInboundMessage) (core.ParsedMessage, error)
}

type ContactResolver interface {
	Resolve(ctx context.Context, input core.ContactResolutionInput) (core.ContactResolutionResult, error)
}

type ThreadResolver interface {
	Resolve(ctx context.Context, input core.ThreadResolutionInput) (core.ThreadResolutionResult, error)
}

type Processor struct {
	parser          MessageParser
	contactResolver ContactResolver
	threadResolver  ThreadResolver
}

type ProcessResult struct {
	ParsedMessage       core.ParsedMessage
	Contact             domain.Contact
	Thread              domain.Thread
	ThreadMatchStrategy string
	ThreadCreated       bool
}

func NewProcessor(parser MessageParser, contactResolver ContactResolver, threadResolver ThreadResolver) Processor {
	return Processor{
		parser:          parser,
		contactResolver: contactResolver,
		threadResolver:  threadResolver,
	}
}

func (p Processor) Process(ctx context.Context, message core.StoredInboundMessage) (ProcessResult, error) {
	parsed, err := p.parser.Parse(ctx, message)
	if err != nil {
		return ProcessResult{}, err
	}

	contactResult, err := p.contactResolver.Resolve(ctx, core.ContactResolutionInput{
		OrganizationID: message.Receipt.OrganizationID,
		ParsedMessage:  parsed,
		ReceivedAt:     message.Receipt.ReceivedAt,
	})
	if err != nil {
		return ProcessResult{}, err
	}

	threadResult, err := p.threadResolver.Resolve(ctx, core.ThreadResolutionInput{
		OrganizationID: message.Receipt.OrganizationID,
		AgentID:        message.Receipt.AgentID,
		InboxID:        message.Receipt.InboxID,
		ContactID:      contactResult.Contact.ID,
		ParsedMessage:  parsed,
		ReceivedAt:     message.Receipt.ReceivedAt,
	})
	if err != nil {
		return ProcessResult{}, err
	}

	return ProcessResult{
		ParsedMessage:       parsed,
		Contact:             contactResult.Contact,
		Thread:              threadResult.Thread,
		ThreadMatchStrategy: threadResult.MatchedBy,
		ThreadCreated:       threadResult.Created,
	}, nil
}

package inbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestProcessorParsesAndResolvesInboundMessage(t *testing.T) {
	receivedAt := time.Date(2026, 4, 2, 23, 0, 0, 0, time.UTC)

	processor := NewProcessor(
		messageParserStub{
			parsed: core.ParsedMessage{
				Subject:           "Re: Hello World",
				SubjectNormalized: "hello world",
				From: core.ParsedAddress{
					Email:       "sender@example.com",
					DisplayName: "Sender Example",
				},
			},
		},
		contactResolverStub{
			result: core.ContactResolutionResult{
				Contact: domain.Contact{
					ID:             "contact-123",
					OrganizationID: "org-123",
					EmailAddress:   "sender@example.com",
				},
			},
		},
		threadResolverStub{
			result: core.ThreadResolutionResult{
				Thread: domain.Thread{
					ID:             "thread-123",
					OrganizationID: "org-123",
					AgentID:        "agent-123",
					InboxID:        "inbox-123",
					ContactID:      "contact-123",
				},
				MatchedBy: "in_reply_to",
			},
		},
	)

	result, err := processor.Process(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          receivedAt,
		},
	})
	if err != nil {
		t.Fatalf("expected process to succeed, got error: %v", err)
	}

	if result.ParsedMessage.SubjectNormalized != "hello world" {
		t.Fatalf("expected parsed message to be returned, got %#v", result.ParsedMessage)
	}

	if result.Contact.ID != "contact-123" {
		t.Fatalf("expected resolved contact, got %#v", result.Contact)
	}

	if result.Thread.ID != "thread-123" {
		t.Fatalf("expected resolved thread, got %#v", result.Thread)
	}

	if result.ThreadMatchStrategy != "in_reply_to" {
		t.Fatalf("expected thread match strategy to be returned, got %q", result.ThreadMatchStrategy)
	}
}

func TestProcessorPassesParsedSenderIntoContactResolution(t *testing.T) {
	parser := messageParserStub{
		parsed: core.ParsedMessage{
			From: core.ParsedAddress{
				Email:       "sender@example.com",
				DisplayName: "Sender Example",
			},
		},
	}
	contacts := &capturingContactResolver{}

	processor := NewProcessor(
		parser,
		contacts,
		threadResolverStub{
			result: core.ThreadResolutionResult{
				Thread: domain.Thread{ID: "thread-123"},
			},
		},
	)

	_, err := processor.Process(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 2, 23, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("expected process to succeed, got error: %v", err)
	}

	if contacts.input.ParsedMessage.From.Email != "sender@example.com" {
		t.Fatalf("expected parsed sender to be passed to contact resolution, got %#v", contacts.input.ParsedMessage.From)
	}
}

type messageParserStub struct {
	parsed core.ParsedMessage
	err    error
}

func (s messageParserStub) Parse(context.Context, core.StoredInboundMessage) (core.ParsedMessage, error) {
	return s.parsed, s.err
}

type contactResolverStub struct {
	result core.ContactResolutionResult
	err    error
}

func (s contactResolverStub) Resolve(context.Context, core.ContactResolutionInput) (core.ContactResolutionResult, error) {
	return s.result, s.err
}

type capturingContactResolver struct {
	input  core.ContactResolutionInput
	result core.ContactResolutionResult
	err    error
}

func (s *capturingContactResolver) Resolve(_ context.Context, input core.ContactResolutionInput) (core.ContactResolutionResult, error) {
	s.input = input
	return s.result, s.err
}

type threadResolverStub struct {
	result core.ThreadResolutionResult
	err    error
}

func (s threadResolverStub) Resolve(context.Context, core.ThreadResolutionInput) (core.ThreadResolutionResult, error) {
	return s.result, s.err
}

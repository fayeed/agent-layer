package inbound

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestServiceHandlesStoredInboundMessage(t *testing.T) {
	processed := ProcessResult{
		ParsedMessage: core.ParsedMessage{
			Subject:           "Re: Hello World",
			SubjectNormalized: "hello world",
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		ThreadMatchStrategy: "in_reply_to",
		ThreadCreated:       false,
	}

	recorded := RecordResult{
		Contact: processed.Contact,
		Thread:  processed.Thread,
		Message: domain.Message{
			ID:                "message-123",
			ThreadID:          "thread-123",
			Subject:           "Re: Hello World",
			SubjectNormalized: "hello world",
		},
	}

	service := NewService(
		processorStub{result: processed},
		recorderStub{result: recorded},
	)

	result, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
		},
	})
	if err != nil {
		t.Fatalf("expected service to succeed, got error: %v", err)
	}

	if result.Message.ID != "message-123" {
		t.Fatalf("expected recorded message in result, got %#v", result.Message)
	}

	if result.ThreadMatchStrategy != "in_reply_to" {
		t.Fatalf("expected thread match strategy to be preserved, got %q", result.ThreadMatchStrategy)
	}

	if result.Contact.ID != "contact-123" {
		t.Fatalf("expected contact in result, got %#v", result.Contact)
	}
}

func TestServicePassesProcessedResultIntoRecorder(t *testing.T) {
	recorder := &capturingRecorder{}
	service := NewService(
		processorStub{
			result: ProcessResult{
				ParsedMessage: core.ParsedMessage{
					SubjectNormalized: "hello world",
				},
				Contact: domain.Contact{
					ID: "contact-123",
				},
				Thread: domain.Thread{
					ID: "thread-123",
				},
				ThreadMatchStrategy: "new_thread",
				ThreadCreated:       true,
			},
		},
		recorder,
	)

	_, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
		},
	})
	if err != nil {
		t.Fatalf("expected service to succeed, got error: %v", err)
	}

	if recorder.processed.ThreadMatchStrategy != "new_thread" {
		t.Fatalf("expected recorder to receive processed result, got %#v", recorder.processed)
	}
}

func TestServiceReturnsExistingInboundMessageForDuplicate(t *testing.T) {
	service := NewServiceWithDuplicateLookup(
		processorStub{
			result: ProcessResult{
				ParsedMessage: core.ParsedMessage{
					MessageIDHeader: "<message-123@example.com>",
				},
			},
		},
		&capturingRecorder{},
		duplicateLookupStub{
			message: domain.Message{
				ID:              "message-123",
				ThreadID:        "thread-123",
				ContactID:       "contact-123",
				MessageIDHeader: "<message-123@example.com>",
			},
			thread: domain.Thread{
				ID: "thread-123",
			},
			contact: domain.Contact{
				ID:           "contact-123",
				EmailAddress: "sender@example.com",
			},
			found: true,
		},
	)

	result, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
		},
	})
	if err != nil {
		t.Fatalf("expected duplicate handling to succeed, got error: %v", err)
	}

	if !result.Duplicate {
		t.Fatal("expected duplicate result to be marked")
	}

	if result.Message.ID != "message-123" || result.Thread.ID != "thread-123" || result.Contact.ID != "contact-123" {
		t.Fatalf("expected existing runtime state to be returned, got %#v", result)
	}

	if result.ThreadMatchStrategy != "duplicate" {
		t.Fatalf("expected duplicate match strategy, got %q", result.ThreadMatchStrategy)
	}
}

type processorStub struct {
	result ProcessResult
	err    error
}

func (s processorStub) Process(context.Context, core.StoredInboundMessage) (ProcessResult, error) {
	return s.result, s.err
}

type recorderStub struct {
	result RecordResult
	err    error
}

func (s recorderStub) Record(context.Context, core.StoredInboundMessage, ProcessResult) (RecordResult, error) {
	return s.result, s.err
}

type capturingRecorder struct {
	stored    core.StoredInboundMessage
	processed ProcessResult
	result    RecordResult
	err       error
	calls     int
}

func (s *capturingRecorder) Record(_ context.Context, stored core.StoredInboundMessage, processed ProcessResult) (RecordResult, error) {
	s.calls++
	s.stored = stored
	s.processed = processed
	return s.result, s.err
}

type duplicateLookupStub struct {
	message domain.Message
	thread  domain.Thread
	contact domain.Contact
	found   bool
	err     error
}

func (s duplicateLookupStub) FindInboundByHeader(context.Context, string, string) (domain.Message, bool, error) {
	return s.message, s.found, s.err
}

func (s duplicateLookupStub) GetByID(context.Context, string) (domain.Thread, error) {
	return s.thread, s.err
}

func (s duplicateLookupStub) GetContactByID(context.Context, string) (domain.Contact, error) {
	return s.contact, s.err
}

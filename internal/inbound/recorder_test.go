package inbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestRecorderPersistsResolvedInboundState(t *testing.T) {
	contacts := &contactRepositoryRecorderStub{}
	threads := &threadRepositoryRecorderStub{}
	messages := &messageRepositoryRecorderStub{}

	recorder := NewRecorder(contacts, threads, messages)
	receivedAt := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)

	result, err := recorder.Record(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          receivedAt,
		},
	}, ProcessResult{
		ParsedMessage: core.ParsedMessage{
			MessageIDHeader:   "<message-123@example.com>",
			InReplyTo:         "<message-122@example.com>",
			References:        []string{"<message-100@example.com>", "<message-122@example.com>"},
			Subject:           "Re: Hello World",
			SubjectNormalized: "hello world",
			TextBody:          "Plain body.",
			HTMLBody:          "<p>HTML body.</p>",
		},
		Contact: domain.Contact{
			ID:             "contact-123",
			OrganizationID: "org-123",
			EmailAddress:   "sender@example.com",
		},
		Thread: domain.Thread{
			ID:                "thread-123",
			OrganizationID:    "org-123",
			AgentID:           "agent-123",
			InboxID:           "inbox-123",
			ContactID:         "contact-123",
			SubjectNormalized: "hello world",
			State:             domain.ThreadStateActive,
		},
	})
	if err != nil {
		t.Fatalf("expected record to succeed, got error: %v", err)
	}

	if contacts.saved.ID != "contact-123" {
		t.Fatalf("expected contact to be saved, got %#v", contacts.saved)
	}

	if threads.saved.ID != "thread-123" {
		t.Fatalf("expected thread to be saved, got %#v", threads.saved)
	}

	if messages.created.ThreadID != "thread-123" {
		t.Fatalf("expected inbound message to be linked to thread, got %#v", messages.created)
	}

	if messages.created.Direction != domain.MessageDirectionInbound {
		t.Fatalf("expected inbound message direction, got %q", messages.created.Direction)
	}

	if messages.created.RawMIMEObjectKey != "raw/test-message.eml" {
		t.Fatalf("expected raw object key to be preserved, got %q", messages.created.RawMIMEObjectKey)
	}

	if result.Message.SubjectNormalized != "hello world" {
		t.Fatalf("expected recorded message to be returned, got %#v", result.Message)
	}
}

func TestRecorderUpdatesThreadPointersToRecordedInboundMessage(t *testing.T) {
	threads := &threadRepositoryRecorderStub{}
	messages := &messageRepositoryRecorderStub{
		returned: domain.Message{ID: "message-123"},
	}

	recorder := NewRecorder(
		&contactRepositoryRecorderStub{},
		threads,
		messages,
	)

	_, err := recorder.Record(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
		},
	}, ProcessResult{
		ParsedMessage: core.ParsedMessage{
			Subject:           "Hello World",
			SubjectNormalized: "hello world",
		},
		Contact: domain.Contact{ID: "contact-123"},
		Thread: domain.Thread{
			ID:             "thread-123",
			OrganizationID: "org-123",
			AgentID:        "agent-123",
			InboxID:        "inbox-123",
			ContactID:      "contact-123",
		},
	})
	if err != nil {
		t.Fatalf("expected record to succeed, got error: %v", err)
	}

	if threads.saved.LastInboundID != "message-123" {
		t.Fatalf("expected last inbound pointer to be updated, got %q", threads.saved.LastInboundID)
	}
}

type contactRepositoryRecorderStub struct {
	saved domain.Contact
}

func (s *contactRepositoryRecorderStub) UpsertByEmail(_ context.Context, contact domain.Contact) (domain.Contact, error) {
	s.saved = contact
	return contact, nil
}

type threadRepositoryRecorderStub struct {
	saved domain.Thread
}

func (s *threadRepositoryRecorderStub) Save(_ context.Context, thread domain.Thread) (domain.Thread, error) {
	s.saved = thread
	return thread, nil
}

type messageRepositoryRecorderStub struct {
	created  domain.Message
	returned domain.Message
}

func (s *messageRepositoryRecorderStub) Create(_ context.Context, message domain.Message) (domain.Message, error) {
	s.created = message
	if s.returned.ID != "" {
		return s.returned, nil
	}
	if message.ID == "" {
		message.ID = "message-generated"
	}
	return message, nil
}

func (s *messageRepositoryRecorderStub) ListByThreadID(context.Context, string, int) ([]domain.Message, error) {
	return nil, nil
}

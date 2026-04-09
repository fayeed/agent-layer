package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestRecorderCreatesQueuedOutboundMessage(t *testing.T) {
	records := &messageRepositoryStub{}
	recorder := NewRecorder(records)
	at := time.Date(2026, 4, 3, 7, 0, 0, 0, time.UTC)

	result, err := recorder.RecordQueuedReply(context.Background(), RecordQueuedReplyInput{
		Organization: domain.Organization{
			ID: "org-123",
		},
		Agent: domain.Agent{
			ID: "agent-123",
		},
		Inbox: domain.Inbox{
			ID: "inbox-123",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		Contact: domain.Contact{
			ID: "contact-123",
		},
		Metadata: ReplyMetadata{
			MessageIDHeader: "<reply-123@agentlayer.local>",
			Subject:         "Re: Hello World",
			InReplyTo:       "<message-100@example.com>",
			References:      []string{"<message-001@example.com>", "<message-100@example.com>"},
		},
		RawMIME:   "raw mime",
		QueuedAt:  at,
		ObjectKey: "outbound/reply-123.eml",
		BodyText:  "Thanks for reaching out.",
	})
	if err != nil {
		t.Fatalf("expected queued reply record to succeed, got error: %v", err)
	}

	if records.created.Direction != domain.MessageDirectionOutbound {
		t.Fatalf("expected outbound direction, got %#v", records.created)
	}

	if records.created.ThreadID != "thread-123" {
		t.Fatalf("expected thread id on outbound message, got %#v", records.created)
	}

	if records.created.RawMIMEObjectKey != "outbound/reply-123.eml" {
		t.Fatalf("expected raw mime object key, got %#v", records.created)
	}

	if records.created.MessageIDHeader != "<reply-123@agentlayer.local>" {
		t.Fatalf("expected message id header, got %#v", records.created)
	}

	if records.created.DeliveryState != DeliveryStateQueued {
		t.Fatalf("expected queued delivery state, got %#v", records.created)
	}

	if result.Message.Subject != "Re: Hello World" {
		t.Fatalf("expected recorded message subject, got %#v", result.Message)
	}
}

func TestRecorderUpdatesThreadLastOutboundPointer(t *testing.T) {
	messages := &messageRepositoryStub{
		returned: domain.Message{ID: "message-123"},
	}
	threads := &threadRepositoryStub{}

	recorder := NewRecorderWithThreads(messages, threads)
	at := time.Date(2026, 4, 3, 7, 5, 0, 0, time.UTC)

	_, err := recorder.RecordQueuedReply(context.Background(), RecordQueuedReplyInput{
		Organization: domain.Organization{ID: "org-123"},
		Agent:        domain.Agent{ID: "agent-123"},
		Inbox:        domain.Inbox{ID: "inbox-123"},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		Contact:   domain.Contact{ID: "contact-123"},
		Metadata:  ReplyMetadata{MessageIDHeader: "<reply-123@agentlayer.local>", Subject: "Re: Hello"},
		RawMIME:   "raw mime",
		ObjectKey: "outbound/reply-123.eml",
		QueuedAt:  at,
	})
	if err != nil {
		t.Fatalf("expected queued reply record to succeed, got error: %v", err)
	}

	if threads.saved.LastOutboundID != "message-123" {
		t.Fatalf("expected last outbound pointer to be updated, got %#v", threads.saved)
	}

	if !threads.saved.LastActivityAt.Equal(at) {
		t.Fatalf("expected last activity timestamp to be updated, got %#v", threads.saved)
	}
}

func TestRecorderPersistsRawMIMEWhenStoreIsConfigured(t *testing.T) {
	messages := &messageRepositoryStub{}
	raw := &rawMessageStoreStub{}
	recorder := NewRecorderWithStore(messages, nil, raw)

	_, err := recorder.RecordQueuedReply(context.Background(), RecordQueuedReplyInput{
		Organization: domain.Organization{ID: "org-123"},
		Agent:        domain.Agent{ID: "agent-123"},
		Inbox:        domain.Inbox{ID: "inbox-123"},
		Thread:       domain.Thread{ID: "thread-123"},
		Contact:      domain.Contact{ID: "contact-123"},
		Metadata:     ReplyMetadata{MessageIDHeader: "<reply-123@agentlayer.local>", Subject: "Re: Hello"},
		RawMIME:      "raw mime",
		ObjectKey:    "outbound/reply-123.eml",
		QueuedAt:     time.Date(2026, 4, 3, 7, 5, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected raw mime persistence to succeed, got error: %v", err)
	}

	if raw.objectKey != "outbound/reply-123.eml" || string(raw.data) != "raw mime" {
		t.Fatalf("expected raw mime to be stored, got %#v", raw)
	}
}

type messageRepositoryStub struct {
	created  domain.Message
	returned domain.Message
}

func (s *messageRepositoryStub) Create(_ context.Context, message domain.Message) (domain.Message, error) {
	s.created = message
	if s.returned.ID != "" {
		return s.returned, nil
	}
	if message.ID == "" {
		message.ID = "message-generated"
	}
	return message, nil
}

type threadRepositoryStub struct {
	saved domain.Thread
}

func (s *threadRepositoryStub) Save(_ context.Context, thread domain.Thread) (domain.Thread, error) {
	s.saved = thread
	return thread, nil
}

type rawMessageStoreStub struct {
	objectKey string
	data      []byte
}

func (s *rawMessageStoreStub) Put(_ context.Context, objectKey string, data []byte) error {
	s.objectKey = objectKey
	s.data = append([]byte(nil), data...)
	return nil
}

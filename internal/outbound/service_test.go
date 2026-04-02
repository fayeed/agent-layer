package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestServiceAssemblesQueuesSendsAndRecordsReply(t *testing.T) {
	queuedAt := time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC)

	service := NewService(
		assemblerStub{
			rawMIME: "mime-body",
			metadata: ReplyMetadata{
				MessageIDHeader: "<reply-123@agentlayer.local>",
				Subject:         "Re: Hello World",
				InReplyTo:       "<message-100@example.com>",
				References:      []string{"<message-100@example.com>"},
			},
		},
		queueRecorderStub{
			result: RecordQueuedReplyResult{
				Thread: domain.Thread{
					ID: "thread-123",
				},
				Message: domain.Message{
					ID:              "message-123",
					ThreadID:        "thread-123",
					MessageIDHeader: "<reply-123@agentlayer.local>",
					DeliveryState:   DeliveryStateQueued,
				},
			},
		},
		senderStub{
			result: core.SendResult{
				ProviderMessageID: "ses-123",
				AcceptedAt:        queuedAt.Add(1 * time.Minute),
			},
		},
		statusRecorderStub{
			result: domain.Message{
				ID:                "message-123",
				ThreadID:          "thread-123",
				ProviderMessageID: "ses-123",
				DeliveryState:     DeliveryStateSent,
			},
		},
		func() time.Time { return queuedAt },
	)

	result, err := service.SendReply(context.Background(), SendReplyInput{
		Organization: domain.Organization{ID: "org-123"},
		Agent:        domain.Agent{ID: "agent-123"},
		Inbox: domain.Inbox{
			ID:           "inbox-123",
			EmailAddress: "agent@example.com",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		ReplyToMessage: domain.Message{
			ID:              "message-100",
			ThreadID:        "thread-123",
			MessageIDHeader: "<message-100@example.com>",
			Subject:         "Hello World",
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
		BodyText:  "Thanks for reaching out.",
		ObjectKey: "outbound/reply-123.eml",
	})
	if err != nil {
		t.Fatalf("expected send reply to succeed, got error: %v", err)
	}

	if result.Message.DeliveryState != DeliveryStateSent {
		t.Fatalf("expected final message to be sent, got %#v", result.Message)
	}

	if result.SendResult.ProviderMessageID != "ses-123" {
		t.Fatalf("expected provider result in response, got %#v", result.SendResult)
	}
}

func TestServicePassesQueuedMessageIntoSenderAndStatusRecorder(t *testing.T) {
	sender := &capturingSender{}
	statuses := &capturingStatusRecorder{}

	service := NewService(
		assemblerStub{
			rawMIME: "mime-body",
			metadata: ReplyMetadata{
				MessageIDHeader: "<reply-123@agentlayer.local>",
				Subject:         "Re: Hello World",
			},
		},
		queueRecorderStub{
			result: RecordQueuedReplyResult{
				Thread: domain.Thread{ID: "thread-123"},
				Message: domain.Message{
					ID:            "message-123",
					DeliveryState: DeliveryStateQueued,
				},
			},
		},
		sender,
		statuses,
		func() time.Time { return time.Date(2026, 4, 3, 9, 0, 0, 0, time.UTC) },
	)

	_, err := service.SendReply(context.Background(), SendReplyInput{
		Organization: domain.Organization{ID: "org-123"},
		Agent:        domain.Agent{ID: "agent-123"},
		Inbox: domain.Inbox{
			ID:           "inbox-123",
			EmailAddress: "agent@example.com",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		ReplyToMessage: domain.Message{
			ID:              "message-100",
			ThreadID:        "thread-123",
			MessageIDHeader: "<message-100@example.com>",
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
		ObjectKey: "outbound/reply-123.eml",
	})
	if err != nil {
		t.Fatalf("expected send reply to succeed, got error: %v", err)
	}

	if sender.input.Message.ID != "message-123" {
		t.Fatalf("expected queued message to be sent, got %#v", sender.input)
	}

	if statuses.input.Message.ID != "message-123" {
		t.Fatalf("expected queued message to be passed into status recorder, got %#v", statuses.input)
	}
}

type assemblerStub struct {
	rawMIME  string
	metadata ReplyMetadata
	err      error
}

func (s assemblerStub) AssembleReply(ReplyAssemblyInput) (string, ReplyMetadata, error) {
	return s.rawMIME, s.metadata, s.err
}

type queueRecorderStub struct {
	result RecordQueuedReplyResult
	err    error
}

func (s queueRecorderStub) RecordQueuedReply(context.Context, RecordQueuedReplyInput) (RecordQueuedReplyResult, error) {
	return s.result, s.err
}

type senderStub struct {
	result core.SendResult
	err    error
}

func (s senderStub) SendQueuedReply(context.Context, SendQueuedReplyInput) (core.SendResult, error) {
	return s.result, s.err
}

type capturingSender struct {
	input  SendQueuedReplyInput
	result core.SendResult
	err    error
}

func (s *capturingSender) SendQueuedReply(_ context.Context, input SendQueuedReplyInput) (core.SendResult, error) {
	s.input = input
	return s.result, s.err
}

type statusRecorderStub struct {
	result domain.Message
	err    error
}

func (s statusRecorderStub) RecordSent(context.Context, RecordSentInput) (domain.Message, error) {
	return s.result, s.err
}

type capturingStatusRecorder struct {
	input  RecordSentInput
	result domain.Message
	err    error
}

func (s *capturingStatusRecorder) RecordSent(_ context.Context, input RecordSentInput) (domain.Message, error) {
	s.input = input
	return s.result, s.err
}

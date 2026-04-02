package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestSenderCallsProviderWithOutboundContext(t *testing.T) {
	provider := &providerStub{
		result: core.SendResult{
			ProviderMessageID: "ses-123",
			AcceptedAt:        time.Date(2026, 4, 3, 8, 0, 0, 0, time.UTC),
		},
	}

	sender := NewSender(provider)

	result, err := sender.SendQueuedReply(context.Background(), SendQueuedReplyInput{
		Organization: domain.Organization{
			ID: "org-123",
		},
		Agent: domain.Agent{
			ID: "agent-123",
		},
		Inbox: domain.Inbox{
			ID:           "inbox-123",
			EmailAddress: "agent@example.com",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		Message: domain.Message{
			ID:              "message-123",
			ThreadID:        "thread-123",
			MessageIDHeader: "<reply-123@agentlayer.local>",
			DeliveryState:   DeliveryStateQueued,
		},
	})
	if err != nil {
		t.Fatalf("expected send to succeed, got error: %v", err)
	}

	if provider.request.Message.ID != "message-123" {
		t.Fatalf("expected provider to receive outbound message, got %#v", provider.request)
	}

	if result.ProviderMessageID != "ses-123" {
		t.Fatalf("expected provider result to be returned, got %#v", result)
	}
}

func TestSenderReturnsProviderErrors(t *testing.T) {
	provider := &providerStub{err: context.DeadlineExceeded}
	sender := NewSender(provider)

	_, err := sender.SendQueuedReply(context.Background(), SendQueuedReplyInput{
		Inbox: domain.Inbox{
			EmailAddress: "agent@example.com",
		},
		Message: domain.Message{
			ID: "message-123",
		},
	})
	if err == nil {
		t.Fatal("expected provider error to be returned")
	}
}

type providerStub struct {
	request core.OutboundSendRequest
	result  core.SendResult
	err     error
}

func (s *providerStub) Send(_ context.Context, request core.OutboundSendRequest) (core.SendResult, error) {
	s.request = request
	return s.result, s.err
}

func (s *providerStub) GetDeliveryStatus(context.Context, string) (core.DeliveryStatus, error) {
	return core.DeliveryStatus{}, nil
}

func (s *providerStub) HealthCheck(context.Context) (core.ProviderHealth, error) {
	return core.ProviderHealth{}, nil
}

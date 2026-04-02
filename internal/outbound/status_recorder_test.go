package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestStatusRecorderMarksMessageSent(t *testing.T) {
	repository := &messageStatusRepositoryStub{}
	recorder := NewStatusRecorder(repository)
	acceptedAt := time.Date(2026, 4, 3, 8, 30, 0, 0, time.UTC)

	result, err := recorder.RecordSent(context.Background(), RecordSentInput{
		Message: domain.Message{
			ID:            "message-123",
			DeliveryState: DeliveryStateQueued,
		},
		SendResult: core.SendResult{
			ProviderMessageID: "ses-123",
			AcceptedAt:        acceptedAt,
		},
	})
	if err != nil {
		t.Fatalf("expected record sent to succeed, got error: %v", err)
	}

	if repository.saved.DeliveryState != DeliveryStateSent {
		t.Fatalf("expected sent delivery state, got %#v", repository.saved)
	}

	if repository.saved.ProviderMessageID != "ses-123" {
		t.Fatalf("expected provider message id, got %#v", repository.saved)
	}

	if !repository.saved.SentAt.Equal(acceptedAt) {
		t.Fatalf("expected sent timestamp %v, got %#v", acceptedAt, repository.saved)
	}

	if result.DeliveryState != DeliveryStateSent {
		t.Fatalf("expected updated message to be returned, got %#v", result)
	}
}

type messageStatusRepositoryStub struct {
	saved domain.Message
}

func (s *messageStatusRepositoryStub) Save(_ context.Context, message domain.Message) (domain.Message, error) {
	s.saved = message
	return message, nil
}

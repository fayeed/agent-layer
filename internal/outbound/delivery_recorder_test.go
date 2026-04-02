package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestDeliveryRecorderMarksMessageDelivered(t *testing.T) {
	repository := &deliveryMessageRepositoryStub{}
	recorder := NewDeliveryRecorder(repository)
	at := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)

	result, err := recorder.RecordStatus(context.Background(), RecordDeliveryStatusInput{
		Message: domain.Message{
			ID:                "message-123",
			ProviderMessageID: "ses-123",
			DeliveryState:     DeliveryStateSent,
		},
		Status:     DeliveryStateDelivered,
		OccurredAt: at,
	})
	if err != nil {
		t.Fatalf("expected delivered status to persist, got error: %v", err)
	}

	if repository.saved.DeliveryState != DeliveryStateDelivered {
		t.Fatalf("expected delivered state, got %#v", repository.saved)
	}

	if !repository.saved.DeliveredAt.Equal(at) {
		t.Fatalf("expected delivered timestamp %v, got %#v", at, repository.saved)
	}

	if result.DeliveryState != DeliveryStateDelivered {
		t.Fatalf("expected updated message to be returned, got %#v", result)
	}
}

func TestDeliveryRecorderMarksMessageBounced(t *testing.T) {
	repository := &deliveryMessageRepositoryStub{}
	recorder := NewDeliveryRecorder(repository)
	at := time.Date(2026, 4, 3, 12, 5, 0, 0, time.UTC)

	result, err := recorder.RecordStatus(context.Background(), RecordDeliveryStatusInput{
		Message: domain.Message{
			ID:                "message-123",
			ProviderMessageID: "ses-123",
			DeliveryState:     DeliveryStateSent,
		},
		Status:     DeliveryStateHardBounce,
		OccurredAt: at,
	})
	if err != nil {
		t.Fatalf("expected bounce status to persist, got error: %v", err)
	}

	if repository.saved.DeliveryState != DeliveryStateHardBounce {
		t.Fatalf("expected hard bounce state, got %#v", repository.saved)
	}

	if !repository.saved.BouncedAt.Equal(at) {
		t.Fatalf("expected bounced timestamp %v, got %#v", at, repository.saved)
	}

	if result.DeliveryState != DeliveryStateHardBounce {
		t.Fatalf("expected updated message to be returned, got %#v", result)
	}
}

func TestDeliveryRecorderMarksMessageComplained(t *testing.T) {
	repository := &deliveryMessageRepositoryStub{}
	recorder := NewDeliveryRecorder(repository)
	at := time.Date(2026, 4, 3, 12, 10, 0, 0, time.UTC)

	result, err := recorder.RecordStatus(context.Background(), RecordDeliveryStatusInput{
		Message: domain.Message{
			ID:                "message-123",
			ProviderMessageID: "ses-123",
			DeliveryState:     DeliveryStateSent,
		},
		Status:     DeliveryStateComplaint,
		OccurredAt: at,
	})
	if err != nil {
		t.Fatalf("expected complaint status to persist, got error: %v", err)
	}

	if repository.saved.DeliveryState != DeliveryStateComplaint {
		t.Fatalf("expected complaint state, got %#v", repository.saved)
	}

	if !repository.saved.BouncedAt.Equal(at) {
		t.Fatalf("expected complaint timestamp to reuse bounced marker %v, got %#v", at, repository.saved)
	}

	if result.DeliveryState != DeliveryStateComplaint {
		t.Fatalf("expected updated message to be returned, got %#v", result)
	}
}

type deliveryMessageRepositoryStub struct {
	saved domain.Message
}

func (s *deliveryMessageRepositoryStub) Save(_ context.Context, message domain.Message) (domain.Message, error) {
	s.saved = message
	return message, nil
}

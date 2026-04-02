package webhooks

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestRecorderMarksSuccessfulDelivery(t *testing.T) {
	repository := &deliveryRepositoryStub{}
	recorder := NewRecorder(repository)
	at := time.Date(2026, 4, 3, 5, 0, 0, 0, time.UTC)

	record, err := recorder.RecordAttempt(context.Background(), RecordAttemptInput{
		Delivery: domain.WebhookDelivery{
			ID:             "delivery-123",
			OrganizationID: "org-123",
			AgentID:        "agent-123",
			EventID:        "event-123",
			EventType:      "message.received",
			Status:         "pending",
			AttemptCount:   0,
			CreatedAt:      at.Add(-1 * time.Minute),
			UpdatedAt:      at.Add(-1 * time.Minute),
		},
		Response: core.WebhookDispatchResult{
			StatusCode:  http.StatusAccepted,
			Body:        []byte(`{"ok":true}`),
			DeliveredAt: at,
		},
	})
	if err != nil {
		t.Fatalf("expected record attempt to succeed, got error: %v", err)
	}

	if repository.saved.Status != DeliveryStatusSucceeded {
		t.Fatalf("expected successful status, got %#v", repository.saved)
	}

	if repository.saved.AttemptCount != 1 {
		t.Fatalf("expected attempt count to increment, got %d", repository.saved.AttemptCount)
	}

	if repository.saved.ResponseCode != http.StatusAccepted {
		t.Fatalf("expected response code to be stored, got %d", repository.saved.ResponseCode)
	}

	if string(repository.saved.ResponseBody) != `{"ok":true}` {
		t.Fatalf("expected response body to be stored, got %q", string(repository.saved.ResponseBody))
	}

	if !record.UpdatedAt.Equal(at) {
		t.Fatalf("expected updated record timestamp %v, got %v", at, record.UpdatedAt)
	}
}

func TestRecorderMarksFailedDelivery(t *testing.T) {
	repository := &deliveryRepositoryStub{}
	recorder := NewRecorder(repository)
	at := time.Date(2026, 4, 3, 5, 5, 0, 0, time.UTC)

	record, err := recorder.RecordAttempt(context.Background(), RecordAttemptInput{
		Delivery: domain.WebhookDelivery{
			ID:             "delivery-123",
			OrganizationID: "org-123",
			AgentID:        "agent-123",
			EventID:        "event-123",
			EventType:      "message.received",
			Status:         "pending",
			AttemptCount:   2,
			CreatedAt:      at.Add(-10 * time.Minute),
			UpdatedAt:      at.Add(-10 * time.Minute),
		},
		Response: core.WebhookDispatchResult{
			StatusCode:  http.StatusInternalServerError,
			Body:        []byte(`{"error":"boom"}`),
			DeliveredAt: at,
		},
	})
	if err != nil {
		t.Fatalf("expected record attempt to succeed, got error: %v", err)
	}

	if repository.saved.Status != DeliveryStatusFailed {
		t.Fatalf("expected failed status, got %#v", repository.saved)
	}

	if repository.saved.AttemptCount != 3 {
		t.Fatalf("expected attempt count to increment, got %d", repository.saved.AttemptCount)
	}

	if record.Status != DeliveryStatusFailed {
		t.Fatalf("expected returned status to be failed, got %q", record.Status)
	}
}

type deliveryRepositoryStub struct {
	saved domain.WebhookDelivery
}

func (s *deliveryRepositoryStub) Save(_ context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	s.saved = delivery
	return delivery, nil
}

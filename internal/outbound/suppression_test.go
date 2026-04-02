package outbound

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestSuppressionServiceSuppressesOnHardBounce(t *testing.T) {
	repository := &suppressionRepositoryStub{}
	service := NewSuppressionService(repository)
	at := time.Date(2026, 4, 3, 14, 0, 0, 0, time.UTC)

	record, changed, err := service.Apply(context.Background(), SuppressionInput{
		Message: domain.Message{
			ID:             "message-123",
			OrganizationID: "org-123",
			ContactID:      "contact-123",
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
		Status:     DeliveryStateHardBounce,
		OccurredAt: at,
	})
	if err != nil {
		t.Fatalf("expected suppression apply to succeed, got error: %v", err)
	}

	if !changed {
		t.Fatal("expected hard bounce to create suppression")
	}

	if repository.saved.EmailAddress != "sender@example.com" {
		t.Fatalf("expected suppressed email to be persisted, got %#v", repository.saved)
	}

	if repository.saved.Reason != DeliveryStateHardBounce {
		t.Fatalf("expected suppression reason to be hard bounce, got %#v", repository.saved)
	}

	if !record.CreatedAt.Equal(at) {
		t.Fatalf("expected suppression timestamp %v, got %#v", at, record)
	}
}

func TestSuppressionServiceSuppressesOnComplaint(t *testing.T) {
	repository := &suppressionRepositoryStub{}
	service := NewSuppressionService(repository)

	record, changed, err := service.Apply(context.Background(), SuppressionInput{
		Message: domain.Message{
			OrganizationID: "org-123",
		},
		Contact: domain.Contact{
			EmailAddress: "sender@example.com",
		},
		Status:     DeliveryStateComplaint,
		OccurredAt: time.Date(2026, 4, 3, 14, 5, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected suppression apply to succeed, got error: %v", err)
	}

	if !changed {
		t.Fatal("expected complaint to create suppression")
	}

	if record.Reason != DeliveryStateComplaint {
		t.Fatalf("expected complaint reason, got %#v", record)
	}
}

func TestSuppressionServiceIgnoresNonSuppressingStatuses(t *testing.T) {
	repository := &suppressionRepositoryStub{}
	service := NewSuppressionService(repository)

	record, changed, err := service.Apply(context.Background(), SuppressionInput{
		Message: domain.Message{
			OrganizationID: "org-123",
		},
		Contact: domain.Contact{
			EmailAddress: "sender@example.com",
		},
		Status:     DeliveryStateDelivered,
		OccurredAt: time.Date(2026, 4, 3, 14, 10, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected suppression apply to succeed, got error: %v", err)
	}

	if changed {
		t.Fatal("expected delivered status to avoid suppression")
	}

	if record != (domain.SuppressedAddress{}) {
		t.Fatalf("expected zero suppression record, got %#v", record)
	}
}

type suppressionRepositoryStub struct {
	saved domain.SuppressedAddress
}

func (s *suppressionRepositoryStub) Save(_ context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	s.saved = record
	return record, nil
}

package contacts

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestResolverReturnsExistingContactAndRefreshesDisplayName(t *testing.T) {
	repository := contactRepositoryStub{
		existing: domain.Contact{
			ID:             "contact-123",
			OrganizationID: "org-123",
			EmailAddress:   "sender@example.com",
			DisplayName:    "Old Name",
			CreatedAt:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		},
		found: true,
	}

	resolver := NewResolver(repository)
	receivedAt := time.Date(2026, 4, 2, 22, 0, 0, 0, time.UTC)

	result, err := resolver.Resolve(context.Background(), core.ContactResolutionInput{
		OrganizationID: "org-123",
		ParsedMessage: core.ParsedMessage{
			From: core.ParsedAddress{
				Email:       "sender@example.com",
				DisplayName: "New Name",
			},
		},
		ReceivedAt: receivedAt,
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed, got error: %v", err)
	}

	if result.Created {
		t.Fatal("expected existing contact, not create")
	}

	if result.Contact.ID != "contact-123" {
		t.Fatalf("expected existing contact id, got %q", result.Contact.ID)
	}

	if result.Contact.DisplayName != "New Name" {
		t.Fatalf("expected display name refresh, got %q", result.Contact.DisplayName)
	}

	if !result.Contact.LastSeenAt.Equal(receivedAt) {
		t.Fatalf("expected last seen %v, got %v", receivedAt, result.Contact.LastSeenAt)
	}
}

func TestResolverCreatesContactWhenMissing(t *testing.T) {
	repository := contactRepositoryStub{}
	resolver := NewResolver(repository)
	receivedAt := time.Date(2026, 4, 2, 22, 30, 0, 0, time.UTC)

	result, err := resolver.Resolve(context.Background(), core.ContactResolutionInput{
		OrganizationID: "org-123",
		ParsedMessage: core.ParsedMessage{
			From: core.ParsedAddress{
				Email:       "new.sender@example.com",
				DisplayName: "Sender Example",
			},
		},
		ReceivedAt: receivedAt,
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed, got error: %v", err)
	}

	if !result.Created {
		t.Fatal("expected contact creation on miss")
	}

	if result.Contact.OrganizationID != "org-123" {
		t.Fatalf("expected organization id to be set, got %q", result.Contact.OrganizationID)
	}

	if result.Contact.EmailAddress != "new.sender@example.com" {
		t.Fatalf("expected email address to be set, got %q", result.Contact.EmailAddress)
	}

	if result.Contact.DisplayName != "Sender Example" {
		t.Fatalf("expected display name to be set, got %q", result.Contact.DisplayName)
	}

	if !result.Contact.CreatedAt.Equal(receivedAt) {
		t.Fatalf("expected created at %v, got %v", receivedAt, result.Contact.CreatedAt)
	}
}

type contactRepositoryStub struct {
	existing domain.Contact
	found    bool
}

func (s contactRepositoryStub) FindByEmail(_ context.Context, _, _ string) (domain.Contact, bool, error) {
	return s.existing, s.found, nil
}

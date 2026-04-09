package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestSuppressionStoreSavesSuppressedAddresses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 22, 0, 0, 0, time.UTC)
	record := domain.SuppressedAddress{
		ID:             "suppression-123",
		OrganizationID: "org-123",
		EmailAddress:   "sender@example.com",
		Reason:         "complaint",
		Source:         "ses",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO suppressed_addresses (
			id, organization_id, email_address, reason, source, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (organization_id, email_address) DO UPDATE
		SET id = EXCLUDED.id,
		    reason = EXCLUDED.reason,
		    source = EXCLUDED.source,
		    updated_at = EXCLUDED.updated_at
	`)).
		WithArgs(record.ID, record.OrganizationID, record.EmailAddress, record.Reason, record.Source, record.CreatedAt, record.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	store := NewSuppressionStore(db)
	if _, err := store.Save(context.Background(), record); err != nil {
		t.Fatalf("expected suppression save to succeed, got error: %v", err)
	}
}

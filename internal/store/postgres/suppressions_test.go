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

func TestSuppressionStoreChecksSuppressedAddresses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(
			SELECT 1
			FROM suppressed_addresses
			WHERE organization_id = $1 AND email_address = $2
		)
	`)).
		WithArgs("org-123", "sender@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	store := NewSuppressionStore(db)
	suppressed, err := store.IsSuppressed(context.Background(), "org-123", "sender@example.com")
	if err != nil {
		t.Fatalf("expected suppression check to succeed, got error: %v", err)
	}
	if !suppressed {
		t.Fatal("expected suppression check to report true")
	}
}

func TestSuppressionStoreGetsAndListsSuppressions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 22, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "organization_id", "email_address", "reason", "source", "created_at", "updated_at",
	}).AddRow(
		"suppression-123", "org-123", "sender@example.com", "hard_bounce", "ses", now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, email_address, reason, source, created_at, updated_at
		FROM suppressed_addresses
		WHERE id = $1
	`)).
		WithArgs("suppression-123").
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, email_address, reason, source, created_at, updated_at
		FROM suppressed_addresses
		ORDER BY updated_at DESC, id DESC
		LIMIT $1
	`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "email_address", "reason", "source", "created_at", "updated_at",
		}).AddRow(
			"suppression-123", "org-123", "sender@example.com", "hard_bounce", "ses", now, now,
		))

	store := NewSuppressionStore(db)
	record, err := store.GetByID(context.Background(), "suppression-123")
	if err != nil {
		t.Fatalf("expected suppression get to succeed, got error: %v", err)
	}
	if record.ID != "suppression-123" {
		t.Fatalf("expected suppression record, got %#v", record)
	}

	list, err := store.List(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected suppression list to succeed, got error: %v", err)
	}
	if len(list) != 1 || list[0].ID != "suppression-123" {
		t.Fatalf("expected suppression list result, got %#v", list)
	}
}

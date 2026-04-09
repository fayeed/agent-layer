package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestContactStoreFindsUpsertsAndGetsContacts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 20, 30, 0, 0, time.UTC)
	contact := domain.Contact{
		ID:             "contact-123",
		OrganizationID: "org-123",
		EmailAddress:   "sender@example.com",
		DisplayName:    "Sender Example",
		LastSeenAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO contacts (
			id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (organization_id, email_address) DO UPDATE
		SET id = EXCLUDED.id,
		    display_name = EXCLUDED.display_name,
		    last_seen_at = EXCLUDED.last_seen_at,
		    updated_at = EXCLUDED.updated_at
	`)).
		WithArgs(contact.ID, contact.OrganizationID, contact.EmailAddress, contact.DisplayName, contact.LastSeenAt, contact.CreatedAt, contact.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{
		"id", "organization_id", "email_address", "display_name", "last_seen_at", "created_at", "updated_at",
	}).AddRow(contact.ID, contact.OrganizationID, contact.EmailAddress, contact.DisplayName, contact.LastSeenAt, contact.CreatedAt, contact.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		FROM contacts
		WHERE organization_id = $1 AND email_address = $2
	`)).
		WithArgs(contact.OrganizationID, contact.EmailAddress).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`)).
		WithArgs(contact.ID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "email_address", "display_name", "last_seen_at", "created_at", "updated_at",
		}).AddRow(contact.ID, contact.OrganizationID, contact.EmailAddress, contact.DisplayName, contact.LastSeenAt, contact.CreatedAt, contact.UpdatedAt))

	store := NewContactStore(db)
	if _, err := store.UpsertByEmail(context.Background(), contact); err != nil {
		t.Fatalf("expected contact upsert to succeed, got error: %v", err)
	}
	found, ok, err := store.FindByEmail(context.Background(), contact.OrganizationID, contact.EmailAddress)
	if err != nil || !ok {
		t.Fatalf("expected contact find to succeed, got ok=%v err=%v", ok, err)
	}
	if found.DisplayName != contact.DisplayName {
		t.Fatalf("expected found contact, got %#v", found)
	}
	got, err := store.GetContactByID(context.Background(), contact.ID)
	if err != nil {
		t.Fatalf("expected contact get to succeed, got error: %v", err)
	}
	if got.EmailAddress != contact.EmailAddress {
		t.Fatalf("expected contact get result, got %#v", got)
	}
}

func TestContactStoreMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	store := NewContactStore(db)
	_, err = store.GetContactByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}

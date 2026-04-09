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

func TestBootstrapStoreSavesAndLoadsBootstrapState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 18, 0, 0, 0, time.UTC)
	inbox := domain.Inbox{
		ID:               "inbox-123",
		OrganizationID:   "org-123",
		AgentID:          "agent-123",
		EmailAddress:     "agent@example.com",
		Domain:           "example.com",
		DisplayName:      "Acme Inbox",
		OutboundIdentity: "acme-support",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO inboxes (
			id, organization_id, agent_id, email_address, domain, display_name, outbound_identity, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE
		SET organization_id = EXCLUDED.organization_id,
		    agent_id = EXCLUDED.agent_id,
		    email_address = EXCLUDED.email_address,
		    domain = EXCLUDED.domain,
		    display_name = EXCLUDED.display_name,
		    outbound_identity = EXCLUDED.outbound_identity,
		    updated_at = EXCLUDED.updated_at
	`)).
		WithArgs(
			inbox.ID, inbox.OrganizationID, inbox.AgentID, inbox.EmailAddress, inbox.Domain,
			inbox.DisplayName, inbox.OutboundIdentity, inbox.CreatedAt, inbox.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{
		"id", "organization_id", "agent_id", "email_address", "domain", "display_name", "outbound_identity", "created_at", "updated_at",
	}).AddRow(
		inbox.ID, inbox.OrganizationID, inbox.AgentID, inbox.EmailAddress, inbox.Domain,
		inbox.DisplayName, inbox.OutboundIdentity, inbox.CreatedAt, inbox.UpdatedAt,
	)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, email_address, domain, display_name, outbound_identity, created_at, updated_at
		FROM inboxes
		WHERE id = $1
	`)).
		WithArgs(inbox.ID).
		WillReturnRows(rows)

	store := NewBootstrapStore(db)

	if _, err := store.SaveInbox(context.Background(), inbox); err != nil {
		t.Fatalf("expected inbox save to succeed, got error: %v", err)
	}

	loaded, err := store.GetInboxByID(context.Background(), inbox.ID)
	if err != nil {
		t.Fatalf("expected inbox lookup to succeed, got error: %v", err)
	}

	if loaded.EmailAddress != inbox.EmailAddress || loaded.OutboundIdentity != inbox.OutboundIdentity {
		t.Fatalf("expected inbox round trip through postgres store, got %#v", loaded)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected sqlmock expectations to be met, got error: %v", err)
	}
}

func TestBootstrapStoreMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	store := NewBootstrapStore(db)

	_, err = store.GetOrganizationByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected sqlmock expectations to be met, got error: %v", err)
	}
}

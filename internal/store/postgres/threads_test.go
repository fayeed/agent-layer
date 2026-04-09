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

func TestThreadStoreSavesGetsAndFindsMostRecentBySubject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 21, 0, 0, 0, time.UTC)
	thread := domain.Thread{
		ID:                "thread-123",
		OrganizationID:    "org-123",
		AgentID:           "agent-123",
		InboxID:           "inbox-123",
		ContactID:         "contact-123",
		SubjectNormalized: "hello world",
		State:             domain.ThreadStateActive,
		LastInboundID:     "message-1",
		LastOutboundID:    "message-2",
		LastActivityAt:    now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO threads (
			id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
			state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE
		SET organization_id = EXCLUDED.organization_id,
		    agent_id = EXCLUDED.agent_id,
		    inbox_id = EXCLUDED.inbox_id,
		    contact_id = EXCLUDED.contact_id,
		    subject_normalized = EXCLUDED.subject_normalized,
		    state = EXCLUDED.state,
		    last_inbound_id = EXCLUDED.last_inbound_id,
		    last_outbound_id = EXCLUDED.last_outbound_id,
		    last_activity_at = EXCLUDED.last_activity_at,
		    updated_at = EXCLUDED.updated_at
	`)).
		WithArgs(thread.ID, thread.OrganizationID, thread.AgentID, thread.InboxID, thread.ContactID, thread.SubjectNormalized, string(thread.State), thread.LastInboundID, thread.LastOutboundID, thread.LastActivityAt, thread.CreatedAt, thread.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rowSet := sqlmock.NewRows([]string{
		"id", "organization_id", "agent_id", "inbox_id", "contact_id", "subject_normalized",
		"state", "last_inbound_id", "last_outbound_id", "last_activity_at", "created_at", "updated_at",
	}).AddRow(
		thread.ID, thread.OrganizationID, thread.AgentID, thread.InboxID, thread.ContactID, thread.SubjectNormalized,
		string(thread.State), thread.LastInboundID, thread.LastOutboundID, thread.LastActivityAt, thread.CreatedAt, thread.UpdatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE id = $1
	`)).
		WithArgs(thread.ID).
		WillReturnRows(rowSet)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE organization_id = $1 AND inbox_id = $2 AND contact_id = $3 AND subject_normalized = $4
		ORDER BY last_activity_at DESC, id DESC
		LIMIT 1
	`)).
		WithArgs(thread.OrganizationID, thread.InboxID, thread.ContactID, thread.SubjectNormalized).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "agent_id", "inbox_id", "contact_id", "subject_normalized",
			"state", "last_inbound_id", "last_outbound_id", "last_activity_at", "created_at", "updated_at",
		}).AddRow(
			thread.ID, thread.OrganizationID, thread.AgentID, thread.InboxID, thread.ContactID, thread.SubjectNormalized,
			string(thread.State), thread.LastInboundID, thread.LastOutboundID, thread.LastActivityAt, thread.CreatedAt, thread.UpdatedAt,
		))

	store := NewThreadStore(db)
	if _, err := store.Save(context.Background(), thread); err != nil {
		t.Fatalf("expected thread save to succeed, got error: %v", err)
	}
	got, err := store.GetByID(context.Background(), thread.ID)
	if err != nil {
		t.Fatalf("expected thread get to succeed, got error: %v", err)
	}
	if got.SubjectNormalized != thread.SubjectNormalized {
		t.Fatalf("expected thread get result, got %#v", got)
	}
	found, ok, err := store.FindMostRecentBySubject(context.Background(), thread.OrganizationID, thread.InboxID, thread.ContactID, thread.SubjectNormalized)
	if err != nil || !ok {
		t.Fatalf("expected subject find to succeed, got ok=%v err=%v", ok, err)
	}
	if found.ID != thread.ID {
		t.Fatalf("expected found thread, got %#v", found)
	}
}

func TestThreadStoreMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	store := NewThreadStore(db)
	_, err = store.GetByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}

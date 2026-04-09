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

func TestReadStoreLoadsThreadContactAndMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 19, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE id = $1
	`)).
		WithArgs("thread-123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "agent_id", "inbox_id", "contact_id", "subject_normalized",
			"state", "last_inbound_id", "last_outbound_id", "last_activity_at", "created_at", "updated_at",
		}).AddRow(
			"thread-123", "org-123", "agent-123", "inbox-123", "contact-123", "hello world",
			"ACTIVE", "message-1", "message-2", now, now, now,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`)).
		WithArgs("contact-123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "email_address", "display_name", "last_seen_at", "created_at", "updated_at",
		}).AddRow(
			"contact-123", "org-123", "sender@example.com", "Sender Example", now, now, now,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE thread_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`)).
		WithArgs("thread-123", 2).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "thread_id", "inbox_id", "contact_id", "direction", "subject",
			"subject_normalized", "message_id_header", "in_reply_to", "references_headers",
			"text_body", "html_body", "raw_mime_object_key", "provider_message_id",
			"delivery_state", "sent_at", "delivered_at", "bounced_at", "created_at",
		}).AddRow(
			"message-123", "org-123", "thread-123", "inbox-123", "contact-123", "inbound", "Hello World",
			"hello world", "<message-123@example.com>", "", `["<root@example.com>"]`,
			"Plain body.", "<p>HTML body.</p>", "raw/test-message.eml", "", "", nil, nil, nil, now,
		))

	store := NewReadStore(db)

	thread, err := store.GetByID(context.Background(), "thread-123")
	if err != nil {
		t.Fatalf("expected thread lookup to succeed, got error: %v", err)
	}
	if thread.SubjectNormalized != "hello world" {
		t.Fatalf("expected thread result, got %#v", thread)
	}

	contact, err := store.GetContactByID(context.Background(), "contact-123")
	if err != nil {
		t.Fatalf("expected contact lookup to succeed, got error: %v", err)
	}
	if contact.EmailAddress != "sender@example.com" {
		t.Fatalf("expected contact result, got %#v", contact)
	}

	messages, err := store.ListByThreadID(context.Background(), "thread-123", 2)
	if err != nil {
		t.Fatalf("expected message list to succeed, got error: %v", err)
	}
	if len(messages) != 1 || len(messages[0].References) != 1 {
		t.Fatalf("expected message list result, got %#v", messages)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected sqlmock expectations to be met, got error: %v", err)
	}
}

func TestReadStoreMapsNotFound(t *testing.T) {
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

	store := NewReadStore(db)
	_, err = store.GetByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}

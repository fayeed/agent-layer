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

func TestMessageStoreCreatesSavesGetsAndFindsMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 21, 30, 0, 0, time.UTC)
	message := domain.Message{
		ID:                "message-123",
		OrganizationID:    "org-123",
		ThreadID:          "thread-123",
		InboxID:           "inbox-123",
		ContactID:         "contact-123",
		Direction:         domain.MessageDirectionInbound,
		Subject:           "Hello World",
		SubjectNormalized: "hello world",
		MessageIDHeader:   "<message-123@example.com>",
		InReplyTo:         "<root@example.com>",
		References:        []string{"<root@example.com>"},
		TextBody:          "Plain body.",
		HTMLBody:          "<p>HTML body.</p>",
		RawMIMEObjectKey:  "raw/test-message.eml",
		CreatedAt:         now,
	}

	insertSQL := regexp.QuoteMeta(`
		INSERT INTO messages (
			id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
			subject_normalized, message_id_header, in_reply_to, references_headers,
			text_body, html_body, raw_mime_object_key, provider_message_id,
			delivery_state, sent_at, delivered_at, bounced_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`)

	mock.ExpectExec(insertSQL).
		WithArgs(
			message.ID, message.OrganizationID, message.ThreadID, message.InboxID, message.ContactID,
			string(message.Direction), message.Subject, message.SubjectNormalized, message.MessageIDHeader,
			message.InReplyTo, stringArrayValue(message.References), message.TextBody, message.HTMLBody,
			message.RawMIMEObjectKey, message.ProviderMessageID, message.DeliveryState, nil, nil, nil, message.CreatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO messages (
			id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
			subject_normalized, message_id_header, in_reply_to, references_headers,
			text_body, html_body, raw_mime_object_key, provider_message_id,
			delivery_state, sent_at, delivered_at, bounced_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (id) DO UPDATE
		SET organization_id = EXCLUDED.organization_id,
		    thread_id = EXCLUDED.thread_id,
		    inbox_id = EXCLUDED.inbox_id,
		    contact_id = EXCLUDED.contact_id,
		    direction = EXCLUDED.direction,
		    subject = EXCLUDED.subject,
		    subject_normalized = EXCLUDED.subject_normalized,
		    message_id_header = EXCLUDED.message_id_header,
		    in_reply_to = EXCLUDED.in_reply_to,
		    references_headers = EXCLUDED.references_headers,
		    text_body = EXCLUDED.text_body,
		    html_body = EXCLUDED.html_body,
		    raw_mime_object_key = EXCLUDED.raw_mime_object_key,
		    provider_message_id = EXCLUDED.provider_message_id,
		    delivery_state = EXCLUDED.delivery_state,
		    sent_at = EXCLUDED.sent_at,
		    delivered_at = EXCLUDED.delivered_at,
		    bounced_at = EXCLUDED.bounced_at
	`)).
		WithArgs(
			message.ID, message.OrganizationID, message.ThreadID, message.InboxID, message.ContactID,
			string(message.Direction), message.Subject, message.SubjectNormalized, message.MessageIDHeader,
			message.InReplyTo, stringArrayValue(message.References), message.TextBody, message.HTMLBody,
			message.RawMIMEObjectKey, message.ProviderMessageID, message.DeliveryState, nil, nil, nil, message.CreatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	msgRows := sqlmock.NewRows([]string{
		"id", "organization_id", "thread_id", "inbox_id", "contact_id", "direction", "subject",
		"subject_normalized", "message_id_header", "in_reply_to", "references_headers",
		"text_body", "html_body", "raw_mime_object_key", "provider_message_id",
		"delivery_state", "sent_at", "delivered_at", "bounced_at", "created_at",
	}).AddRow(
		message.ID, message.OrganizationID, message.ThreadID, message.InboxID, message.ContactID, string(message.Direction),
		message.Subject, message.SubjectNormalized, message.MessageIDHeader, message.InReplyTo, `["<root@example.com>"]`,
		message.TextBody, message.HTMLBody, message.RawMIMEObjectKey, message.ProviderMessageID, message.DeliveryState,
		nil, nil, nil, message.CreatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE id = $1
	`)).
		WithArgs(message.ID).
		WillReturnRows(msgRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE inbox_id = $1 AND message_id_header = $2
		LIMIT 1
	`)).
		WithArgs(message.InboxID, message.MessageIDHeader).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "thread_id", "inbox_id", "contact_id", "direction", "subject",
			"subject_normalized", "message_id_header", "in_reply_to", "references_headers",
			"text_body", "html_body", "raw_mime_object_key", "provider_message_id",
			"delivery_state", "sent_at", "delivered_at", "bounced_at", "created_at",
		}).AddRow(
			message.ID, message.OrganizationID, message.ThreadID, message.InboxID, message.ContactID, string(message.Direction),
			message.Subject, message.SubjectNormalized, message.MessageIDHeader, message.InReplyTo, `["<root@example.com>"]`,
			message.TextBody, message.HTMLBody, message.RawMIMEObjectKey, message.ProviderMessageID, message.DeliveryState,
			nil, nil, nil, message.CreatedAt,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE provider_message_id = $1
		LIMIT 1
	`)).
		WithArgs("provider-123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "thread_id", "inbox_id", "contact_id", "direction", "subject",
			"subject_normalized", "message_id_header", "in_reply_to", "references_headers",
			"text_body", "html_body", "raw_mime_object_key", "provider_message_id",
			"delivery_state", "sent_at", "delivered_at", "bounced_at", "created_at",
		}).AddRow(
			message.ID, message.OrganizationID, message.ThreadID, message.InboxID, message.ContactID, string(message.Direction),
			message.Subject, message.SubjectNormalized, message.MessageIDHeader, message.InReplyTo, `["<root@example.com>"]`,
			message.TextBody, message.HTMLBody, message.RawMIMEObjectKey, "provider-123", message.DeliveryState,
			nil, nil, nil, message.CreatedAt,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT thread_id
		FROM messages
		WHERE message_id_header = $1
		LIMIT 1
	`)).
		WithArgs(message.MessageIDHeader).
		WillReturnRows(sqlmock.NewRows([]string{"thread_id"}).AddRow(message.ThreadID))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE id = $1
	`)).
		WithArgs(message.ThreadID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "agent_id", "inbox_id", "contact_id", "subject_normalized",
			"state", "last_inbound_id", "last_outbound_id", "last_activity_at", "created_at", "updated_at",
		}).AddRow(
			message.ThreadID, "org-123", "agent-123", message.InboxID, message.ContactID, message.SubjectNormalized,
			"ACTIVE", "", "", message.CreatedAt, message.CreatedAt, message.CreatedAt,
		))

	store := NewMessageStore(db)
	if _, err := store.Create(context.Background(), message); err != nil {
		t.Fatalf("expected message create to succeed, got error: %v", err)
	}
	if _, err := store.Save(context.Background(), message); err != nil {
		t.Fatalf("expected message save to succeed, got error: %v", err)
	}
	got, err := store.GetMessageByID(context.Background(), message.ID)
	if err != nil {
		t.Fatalf("expected message get to succeed, got error: %v", err)
	}
	if got.MessageIDHeader != message.MessageIDHeader || len(got.References) != 1 {
		t.Fatalf("expected message get result, got %#v", got)
	}
	dup, ok, err := store.FindInboundByHeader(context.Background(), message.InboxID, message.MessageIDHeader)
	if err != nil || !ok {
		t.Fatalf("expected inbound-by-header lookup to succeed, got ok=%v err=%v", ok, err)
	}
	if dup.ID != message.ID {
		t.Fatalf("expected inbound-by-header result, got %#v", dup)
	}
	provider, ok, err := store.FindByProviderMessageID(context.Background(), "provider-123")
	if err != nil || !ok {
		t.Fatalf("expected provider lookup to succeed, got ok=%v err=%v", ok, err)
	}
	if provider.ProviderMessageID != "provider-123" {
		t.Fatalf("expected provider lookup result, got %#v", provider)
	}
	thread, ok, err := store.FindByMessageID(context.Background(), message.MessageIDHeader)
	if err != nil || !ok {
		t.Fatalf("expected thread-by-message lookup to succeed, got ok=%v err=%v", ok, err)
	}
	if thread.ID != message.ThreadID {
		t.Fatalf("expected thread-by-message result, got %#v", thread)
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO reply_submissions (submission_key, message_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (submission_key) DO UPDATE
		SET message_id = EXCLUDED.message_id
	`)).
		WithArgs("reply:key", message.ID, message.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT message_id
		FROM reply_submissions
		WHERE submission_key = $1
	`)).
		WithArgs("reply:key").
		WillReturnRows(sqlmock.NewRows([]string{"message_id"}).AddRow(message.ID))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE id = $1
	`)).
		WithArgs(message.ID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "thread_id", "inbox_id", "contact_id", "direction", "subject",
			"subject_normalized", "message_id_header", "in_reply_to", "references_headers",
			"text_body", "html_body", "raw_mime_object_key", "provider_message_id",
			"delivery_state", "sent_at", "delivered_at", "bounced_at", "created_at",
		}).AddRow(
			message.ID, message.OrganizationID, message.ThreadID, message.InboxID, message.ContactID, string(message.Direction),
			message.Subject, message.SubjectNormalized, message.MessageIDHeader, message.InReplyTo, `["<root@example.com>"]`,
			message.TextBody, message.HTMLBody, message.RawMIMEObjectKey, message.ProviderMessageID, message.DeliveryState,
			nil, nil, nil, message.CreatedAt,
		))

	if err := store.SaveReplySubmission(context.Background(), "reply:key", message.ID, message.CreatedAt); err != nil {
		t.Fatalf("expected reply submission save to succeed, got error: %v", err)
	}
	replyMessage, found, err := store.FindReplyBySubmissionKey(context.Background(), "reply:key")
	if err != nil || !found {
		t.Fatalf("expected reply submission lookup to succeed, got found=%v err=%v", found, err)
	}
	if replyMessage.ID != message.ID {
		t.Fatalf("expected reply submission message, got %#v", replyMessage)
	}
}

func TestMessageStoreMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	store := NewMessageStore(db)
	_, err = store.GetMessageByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestInboundReceiptStoreSavesGetsAndListsReceipts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 18, 30, 0, 0, time.UTC)
	receipt := inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-newer",
		OrganizationID:      "org-123",
		AgentID:             "agent-123",
		InboxID:             "inbox-123",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@example.com"},
		RawMessageObjectKey: "raw/newer.eml",
		ReceivedAt:          now,
	}
	recipientsJSON, _ := json.Marshal(receipt.EnvelopeRecipients)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO inbound_receipts (
			raw_message_object_key, smtp_transaction_id, organization_id, agent_id, inbox_id,
			envelope_sender, envelope_recipients, received_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (raw_message_object_key) DO UPDATE
		SET smtp_transaction_id = EXCLUDED.smtp_transaction_id,
		    organization_id = EXCLUDED.organization_id,
		    agent_id = EXCLUDED.agent_id,
		    inbox_id = EXCLUDED.inbox_id,
		    envelope_sender = EXCLUDED.envelope_sender,
		    envelope_recipients = EXCLUDED.envelope_recipients,
		    received_at = EXCLUDED.received_at
	`)).
		WithArgs(
			receipt.RawMessageObjectKey, receipt.SMTPTransactionID, receipt.OrganizationID, receipt.AgentID,
			receipt.InboxID, receipt.EnvelopeSender, stringArrayValue(receipt.EnvelopeRecipients), receipt.ReceivedAt, receipt.ReceivedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT raw_message_object_key, smtp_transaction_id, organization_id, agent_id, inbox_id,
		       envelope_sender, envelope_recipients, received_at, created_at
		FROM inbound_receipts
		WHERE raw_message_object_key = $1
	`)).
		WithArgs(receipt.RawMessageObjectKey).
		WillReturnRows(sqlmock.NewRows([]string{
			"raw_message_object_key", "smtp_transaction_id", "organization_id", "agent_id", "inbox_id",
			"envelope_sender", "envelope_recipients", "received_at", "created_at",
		}).AddRow(
			receipt.RawMessageObjectKey, receipt.SMTPTransactionID, receipt.OrganizationID, receipt.AgentID, receipt.InboxID,
			receipt.EnvelopeSender, recipientsJSON, receipt.ReceivedAt, receipt.ReceivedAt,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT raw_message_object_key, smtp_transaction_id, organization_id, agent_id, inbox_id,
		       envelope_sender, envelope_recipients, received_at, created_at
		FROM inbound_receipts
		ORDER BY received_at DESC, raw_message_object_key DESC
		LIMIT $1
	`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{
			"raw_message_object_key", "smtp_transaction_id", "organization_id", "agent_id", "inbox_id",
			"envelope_sender", "envelope_recipients", "received_at", "created_at",
		}).AddRow(
			receipt.RawMessageObjectKey, receipt.SMTPTransactionID, receipt.OrganizationID, receipt.AgentID, receipt.InboxID,
			receipt.EnvelopeSender, recipientsJSON, receipt.ReceivedAt, receipt.ReceivedAt,
		))

	store := NewInboundReceiptStore(db)

	if err := store.SaveInboundReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("expected receipt save to succeed, got error: %v", err)
	}

	loaded, err := store.GetInboundReceiptByObjectKey(context.Background(), receipt.RawMessageObjectKey)
	if err != nil {
		t.Fatalf("expected receipt lookup to succeed, got error: %v", err)
	}
	if loaded.SMTPTransactionID != receipt.SMTPTransactionID || len(loaded.EnvelopeRecipients) != 1 {
		t.Fatalf("expected receipt lookup result, got %#v", loaded)
	}

	list, err := store.ListInboundReceipts(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected receipt list to succeed, got error: %v", err)
	}
	if len(list) != 1 || list[0].RawMessageObjectKey != receipt.RawMessageObjectKey {
		t.Fatalf("expected listed receipt, got %#v", list)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected sqlmock expectations to be met, got error: %v", err)
	}
}

func TestInboundReceiptStoreMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT raw_message_object_key, smtp_transaction_id, organization_id, agent_id, inbox_id,
		       envelope_sender, envelope_recipients, received_at, created_at
		FROM inbound_receipts
		WHERE raw_message_object_key = $1
	`)).
		WithArgs("raw/missing.eml").
		WillReturnError(sql.ErrNoRows)

	store := NewInboundReceiptStore(db)
	_, err = store.GetInboundReceiptByObjectKey(context.Background(), "raw/missing.eml")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected sqlmock expectations to be met, got error: %v", err)
	}
}

func TestStringArrayHelpersRoundTripJSONStorage(t *testing.T) {
	value, err := pqTextArray([]string{"a@example.com", "b@example.com"}).Value()
	if err != nil {
		t.Fatalf("expected array value conversion to succeed, got error: %v", err)
	}

	var out []string
	if err := pqArrayScan(&out).Scan(value); err != nil {
		t.Fatalf("expected array scan to succeed, got error: %v", err)
	}

	if len(out) != 2 || out[0] != "a@example.com" || out[1] != "b@example.com" {
		t.Fatalf("expected array round trip, got %#v", out)
	}
}

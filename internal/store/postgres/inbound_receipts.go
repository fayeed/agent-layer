package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s InboundReceiptStore) SaveInboundReceipt(ctx context.Context, receipt inbound.DurableReceiptRequest) error {
	model := store.InboundReceiptToModel(receipt)
	_, err := s.db.ExecContext(ctx, `
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
	`, model.RawMessageObjectKey, model.SMTPTransactionID, model.OrganizationID, model.AgentID, model.InboxID, model.EnvelopeSender, pqTextArray(model.EnvelopeRecipients), model.ReceivedAt, model.CreatedAt)
	return err
}

func (s InboundReceiptStore) GetInboundReceiptByObjectKey(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error) {
	var model store.InboundReceiptModel
	var recipients []string
	err := s.db.QueryRowContext(ctx, `
		SELECT raw_message_object_key, smtp_transaction_id, organization_id, agent_id, inbox_id,
		       envelope_sender, envelope_recipients, received_at, created_at
		FROM inbound_receipts
		WHERE raw_message_object_key = $1
	`, objectKey).Scan(
		&model.RawMessageObjectKey,
		&model.SMTPTransactionID,
		&model.OrganizationID,
		&model.AgentID,
		&model.InboxID,
		&model.EnvelopeSender,
		pqArrayScan(&recipients),
		&model.ReceivedAt,
		&model.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return inbound.DurableReceiptRequest{}, domain.ErrNotFound
		}
		return inbound.DurableReceiptRequest{}, err
	}
	model.EnvelopeRecipients = recipients
	return store.InboundReceiptFromModel(model), nil
}

func (s InboundReceiptStore) ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT raw_message_object_key, smtp_transaction_id, organization_id, agent_id, inbox_id,
		       envelope_sender, envelope_recipients, received_at, created_at
		FROM inbound_receipts
		ORDER BY received_at DESC, raw_message_object_key DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []inbound.DurableReceiptRequest
	for rows.Next() {
		var model store.InboundReceiptModel
		var recipients []string
		if err := rows.Scan(
			&model.RawMessageObjectKey,
			&model.SMTPTransactionID,
			&model.OrganizationID,
			&model.AgentID,
			&model.InboxID,
			&model.EnvelopeSender,
			pqArrayScan(&recipients),
			&model.ReceivedAt,
			&model.CreatedAt,
		); err != nil {
			return nil, err
		}
		model.EnvelopeRecipients = recipients
		out = append(out, store.InboundReceiptFromModel(model))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

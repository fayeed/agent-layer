package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s ReadStore) GetByID(ctx context.Context, threadID string) (domain.Thread, error) {
	var model store.ThreadModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE id = $1
	`, threadID).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.AgentID,
		&model.InboxID,
		&model.ContactID,
		&model.SubjectNormalized,
		&model.State,
		&model.LastInboundID,
		&model.LastOutboundID,
		&model.LastActivityAt,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Thread{}, domain.ErrNotFound
		}
		return domain.Thread{}, err
	}
	return store.ThreadFromModel(model), nil
}

func (s ReadStore) ListByThreadID(ctx context.Context, threadID string, limit int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE thread_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, threadID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Message
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s ReadStore) GetContactByID(ctx context.Context, contactID string) (domain.Contact, error) {
	var model store.ContactModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`, contactID).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.EmailAddress,
		&model.DisplayName,
		&model.LastSeenAt,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contact{}, domain.ErrNotFound
		}
		return domain.Contact{}, err
	}
	return store.ContactFromModel(model), nil
}

func scanMessage(scanner interface{ Scan(dest ...any) error }) (domain.Message, error) {
	var model store.MessageModel
	var references []string
	var sentAt sql.NullTime
	var deliveredAt sql.NullTime
	var bouncedAt sql.NullTime
	err := scanner.Scan(
		&model.ID,
		&model.OrganizationID,
		&model.ThreadID,
		&model.InboxID,
		&model.ContactID,
		&model.Direction,
		&model.Subject,
		&model.SubjectNormalized,
		&model.MessageIDHeader,
		&model.InReplyTo,
		pqArrayScan(&references),
		&model.TextBody,
		&model.HTMLBody,
		&model.RawMIMEObjectKey,
		&model.ProviderMessageID,
		&model.DeliveryState,
		&sentAt,
		&deliveredAt,
		&bouncedAt,
		&model.CreatedAt,
	)
	if err != nil {
		return domain.Message{}, err
	}
	model.References = references
	if sentAt.Valid {
		model.SentAt = sentAt.Time
	}
	if deliveredAt.Valid {
		model.DeliveredAt = deliveredAt.Time
	}
	if bouncedAt.Valid {
		model.BouncedAt = bouncedAt.Time
	}
	return store.MessageFromModel(model), nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s MessageStore) Create(ctx context.Context, message domain.Message) (domain.Message, error) {
	model := store.MessageToModel(message)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO messages (
			id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
			subject_normalized, message_id_header, in_reply_to, references_headers,
			text_body, html_body, raw_mime_object_key, provider_message_id,
			delivery_state, sent_at, delivered_at, bounced_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`, model.ID, model.OrganizationID, model.ThreadID, model.InboxID, model.ContactID, model.Direction, model.Subject, model.SubjectNormalized, model.MessageIDHeader, model.InReplyTo, pqTextArray(model.References), model.TextBody, model.HTMLBody, model.RawMIMEObjectKey, model.ProviderMessageID, model.DeliveryState, nullableTime(model.SentAt), nullableTime(model.DeliveredAt), nullableTime(model.BouncedAt), model.CreatedAt)
	if err != nil {
		return domain.Message{}, err
	}
	return message, nil
}

func (s MessageStore) Save(ctx context.Context, message domain.Message) (domain.Message, error) {
	model := store.MessageToModel(message)
	_, err := s.db.ExecContext(ctx, `
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
	`, model.ID, model.OrganizationID, model.ThreadID, model.InboxID, model.ContactID, model.Direction, model.Subject, model.SubjectNormalized, model.MessageIDHeader, model.InReplyTo, pqTextArray(model.References), model.TextBody, model.HTMLBody, model.RawMIMEObjectKey, model.ProviderMessageID, model.DeliveryState, nullableTime(model.SentAt), nullableTime(model.DeliveredAt), nullableTime(model.BouncedAt), model.CreatedAt)
	if err != nil {
		return domain.Message{}, err
	}
	return message, nil
}

func (s MessageStore) GetMessageByID(ctx context.Context, messageID string) (domain.Message, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE id = $1
	`, messageID)
	message, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Message{}, domain.ErrNotFound
		}
		return domain.Message{}, err
	}
	return message, nil
}

func (s MessageStore) FindByMessageID(ctx context.Context, messageIDHeader string) (domain.Thread, bool, error) {
	var threadID string
	err := s.db.QueryRowContext(ctx, `
		SELECT thread_id
		FROM messages
		WHERE message_id_header = $1
		LIMIT 1
	`, messageIDHeader).Scan(&threadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Thread{}, false, nil
		}
		return domain.Thread{}, false, err
	}
	threadStore := NewThreadStore(s.db)
	thread, err := threadStore.GetByID(ctx, threadID)
	if err != nil {
		return domain.Thread{}, false, err
	}
	return thread, true, nil
}

func (s MessageStore) FindInboundByHeader(ctx context.Context, inboxID, messageIDHeader string) (domain.Message, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE inbox_id = $1 AND message_id_header = $2
		LIMIT 1
	`, inboxID, messageIDHeader)
	message, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Message{}, false, nil
		}
		return domain.Message{}, false, err
	}
	return message, true, nil
}

func (s MessageStore) FindByProviderMessageID(ctx context.Context, providerMessageID string) (domain.Message, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, thread_id, inbox_id, contact_id, direction, subject,
		       subject_normalized, message_id_header, in_reply_to, references_headers,
		       text_body, html_body, raw_mime_object_key, provider_message_id,
		       delivery_state, sent_at, delivered_at, bounced_at, created_at
		FROM messages
		WHERE provider_message_id = $1
		LIMIT 1
	`, providerMessageID)
	message, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Message{}, false, nil
		}
		return domain.Message{}, false, err
	}
	return message, true, nil
}

func (s MessageStore) SaveReplySubmission(ctx context.Context, submissionKey, messageID string, createdAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO reply_submissions (submission_key, message_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (submission_key) DO UPDATE
		SET message_id = EXCLUDED.message_id
	`, submissionKey, messageID, createdAt)
	return err
}

func (s MessageStore) FindReplyBySubmissionKey(ctx context.Context, submissionKey string) (domain.Message, bool, error) {
	var messageID string
	err := s.db.QueryRowContext(ctx, `
		SELECT message_id
		FROM reply_submissions
		WHERE submission_key = $1
	`, submissionKey).Scan(&messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Message{}, false, nil
		}
		return domain.Message{}, false, err
	}
	message, err := s.GetMessageByID(ctx, messageID)
	if err != nil {
		return domain.Message{}, false, err
	}
	return message, true, nil
}

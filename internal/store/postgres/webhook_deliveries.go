package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s WebhookDeliveryStore) SaveWebhookDelivery(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	model, err := store.WebhookDeliveryToModel(delivery)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webhook_deliveries (
			id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
			request_headers, status, attempt_count, last_attempt_at, next_attempt_at,
			response_code, response_body, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE
		SET organization_id = EXCLUDED.organization_id,
		    agent_id = EXCLUDED.agent_id,
		    event_type = EXCLUDED.event_type,
		    event_id = EXCLUDED.event_id,
		    request_url = EXCLUDED.request_url,
		    request_payload = EXCLUDED.request_payload,
		    request_headers = EXCLUDED.request_headers,
		    status = EXCLUDED.status,
		    attempt_count = EXCLUDED.attempt_count,
		    last_attempt_at = EXCLUDED.last_attempt_at,
		    next_attempt_at = EXCLUDED.next_attempt_at,
		    response_code = EXCLUDED.response_code,
		    response_body = EXCLUDED.response_body,
		    updated_at = EXCLUDED.updated_at
	`, model.ID, model.OrganizationID, model.AgentID, model.EventType, model.EventID, model.RequestURL, model.RequestPayload, model.RequestHeaders, model.Status, model.AttemptCount, nullableTime(model.LastAttemptAt), nullableTime(model.NextAttemptAt), model.ResponseCode, model.ResponseBody, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}

	return delivery, nil
}

func (s WebhookDeliveryStore) GetWebhookDeliveryByID(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	var model store.WebhookDeliveryModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
		       request_headers, status, attempt_count, last_attempt_at, next_attempt_at, response_code, response_body,
		       created_at, updated_at
		FROM webhook_deliveries
		WHERE id = $1
	`, deliveryID).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.AgentID,
		&model.EventType,
		&model.EventID,
		&model.RequestURL,
		&model.RequestPayload,
		&model.RequestHeaders,
		&model.Status,
		&model.AttemptCount,
		&model.LastAttemptAt,
		&model.NextAttemptAt,
		&model.ResponseCode,
		&model.ResponseBody,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WebhookDelivery{}, domain.ErrNotFound
		}
		return domain.WebhookDelivery{}, err
	}

	return store.WebhookDeliveryFromModel(model)
}

func (s WebhookDeliveryStore) ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
		       request_headers, status, attempt_count, last_attempt_at, next_attempt_at, response_code, response_body,
		       created_at, updated_at
		FROM webhook_deliveries
		ORDER BY updated_at DESC, id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.WebhookDelivery
	for rows.Next() {
		var model store.WebhookDeliveryModel
		if err := rows.Scan(
			&model.ID,
			&model.OrganizationID,
			&model.AgentID,
			&model.EventType,
			&model.EventID,
			&model.RequestURL,
			&model.RequestPayload,
			&model.RequestHeaders,
			&model.Status,
			&model.AttemptCount,
			&model.LastAttemptAt,
			&model.NextAttemptAt,
			&model.ResponseCode,
			&model.ResponseBody,
			&model.CreatedAt,
			&model.UpdatedAt,
		); err != nil {
			return nil, err
		}

		delivery, err := store.WebhookDeliveryFromModel(model)
		if err != nil {
			return nil, err
		}
		out = append(out, delivery)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

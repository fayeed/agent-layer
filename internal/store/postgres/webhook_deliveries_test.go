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

func TestWebhookDeliveryStoreSavesGetsAndListsDeliveries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 19, 30, 0, 0, time.UTC)
	delivery := domain.WebhookDelivery{
		ID:             "delivery-123",
		OrganizationID: "org-123",
		AgentID:        "agent-123",
		EventType:      "message.received",
		EventID:        "event-123",
		RequestURL:     "https://example.com/webhook",
		RequestPayload: []byte(`{"ok":true}`),
		RequestHeaders: map[string]string{"X-Test": "1"},
		Status:         "succeeded",
		AttemptCount:   1,
		ResponseCode:   202,
		ResponseBody:   []byte(`{"accepted":true}`),
		LastAttemptAt:  now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO webhook_deliveries (
			id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
			request_headers, status, attempt_count, last_attempt_at, response_code, response_body,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
		    response_code = EXCLUDED.response_code,
		    response_body = EXCLUDED.response_body,
		    updated_at = EXCLUDED.updated_at
	`)).
		WithArgs(
			delivery.ID, delivery.OrganizationID, delivery.AgentID, delivery.EventType, delivery.EventID,
			delivery.RequestURL, delivery.RequestPayload, sqlmock.AnyArg(), delivery.Status, delivery.AttemptCount,
			delivery.LastAttemptAt, delivery.ResponseCode, delivery.ResponseBody, delivery.CreatedAt, delivery.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{
		"id", "organization_id", "agent_id", "event_type", "event_id", "request_url", "request_payload",
		"request_headers", "status", "attempt_count", "last_attempt_at", "response_code", "response_body",
		"created_at", "updated_at",
	}).AddRow(
		delivery.ID, delivery.OrganizationID, delivery.AgentID, delivery.EventType, delivery.EventID, delivery.RequestURL,
		delivery.RequestPayload, []byte(`{"X-Test":"1"}`), delivery.Status, delivery.AttemptCount, delivery.LastAttemptAt,
		delivery.ResponseCode, delivery.ResponseBody, delivery.CreatedAt, delivery.UpdatedAt,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
		       request_headers, status, attempt_count, last_attempt_at, response_code, response_body,
		       created_at, updated_at
		FROM webhook_deliveries
		WHERE id = $1
	`)).
		WithArgs(delivery.ID).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
		       request_headers, status, attempt_count, last_attempt_at, response_code, response_body,
		       created_at, updated_at
		FROM webhook_deliveries
		ORDER BY updated_at DESC, id DESC
		LIMIT $1
	`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "agent_id", "event_type", "event_id", "request_url", "request_payload",
			"request_headers", "status", "attempt_count", "last_attempt_at", "response_code", "response_body",
			"created_at", "updated_at",
		}).AddRow(
			delivery.ID, delivery.OrganizationID, delivery.AgentID, delivery.EventType, delivery.EventID, delivery.RequestURL,
			delivery.RequestPayload, []byte(`{"X-Test":"1"}`), delivery.Status, delivery.AttemptCount, delivery.LastAttemptAt,
			delivery.ResponseCode, delivery.ResponseBody, delivery.CreatedAt, delivery.UpdatedAt,
		))

	store := NewWebhookDeliveryStore(db)

	if _, err := store.SaveWebhookDelivery(context.Background(), delivery); err != nil {
		t.Fatalf("expected webhook delivery save to succeed, got error: %v", err)
	}

	got, err := store.GetWebhookDeliveryByID(context.Background(), delivery.ID)
	if err != nil {
		t.Fatalf("expected webhook delivery lookup to succeed, got error: %v", err)
	}
	if got.RequestHeaders["X-Test"] != "1" {
		t.Fatalf("expected webhook delivery result, got %#v", got)
	}

	list, err := store.ListWebhookDeliveries(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}
	if len(list) != 1 || list[0].ID != delivery.ID {
		t.Fatalf("expected webhook delivery list result, got %#v", list)
	}
}

func TestWebhookDeliveryStoreMapsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, agent_id, event_type, event_id, request_url, request_payload,
		       request_headers, status, attempt_count, last_attempt_at, response_code, response_body,
		       created_at, updated_at
		FROM webhook_deliveries
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	store := NewWebhookDeliveryStore(db)
	_, err = store.GetWebhookDeliveryByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}

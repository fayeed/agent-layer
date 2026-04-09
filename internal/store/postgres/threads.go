package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s ThreadStore) Save(ctx context.Context, thread domain.Thread) (domain.Thread, error) {
	model := store.ThreadToModel(thread)
	_, err := s.db.ExecContext(ctx, `
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
	`, model.ID, model.OrganizationID, model.AgentID, model.InboxID, model.ContactID, model.SubjectNormalized, model.State, model.LastInboundID, model.LastOutboundID, model.LastActivityAt, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.Thread{}, err
	}
	return thread, nil
}

func (s ThreadStore) GetByID(ctx context.Context, threadID string) (domain.Thread, error) {
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

func (s ThreadStore) FindMostRecentBySubject(ctx context.Context, organizationID, inboxID, contactID, subjectNormalized string) (domain.Thread, bool, error) {
	var model store.ThreadModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, agent_id, inbox_id, contact_id, subject_normalized,
		       state, last_inbound_id, last_outbound_id, last_activity_at, created_at, updated_at
		FROM threads
		WHERE organization_id = $1 AND inbox_id = $2 AND contact_id = $3 AND subject_normalized = $4
		ORDER BY last_activity_at DESC, id DESC
		LIMIT 1
	`, organizationID, inboxID, contactID, subjectNormalized).Scan(
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
			return domain.Thread{}, false, nil
		}
		return domain.Thread{}, false, err
	}
	return store.ThreadFromModel(model), true, nil
}

package postgres

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s ContactMemoryStore) CreateMemory(ctx context.Context, entry domain.ContactMemoryEntry, organizationID string) (domain.ContactMemoryEntry, error) {
	model := store.ContactMemoryToModel(entry, organizationID)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contact_memory (id, organization_id, contact_id, thread_id, note, tags, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, model.ID, model.OrganizationID, model.ContactID, model.ThreadID, model.Note, pqTextArray(model.Tags), model.CreatedAt)
	if err != nil {
		return domain.ContactMemoryEntry{}, err
	}
	return entry, nil
}

func (s ContactMemoryStore) ListMemoryByContactID(ctx context.Context, contactID string, limit int) ([]domain.ContactMemoryEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, contact_id, thread_id, note, tags, created_at
		FROM contact_memory
		WHERE contact_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, contactID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.ContactMemoryEntry
	for rows.Next() {
		var model store.ContactMemoryModel
		var tags []string
		if err := rows.Scan(
			&model.ID,
			&model.OrganizationID,
			&model.ContactID,
			&model.ThreadID,
			&model.Note,
			pqArrayScan(&tags),
			&model.CreatedAt,
		); err != nil {
			return nil, err
		}
		model.Tags = tags
		out = append(out, store.ContactMemoryFromModel(model))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s ContactStore) FindByEmail(ctx context.Context, organizationID, emailAddress string) (domain.Contact, bool, error) {
	var model store.ContactModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		FROM contacts
		WHERE organization_id = $1 AND email_address = $2
	`, organizationID, emailAddress).Scan(
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
			return domain.Contact{}, false, nil
		}
		return domain.Contact{}, false, err
	}
	return store.ContactFromModel(model), true, nil
}

func (s ContactStore) UpsertByEmail(ctx context.Context, contact domain.Contact) (domain.Contact, error) {
	model := store.ContactToModel(contact)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contacts (
			id, organization_id, email_address, display_name, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (organization_id, email_address) DO UPDATE
		SET id = EXCLUDED.id,
		    display_name = EXCLUDED.display_name,
		    last_seen_at = EXCLUDED.last_seen_at,
		    updated_at = EXCLUDED.updated_at
	`, model.ID, model.OrganizationID, model.EmailAddress, model.DisplayName, model.LastSeenAt, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.Contact{}, err
	}
	return contact, nil
}

func (s ContactStore) GetContactByID(ctx context.Context, contactID string) (domain.Contact, error) {
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

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

type SuppressionStore struct {
	db *sql.DB
}

func NewSuppressionStore(db *sql.DB) SuppressionStore {
	return SuppressionStore{db: db}
}

func (s SuppressionStore) Save(ctx context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	model := store.SuppressedAddressToModel(record)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO suppressed_addresses (
			id, organization_id, email_address, reason, source, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (organization_id, email_address) DO UPDATE
		SET id = EXCLUDED.id,
		    reason = EXCLUDED.reason,
		    source = EXCLUDED.source,
		    updated_at = EXCLUDED.updated_at
	`, model.ID, model.OrganizationID, model.EmailAddress, model.Reason, model.Source, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.SuppressedAddress{}, err
	}
	return record, nil
}

func (s SuppressionStore) IsSuppressed(ctx context.Context, organizationID, emailAddress string) (bool, error) {
	var suppressed bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM suppressed_addresses
			WHERE organization_id = $1 AND email_address = $2
		)
	`, organizationID, emailAddress).Scan(&suppressed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return suppressed, nil
}

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

func (s SuppressionStore) GetByID(ctx context.Context, suppressionID string) (domain.SuppressedAddress, error) {
	var model store.SuppressedAddressModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, email_address, reason, source, created_at, updated_at
		FROM suppressed_addresses
		WHERE id = $1
	`, suppressionID).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.EmailAddress,
		&model.Reason,
		&model.Source,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SuppressedAddress{}, domain.ErrNotFound
		}
		return domain.SuppressedAddress{}, err
	}
	return store.SuppressedAddressFromModel(model), nil
}

func (s SuppressionStore) List(ctx context.Context, limit int) ([]domain.SuppressedAddress, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, email_address, reason, source, created_at, updated_at
		FROM suppressed_addresses
		ORDER BY updated_at DESC, id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SuppressedAddress
	for rows.Next() {
		var model store.SuppressedAddressModel
		if err := rows.Scan(
			&model.ID,
			&model.OrganizationID,
			&model.EmailAddress,
			&model.Reason,
			&model.Source,
			&model.CreatedAt,
			&model.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, store.SuppressedAddressFromModel(model))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

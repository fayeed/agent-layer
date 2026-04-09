package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/store"
)

func (s BootstrapStore) SaveOrganization(ctx context.Context, organization domain.Organization) (domain.Organization, error) {
	model := store.OrganizationToModel(organization)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    updated_at = EXCLUDED.updated_at
	`, model.ID, model.Name, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.Organization{}, err
	}
	return organization, nil
}

func (s BootstrapStore) GetOrganizationByID(ctx context.Context, organizationID string) (domain.Organization, error) {
	var model store.OrganizationModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`, organizationID).Scan(
		&model.ID,
		&model.Name,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Organization{}, domain.ErrNotFound
		}
		return domain.Organization{}, err
	}
	return store.OrganizationFromModel(model), nil
}

func (s BootstrapStore) SaveAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error) {
	model := store.AgentToModel(agent)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agents (
			id, organization_id, name, status, webhook_url, webhook_secret, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET organization_id = EXCLUDED.organization_id,
		    name = EXCLUDED.name,
		    status = EXCLUDED.status,
		    webhook_url = EXCLUDED.webhook_url,
		    webhook_secret = EXCLUDED.webhook_secret,
		    updated_at = EXCLUDED.updated_at
	`, model.ID, model.OrganizationID, model.Name, model.Status, model.WebhookURL, model.WebhookSecret, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.Agent{}, err
	}
	return agent, nil
}

func (s BootstrapStore) GetAgentByID(ctx context.Context, agentID string) (domain.Agent, error) {
	var model store.AgentModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, name, status, webhook_url, webhook_secret, created_at, updated_at
		FROM agents
		WHERE id = $1
	`, agentID).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.Name,
		&model.Status,
		&model.WebhookURL,
		&model.WebhookSecret,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Agent{}, domain.ErrNotFound
		}
		return domain.Agent{}, err
	}
	return store.AgentFromModel(model), nil
}

func (s BootstrapStore) SaveInbox(ctx context.Context, inbox domain.Inbox) (domain.Inbox, error) {
	model := store.InboxToModel(inbox)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO inboxes (
			id, organization_id, agent_id, email_address, domain, display_name, outbound_identity, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE
		SET organization_id = EXCLUDED.organization_id,
		    agent_id = EXCLUDED.agent_id,
		    email_address = EXCLUDED.email_address,
		    domain = EXCLUDED.domain,
		    display_name = EXCLUDED.display_name,
		    outbound_identity = EXCLUDED.outbound_identity,
		    updated_at = EXCLUDED.updated_at
	`, model.ID, model.OrganizationID, model.AgentID, model.EmailAddress, model.Domain, model.DisplayName, model.OutboundIdentity, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return domain.Inbox{}, err
	}
	return inbox, nil
}

func (s BootstrapStore) GetInboxByID(ctx context.Context, inboxID string) (domain.Inbox, error) {
	var model store.InboxModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, agent_id, email_address, domain, display_name, outbound_identity, created_at, updated_at
		FROM inboxes
		WHERE id = $1
	`, inboxID).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.AgentID,
		&model.EmailAddress,
		&model.Domain,
		&model.DisplayName,
		&model.OutboundIdentity,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Inbox{}, domain.ErrNotFound
		}
		return domain.Inbox{}, err
	}
	return store.InboxFromModel(model), nil
}

func (s BootstrapStore) FindByEmailAddress(ctx context.Context, emailAddress string) (domain.Inbox, bool, error) {
	var model store.InboxModel
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, agent_id, email_address, domain, display_name, outbound_identity, created_at, updated_at
		FROM inboxes
		WHERE email_address = $1
	`, emailAddress).Scan(
		&model.ID,
		&model.OrganizationID,
		&model.AgentID,
		&model.EmailAddress,
		&model.Domain,
		&model.DisplayName,
		&model.OutboundIdentity,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Inbox{}, false, nil
		}
		return domain.Inbox{}, false, err
	}
	return store.InboxFromModel(model), true, nil
}

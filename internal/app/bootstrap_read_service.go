package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type OrganizationGetter interface {
	GetOrganizationByID(ctx context.Context, organizationID string) (domain.Organization, error)
}

type AgentGetter interface {
	GetAgentByID(ctx context.Context, agentID string) (domain.Agent, error)
}

type InboxGetter interface {
	GetInboxByID(ctx context.Context, inboxID string) (domain.Inbox, error)
}

type BootstrapReadResult struct {
	Organization domain.Organization
	Agent        domain.Agent
	Inbox        domain.Inbox
}

type BootstrapReadService struct {
	organizations OrganizationGetter
	agents        AgentGetter
	inboxes       InboxGetter
}

func NewBootstrapReadService(
	organizations OrganizationGetter,
	agents AgentGetter,
	inboxes InboxGetter,
) BootstrapReadService {
	return BootstrapReadService{
		organizations: organizations,
		agents:        agents,
		inboxes:       inboxes,
	}
}

func (s BootstrapReadService) GetBootstrap(ctx context.Context) (BootstrapReadResult, error) {
	organization, err := s.organizations.GetOrganizationByID(ctx, "org-local")
	if err != nil {
		return BootstrapReadResult{}, err
	}

	agent, err := s.agents.GetAgentByID(ctx, "agent-local")
	if err != nil {
		return BootstrapReadResult{}, err
	}

	inbox, err := s.inboxes.GetInboxByID(ctx, "inbox-local")
	if err != nil {
		return BootstrapReadResult{}, err
	}

	return BootstrapReadResult{
		Organization: organization,
		Agent:        agent,
		Inbox:        inbox,
	}, nil
}

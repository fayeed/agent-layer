package app

import (
	"context"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type OrganizationWriter interface {
	SaveOrganization(ctx context.Context, organization domain.Organization) (domain.Organization, error)
}

type AgentWriter interface {
	SaveAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error)
}

type InboxWriter interface {
	SaveInbox(ctx context.Context, inbox domain.Inbox) (domain.Inbox, error)
}

type BootstrapInput struct {
	OrganizationName string
	AgentName        string
	AgentStatus      domain.AgentStatus
	WebhookURL       string
	WebhookSecret    string
	InboxAddress     string
	InboxDomain      string
	InboxDisplayName string
}

type BootstrapResult struct {
	Organization domain.Organization
	Agent        domain.Agent
	Inbox        domain.Inbox
}

type BootstrapService struct {
	organizations OrganizationWriter
	agents        AgentWriter
	inboxes       InboxWriter
	now           func() time.Time
}

func NewBootstrapService(
	organizations OrganizationWriter,
	agents AgentWriter,
	inboxes InboxWriter,
	now func() time.Time,
) BootstrapService {
	if now == nil {
		now = time.Now
	}
	return BootstrapService{
		organizations: organizations,
		agents:        agents,
		inboxes:       inboxes,
		now:           now,
	}
}

func (s BootstrapService) BootstrapLocal(ctx context.Context, input BootstrapInput) (BootstrapResult, error) {
	now := s.now().UTC()
	status := input.AgentStatus
	if status == "" {
		status = domain.AgentStatusActive
	}

	organization, err := s.organizations.SaveOrganization(ctx, domain.Organization{
		ID:        "org-local",
		Name:      input.OrganizationName,
		UpdatedAt: now,
	})
	if err != nil {
		return BootstrapResult{}, err
	}

	agent, err := s.agents.SaveAgent(ctx, domain.Agent{
		ID:             "agent-local",
		OrganizationID: organization.ID,
		Name:           input.AgentName,
		Status:         status,
		WebhookURL:     input.WebhookURL,
		WebhookSecret:  input.WebhookSecret,
		UpdatedAt:      now,
	})
	if err != nil {
		return BootstrapResult{}, err
	}

	inbox, err := s.inboxes.SaveInbox(ctx, domain.Inbox{
		ID:             "inbox-local",
		OrganizationID: organization.ID,
		AgentID:        agent.ID,
		EmailAddress:   input.InboxAddress,
		Domain:         input.InboxDomain,
		DisplayName:    input.InboxDisplayName,
		UpdatedAt:      now,
	})
	if err != nil {
		return BootstrapResult{}, err
	}

	return BootstrapResult{
		Organization: organization,
		Agent:        agent,
		Inbox:        inbox,
	}, nil
}

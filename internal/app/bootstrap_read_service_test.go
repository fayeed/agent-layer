package app

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestBootstrapReadServiceLoadsLocalRuntimeConfig(t *testing.T) {
	store := &bootstrapReadStoreStub{
		organization: domain.Organization{ID: "org-local", Name: "Acme Support"},
		agent: domain.Agent{
			ID:             "agent-local",
			OrganizationID: "org-local",
			Name:           "Acme Agent",
			Status:         domain.AgentStatusActive,
			WebhookURL:     "https://example.com/webhook",
		},
		inbox: domain.Inbox{
			ID:             "inbox-local",
			OrganizationID: "org-local",
			AgentID:        "agent-local",
			EmailAddress:   "agent@example.com",
		},
	}

	service := NewBootstrapReadService(store, store, store)
	result, err := service.GetBootstrap(context.Background())
	if err != nil {
		t.Fatalf("expected bootstrap read to succeed, got error: %v", err)
	}

	if result.Organization.ID != "org-local" || result.Agent.ID != "agent-local" || result.Inbox.ID != "inbox-local" {
		t.Fatalf("expected loaded local runtime config, got %#v", result)
	}
}

type bootstrapReadStoreStub struct {
	organization domain.Organization
	agent        domain.Agent
	inbox        domain.Inbox
}

func (s *bootstrapReadStoreStub) GetOrganizationByID(context.Context, string) (domain.Organization, error) {
	return s.organization, nil
}

func (s *bootstrapReadStoreStub) GetAgentByID(context.Context, string) (domain.Agent, error) {
	return s.agent, nil
}

func (s *bootstrapReadStoreStub) GetInboxByID(context.Context, string) (domain.Inbox, error) {
	return s.inbox, nil
}

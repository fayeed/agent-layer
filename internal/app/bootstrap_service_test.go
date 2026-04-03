package app

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestBootstrapServiceSavesLocalRuntimeConfig(t *testing.T) {
	now := time.Date(2026, 4, 3, 22, 0, 0, 0, time.UTC)
	store := &bootstrapStoreStub{}
	service := NewBootstrapService(store, store, store, func() time.Time { return now })

	result, err := service.BootstrapLocal(context.Background(), BootstrapInput{
		OrganizationName: "Acme Support",
		AgentName:        "Acme Agent",
		AgentStatus:      domain.AgentStatusPaused,
		WebhookURL:       "https://example.com/webhook",
		WebhookSecret:    "super-secret",
		InboxAddress:     "agent@example.com",
		InboxDomain:      "example.com",
		InboxDisplayName: "Acme Inbox",
	})
	if err != nil {
		t.Fatalf("expected bootstrap to succeed, got error: %v", err)
	}

	if result.Organization.ID != "org-local" || result.Agent.ID != "agent-local" || result.Inbox.ID != "inbox-local" {
		t.Fatalf("expected stable local ids, got %#v", result)
	}

	if store.agent.WebhookURL != "https://example.com/webhook" {
		t.Fatalf("expected webhook config to be saved, got %#v", store.agent)
	}

	if store.inbox.EmailAddress != "agent@example.com" {
		t.Fatalf("expected inbox to be saved, got %#v", store.inbox)
	}
}

func TestBootstrapServiceDefaultsAgentStatusToActive(t *testing.T) {
	store := &bootstrapStoreStub{}
	service := NewBootstrapService(store, store, store, time.Now)

	_, err := service.BootstrapLocal(context.Background(), BootstrapInput{})
	if err != nil {
		t.Fatalf("expected bootstrap to succeed, got error: %v", err)
	}

	if store.agent.Status != domain.AgentStatusActive {
		t.Fatalf("expected default agent status to be active, got %#v", store.agent)
	}
}

type bootstrapStoreStub struct {
	organization domain.Organization
	agent        domain.Agent
	inbox        domain.Inbox
}

func (s *bootstrapStoreStub) SaveOrganization(_ context.Context, organization domain.Organization) (domain.Organization, error) {
	s.organization = organization
	return organization, nil
}

func (s *bootstrapStoreStub) SaveAgent(_ context.Context, agent domain.Agent) (domain.Agent, error) {
	s.agent = agent
	return agent, nil
}

func (s *bootstrapStoreStub) SaveInbox(_ context.Context, inbox domain.Inbox) (domain.Inbox, error) {
	s.inbox = inbox
	return inbox, nil
}

package threading

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestResolverMatchesThreadByInReplyTo(t *testing.T) {
	repository := threadLookupRepositoryStub{
		byMessageID: map[string]domain.Thread{
			"<outbound-123@example.com>": {
				ID:             "thread-123",
				OrganizationID: "org-123",
				AgentID:        "agent-123",
				InboxID:        "inbox-123",
				ContactID:      "contact-123",
				State:          domain.ThreadStateActive,
			},
		},
	}

	resolver := NewResolver(repository)

	result, err := resolver.Resolve(context.Background(), core.ThreadResolutionInput{
		OrganizationID: "org-123",
		AgentID:        "agent-123",
		InboxID:        "inbox-123",
		ContactID:      "contact-123",
		ParsedMessage: core.ParsedMessage{
			InReplyTo: "<outbound-123@example.com>",
			Subject:   "Re: Hello World",
		},
		ReceivedAt: time.Date(2026, 4, 2, 21, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed, got error: %v", err)
	}

	if result.Thread.ID != "thread-123" {
		t.Fatalf("expected thread match by in-reply-to, got %q", result.Thread.ID)
	}

	if result.MatchedBy != MatchStrategyInReplyTo {
		t.Fatalf("expected match strategy %q, got %q", MatchStrategyInReplyTo, result.MatchedBy)
	}

	if result.Created {
		t.Fatal("expected existing thread match, not thread creation")
	}
}

func TestResolverFallsBackToReferences(t *testing.T) {
	repository := threadLookupRepositoryStub{
		byMessageID: map[string]domain.Thread{
			"<older-456@example.com>": {
				ID:             "thread-456",
				OrganizationID: "org-123",
				AgentID:        "agent-123",
				InboxID:        "inbox-123",
				ContactID:      "contact-123",
				State:          domain.ThreadStateActive,
			},
		},
	}

	resolver := NewResolver(repository)

	result, err := resolver.Resolve(context.Background(), core.ThreadResolutionInput{
		OrganizationID: "org-123",
		AgentID:        "agent-123",
		InboxID:        "inbox-123",
		ContactID:      "contact-123",
		ParsedMessage: core.ParsedMessage{
			References: []string{"<missing@example.com>", "<older-456@example.com>"},
			Subject:    "Re: Hello Again",
		},
		ReceivedAt: time.Date(2026, 4, 2, 21, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed, got error: %v", err)
	}

	if result.Thread.ID != "thread-456" {
		t.Fatalf("expected thread match by references, got %q", result.Thread.ID)
	}

	if result.MatchedBy != MatchStrategyReferences {
		t.Fatalf("expected match strategy %q, got %q", MatchStrategyReferences, result.MatchedBy)
	}

	if result.Created {
		t.Fatal("expected existing thread match, not thread creation")
	}
}

func TestResolverCreatesThreadWhenNoHeaderMatchExists(t *testing.T) {
	resolver := NewResolver(threadLookupRepositoryStub{})

	receivedAt := time.Date(2026, 4, 2, 21, 0, 0, 0, time.UTC)
	result, err := resolver.Resolve(context.Background(), core.ThreadResolutionInput{
		OrganizationID: "org-123",
		AgentID:        "agent-123",
		InboxID:        "inbox-123",
		ContactID:      "contact-123",
		ParsedMessage: core.ParsedMessage{
			Subject:           "Re: Hello New Thread",
			SubjectNormalized: "hello new thread",
		},
		ReceivedAt: receivedAt,
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed, got error: %v", err)
	}

	if !result.Created {
		t.Fatal("expected a new thread to be created")
	}

	if result.MatchedBy != MatchStrategyNewThread {
		t.Fatalf("expected match strategy %q, got %q", MatchStrategyNewThread, result.MatchedBy)
	}

	if result.Thread.OrganizationID != "org-123" || result.Thread.AgentID != "agent-123" {
		t.Fatalf("expected new thread ownership fields to be set, got %#v", result.Thread)
	}

	if result.Thread.SubjectNormalized != "hello new thread" {
		t.Fatalf("expected normalized subject to be carried onto new thread, got %q", result.Thread.SubjectNormalized)
	}

	if !result.Thread.LastActivityAt.Equal(receivedAt) {
		t.Fatalf("expected last activity time %v, got %v", receivedAt, result.Thread.LastActivityAt)
	}
}

type threadLookupRepositoryStub struct {
	byMessageID map[string]domain.Thread
}

func (s threadLookupRepositoryStub) FindByMessageID(_ context.Context, messageID string) (domain.Thread, bool, error) {
	thread, ok := s.byMessageID[messageID]
	return thread, ok, nil
}

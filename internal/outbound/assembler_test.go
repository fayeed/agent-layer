package outbound

import (
	"strings"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestAssemblerBuildsThreadedReplyMIME(t *testing.T) {
	assembler := NewAssembler(func() string {
		return "<reply-123@agentlayer.local>"
	})

	raw, metadata, err := assembler.AssembleReply(ReplyAssemblyInput{
		Organization: domain.Organization{
			ID: "org-123",
		},
		Agent: domain.Agent{
			ID:   "agent-123",
			Name: "Support Agent",
		},
		Inbox: domain.Inbox{
			ID:           "inbox-123",
			EmailAddress: "agent@example.com",
			DisplayName:  "Agent Layer",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		ReplyToMessage: domain.Message{
			ID:              "message-100",
			ThreadID:        "thread-123",
			MessageIDHeader: "<message-100@example.com>",
			References:      []string{"<message-001@example.com>", "<message-050@example.com>"},
			Subject:         "Hello World",
			CreatedAt:       time.Date(2026, 4, 3, 6, 0, 0, 0, time.UTC),
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
			DisplayName:  "Sender Example",
		},
		BodyText: "Thanks for reaching out.",
	})
	if err != nil {
		t.Fatalf("expected assembly to succeed, got error: %v", err)
	}

	if metadata.MessageIDHeader != "<reply-123@agentlayer.local>" {
		t.Fatalf("expected generated message id header, got %q", metadata.MessageIDHeader)
	}

	if metadata.InReplyTo != "<message-100@example.com>" {
		t.Fatalf("expected in-reply-to to target replied message, got %q", metadata.InReplyTo)
	}

	expectedReferences := []string{"<message-001@example.com>", "<message-050@example.com>", "<message-100@example.com>"}
	if len(metadata.References) != len(expectedReferences) {
		t.Fatalf("expected references to be extended, got %#v", metadata.References)
	}

	for i, ref := range expectedReferences {
		if metadata.References[i] != ref {
			t.Fatalf("expected reference %d to be %q, got %#v", i, ref, metadata.References)
		}
	}

	if !strings.Contains(raw, "From: Agent Layer <agent@example.com>\r\n") {
		t.Fatalf("expected from header in mime, got %q", raw)
	}

	if !strings.Contains(raw, "To: Sender Example <sender@example.com>\r\n") {
		t.Fatalf("expected to header in mime, got %q", raw)
	}

	if !strings.Contains(raw, "Subject: Re: Hello World\r\n") {
		t.Fatalf("expected reply subject in mime, got %q", raw)
	}

	if !strings.Contains(raw, "In-Reply-To: <message-100@example.com>\r\n") {
		t.Fatalf("expected in-reply-to header in mime, got %q", raw)
	}

	if !strings.Contains(raw, "References: <message-001@example.com> <message-050@example.com> <message-100@example.com>\r\n") {
		t.Fatalf("expected references header in mime, got %q", raw)
	}

	if !strings.HasSuffix(raw, "\r\n\r\nThanks for reaching out.") {
		t.Fatalf("expected plain text body in mime, got %q", raw)
	}
}

func TestAssemblerPrefixesSubjectWhenNeeded(t *testing.T) {
	assembler := NewAssembler(func() string {
		return "<reply-456@agentlayer.local>"
	})

	_, metadata, err := assembler.AssembleReply(ReplyAssemblyInput{
		Inbox: domain.Inbox{
			EmailAddress: "agent@example.com",
		},
		Contact: domain.Contact{
			EmailAddress: "sender@example.com",
		},
		ReplyToMessage: domain.Message{
			MessageIDHeader: "<message-200@example.com>",
			Subject:         "Re: Already Replied",
		},
		BodyText: "Following up.",
	})
	if err != nil {
		t.Fatalf("expected assembly to succeed, got error: %v", err)
	}

	if metadata.Subject != "Re: Already Replied" {
		t.Fatalf("expected subject to avoid double prefix, got %q", metadata.Subject)
	}
}

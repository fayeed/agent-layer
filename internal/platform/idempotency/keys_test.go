package idempotency

import "testing"

func TestInboundMessageKeyIsDeterministic(t *testing.T) {
	first := InboundMessageKey("inbox-1", "message-id-123")
	second := InboundMessageKey("inbox-1", "message-id-123")

	if first != second {
		t.Fatalf("expected deterministic inbound key, got %q and %q", first, second)
	}

	if first == InboundMessageKey("inbox-2", "message-id-123") {
		t.Fatal("expected different inboxes to produce different inbound keys")
	}
}

func TestReplySubmissionKeyIsDeterministic(t *testing.T) {
	first := ReplySubmissionKey("thread-1", "request-1")
	second := ReplySubmissionKey("thread-1", "request-1")

	if first != second {
		t.Fatalf("expected deterministic reply key, got %q and %q", first, second)
	}

	if first == ReplySubmissionKey("thread-1", "request-2") {
		t.Fatal("expected different request IDs to produce different reply keys")
	}
}

package smtpedge

import (
	"strings"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestNewSessionIDIncludesTimestampAndPrefix(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 34, 56, 123456789, time.UTC)

	sessionID := NewSessionID(now)

	if !strings.HasPrefix(sessionID, "smtp-20260409T123456.123456789Z-") {
		t.Fatalf("expected timestamped smtp session id, got %q", sessionID)
	}
}

func TestNewRawMessageObjectKeyIncludesDateAndInbox(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 34, 56, 0, time.UTC)

	objectKey := NewRawMessageObjectKey(now, domain.Inbox{ID: "Inbox Local/Primary"})

	if !strings.HasPrefix(objectKey, "raw/2026/04/09/inbox-local-primary/") {
		t.Fatalf("expected dated inbox-scoped object key, got %q", objectKey)
	}

	if !strings.HasSuffix(objectKey, ".eml") {
		t.Fatalf("expected eml object key suffix, got %q", objectKey)
	}
}

func TestSanitizePathSegmentNormalizesUnsafeCharacters(t *testing.T) {
	got := sanitizePathSegment(" Inbox Local/Primary ")

	if got != "inbox-local-primary" {
		t.Fatalf("expected sanitized path segment, got %q", got)
	}
}

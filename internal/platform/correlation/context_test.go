package correlation

import (
	"context"
	"testing"
)

func TestWithIDAndID(t *testing.T) {
	ctx := context.Background()

	if _, ok := ID(ctx); ok {
		t.Fatal("expected empty context to have no correlation ID")
	}

	ctx = WithID(ctx, "corr-123")

	got, ok := ID(ctx)
	if !ok {
		t.Fatal("expected correlation ID to be present")
	}

	if got != "corr-123" {
		t.Fatalf("expected correlation ID %q, got %q", "corr-123", got)
	}
}

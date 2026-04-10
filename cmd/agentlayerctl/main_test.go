package main

import (
	"os"
	"testing"
)

func TestEnvHelpers(t *testing.T) {
	t.Setenv("AGENTLAYER_BASE_URL", "http://localhost:9999")
	t.Setenv("AGENTLAYER_WEBHOOK_LIMIT", "7")
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_LIMIT", "9")

	if got := envOrDefault("AGENTLAYER_BASE_URL", "http://localhost:8080"); got != "http://localhost:9999" {
		t.Fatalf("expected env override, got %q", got)
	}

	if got := envIntOrDefault("AGENTLAYER_WEBHOOK_LIMIT", 5); got != 7 {
		t.Fatalf("expected parsed int env, got %d", got)
	}
	if got := envIntOrDefault("AGENTLAYER_WEBHOOK_RETRY_LIMIT", 20); got != 9 {
		t.Fatalf("expected parsed retry int env, got %d", got)
	}

	if got := envIntOrDefault("AGENTLAYER_UNKNOWN_LIMIT", 5); got != 5 {
		t.Fatalf("expected fallback int, got %d", got)
	}
}

func TestUsageIsCallable(t *testing.T) {
	previous := os.Stderr
	defer func() { os.Stderr = previous }()
	usage()
}

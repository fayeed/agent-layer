package dev

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
)

func TestEmailProviderSendReturnsAcceptedResult(t *testing.T) {
	now := time.Date(2026, 4, 3, 19, 0, 0, 0, time.UTC)
	provider := NewEmailProvider(func() time.Time { return now })

	result, err := provider.Send(context.Background(), core.OutboundSendRequest{})
	if err != nil {
		t.Fatalf("expected send to succeed, got error: %v", err)
	}

	if result.ProviderMessageID == "" {
		t.Fatalf("expected generated provider message id, got %#v", result)
	}

	if !result.AcceptedAt.Equal(now) {
		t.Fatalf("expected accepted time %v, got %#v", now, result)
	}
}

func TestEmailProviderHealthCheckIsHealthy(t *testing.T) {
	now := time.Date(2026, 4, 3, 19, 0, 0, 0, time.UTC)
	provider := NewEmailProvider(func() time.Time { return now })

	health, err := provider.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("expected health check to succeed, got error: %v", err)
	}

	if !health.Healthy || health.ProviderName != "dev" {
		t.Fatalf("expected healthy dev provider, got %#v", health)
	}
}

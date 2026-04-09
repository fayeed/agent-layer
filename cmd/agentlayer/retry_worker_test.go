package main

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/webhooks"
)

func TestWebhookRetryHelpers(t *testing.T) {
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_ENABLED", "")
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_INTERVAL", "")
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_LIMIT", "")

	if !webhookRetryEnabled() {
		t.Fatal("expected webhook retry worker to be enabled by default")
	}
	if got := webhookRetryInterval(); got != 30*time.Second {
		t.Fatalf("expected default retry interval, got %v", got)
	}
	if got := webhookRetryLimit(); got != 20 {
		t.Fatalf("expected default retry limit, got %d", got)
	}
}

func TestWebhookRetryHelpersReadEnv(t *testing.T) {
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_ENABLED", "false")
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_INTERVAL", "5s")
	t.Setenv("AGENTLAYER_WEBHOOK_RETRY_LIMIT", "7")

	if webhookRetryEnabled() {
		t.Fatal("expected webhook retry worker to be disabled")
	}
	if got := webhookRetryInterval(); got != 5*time.Second {
		t.Fatalf("expected retry interval from env, got %v", got)
	}
	if got := webhookRetryLimit(); got != 7 {
		t.Fatalf("expected retry limit from env, got %d", got)
	}
}

func TestWebhookRetryWorkerRunsImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := make(chan int, 2)
	done := make(chan struct{})
	go func() {
		defer close(done)
		custom := retryWorkerServiceStub{calls: calls}
		webhookRetryWorker{
			service:  custom.service(),
			interval: time.Hour,
			limit:    3,
		}.Run(ctx)
	}()

	select {
	case <-calls:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("expected retry worker to run immediately")
	}

	<-done
}

type retryWorkerServiceStub struct {
	calls chan int
}

func (s retryWorkerServiceStub) service() webhooks.RetrySweepService {
	return webhooks.NewRetrySweepService(
		retryWorkerListerStub{calls: s.calls},
		&retryWorkerReplayerStub{},
		func() time.Time { return time.Now().UTC() },
	)
}

type retryWorkerListerStub struct {
	calls chan int
}

func (s retryWorkerListerStub) ListWebhookDeliveries(context.Context, int) ([]domain.WebhookDelivery, error) {
	if s.calls != nil {
		s.calls <- 1
	}
	return nil, nil
}

type retryWorkerReplayerStub struct{}

func (s *retryWorkerReplayerStub) ReplayDelivery(context.Context, string) (webhooks.DeliveryResult, error) {
	return webhooks.DeliveryResult{}, nil
}

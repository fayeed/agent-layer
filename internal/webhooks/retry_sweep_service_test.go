package webhooks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestRetrySweepServiceReplaysDueRetryingDeliveries(t *testing.T) {
	now := time.Date(2026, 4, 9, 20, 0, 0, 0, time.UTC)
	replayer := &retrySweepReplayerStub{
		results: map[string]DeliveryResult{
			"delivery-due": {
				Delivery: domain.WebhookDelivery{ID: "delivery-due", Status: DeliveryStatusSucceeded, AttemptCount: 2},
			},
		},
	}

	service := NewRetrySweepService(
		retrySweepListerStub{
			deliveries: []domain.WebhookDelivery{
				{ID: "delivery-due", Status: DeliveryStatusRetrying, NextAttemptAt: now.Add(-1 * time.Minute)},
				{ID: "delivery-later", Status: DeliveryStatusRetrying, NextAttemptAt: now.Add(2 * time.Minute)},
				{ID: "delivery-dead", Status: DeliveryStatusDeadLetter, NextAttemptAt: now.Add(-1 * time.Minute)},
			},
		},
		replayer,
		func() time.Time { return now },
	)

	result, err := service.RetryDueDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected retry sweep to succeed, got error: %v", err)
	}

	if result.Attempted != 1 || result.Succeeded != 1 || result.Failed != 0 || result.Skipped != 2 {
		t.Fatalf("expected retry sweep counts, got %#v", result)
	}
	if len(replayer.ids) != 1 || replayer.ids[0] != "delivery-due" {
		t.Fatalf("expected only due delivery to be replayed, got %#v", replayer.ids)
	}
}

func TestRetrySweepServiceContinuesAfterReplayFailure(t *testing.T) {
	now := time.Date(2026, 4, 9, 20, 0, 0, 0, time.UTC)
	replayer := &retrySweepReplayerStub{
		results: map[string]DeliveryResult{
			"delivery-ok": {
				Delivery: domain.WebhookDelivery{ID: "delivery-ok", Status: DeliveryStatusSucceeded, AttemptCount: 2},
			},
		},
		errors: map[string]error{
			"delivery-fail": errors.New("boom"),
		},
	}

	service := NewRetrySweepService(
		retrySweepListerStub{
			deliveries: []domain.WebhookDelivery{
				{ID: "delivery-fail", Status: DeliveryStatusRetrying, NextAttemptAt: now.Add(-1 * time.Minute)},
				{ID: "delivery-ok", Status: DeliveryStatusRetrying, NextAttemptAt: now.Add(-1 * time.Minute)},
			},
		},
		replayer,
		func() time.Time { return now },
	)

	result, err := service.RetryDueDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected retry sweep to succeed, got error: %v", err)
	}

	if result.Attempted != 2 || result.Succeeded != 1 || result.Failed != 1 || result.Skipped != 0 {
		t.Fatalf("expected retry sweep counts, got %#v", result)
	}
	if len(result.Deliveries) != 1 || result.Deliveries[0].ID != "delivery-ok" {
		t.Fatalf("expected successful replay result, got %#v", result.Deliveries)
	}
}

type retrySweepListerStub struct {
	deliveries []domain.WebhookDelivery
	err        error
}

func (s retrySweepListerStub) ListWebhookDeliveries(context.Context, int) ([]domain.WebhookDelivery, error) {
	return s.deliveries, s.err
}

type retrySweepReplayerStub struct {
	results map[string]DeliveryResult
	errors  map[string]error
	ids     []string
}

func (s *retrySweepReplayerStub) ReplayDelivery(_ context.Context, deliveryID string) (DeliveryResult, error) {
	s.ids = append(s.ids, deliveryID)
	if err := s.errors[deliveryID]; err != nil {
		return DeliveryResult{}, err
	}
	return s.results[deliveryID], nil
}

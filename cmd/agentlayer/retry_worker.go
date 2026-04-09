package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/agentlayer/agentlayer/internal/webhooks"
)

type backgroundWorker interface {
	Run(ctx context.Context)
}

type webhookRetryWorker struct {
	service  webhooks.RetrySweepService
	interval time.Duration
	limit    int
}

func newWebhookRetryWorker() backgroundWorker {
	if !webhookRetryEnabled() {
		return nil
	}
	return webhookRetryWorker{
		service:  newWebhookRetrySweepService(),
		interval: webhookRetryInterval(),
		limit:    webhookRetryLimit(),
	}
}

func (w webhookRetryWorker) Run(ctx context.Context) {
	if w.interval <= 0 {
		return
	}

	w.runOnce(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w webhookRetryWorker) runOnce(ctx context.Context) {
	result, err := w.service.RetryDueDeliveries(ctx, w.limit)
	if err != nil {
		log.Printf("agentlayer webhook retry worker failed: %v", err)
		return
	}
	if result.Attempted > 0 || result.Failed > 0 {
		log.Printf(
			"agentlayer webhook retry worker attempted=%d succeeded=%d failed=%d skipped=%d",
			result.Attempted,
			result.Succeeded,
			result.Failed,
			result.Skipped,
		)
	}
}

func webhookRetryEnabled() bool {
	value := os.Getenv("AGENTLAYER_WEBHOOK_RETRY_ENABLED")
	if value == "" {
		return true
	}
	return value == "1" || value == "true" || value == "TRUE"
}

func webhookRetryInterval() time.Duration {
	value := os.Getenv("AGENTLAYER_WEBHOOK_RETRY_INTERVAL")
	if value == "" {
		return 30 * time.Second
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 30 * time.Second
	}
	return duration
}

func webhookRetryLimit() int {
	value := os.Getenv("AGENTLAYER_WEBHOOK_RETRY_LIMIT")
	if value == "" {
		return 20
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return 20
	}
	return limit
}

package dev

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
)

type Clock func() time.Time

type EmailProvider struct {
	now Clock
}

func NewEmailProvider(now Clock) EmailProvider {
	if now == nil {
		now = time.Now
	}
	return EmailProvider{now: now}
}

func (p EmailProvider) Send(_ context.Context, _ core.OutboundSendRequest) (core.SendResult, error) {
	return core.SendResult{
		ProviderMessageID: "dev-" + randomID(),
		AcceptedAt:        p.now().UTC(),
	}, nil
}

func (p EmailProvider) GetDeliveryStatus(_ context.Context, providerMessageID string) (core.DeliveryStatus, error) {
	return core.DeliveryStatus{
		ProviderMessageID: providerMessageID,
		State:             "accepted",
		UpdatedAt:         p.now().UTC(),
	}, nil
}

func (p EmailProvider) HealthCheck(_ context.Context) (core.ProviderHealth, error) {
	return core.ProviderHealth{
		ProviderName: "dev",
		Healthy:      true,
		CheckedAt:    p.now().UTC(),
		Details:      "local development provider",
	}, nil
}

func randomID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "generated"
	}
	return hex.EncodeToString(buf[:])
}

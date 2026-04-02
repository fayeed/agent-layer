package outbound

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type EmailProvider interface {
	Send(ctx context.Context, request core.OutboundSendRequest) (core.SendResult, error)
	GetDeliveryStatus(ctx context.Context, providerMessageID string) (core.DeliveryStatus, error)
	HealthCheck(ctx context.Context) (core.ProviderHealth, error)
}

type SendQueuedReplyInput struct {
	Organization domain.Organization
	Agent        domain.Agent
	Inbox        domain.Inbox
	Thread       domain.Thread
	Message      domain.Message
}

type Sender struct {
	provider EmailProvider
}

func NewSender(provider EmailProvider) Sender {
	return Sender{provider: provider}
}

func (s Sender) SendQueuedReply(ctx context.Context, input SendQueuedReplyInput) (core.SendResult, error) {
	return s.provider.Send(ctx, core.OutboundSendRequest{
		Organization: input.Organization,
		Agent:        input.Agent,
		Inbox:        input.Inbox,
		Thread:       input.Thread,
		Message:      input.Message,
	})
}

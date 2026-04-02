package outbound

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

const DeliveryStateSent = "sent"

type MessageStatusRepository interface {
	Save(ctx context.Context, message domain.Message) (domain.Message, error)
}

type RecordSentInput struct {
	Message    domain.Message
	SendResult core.SendResult
}

type StatusRecorder struct {
	repository MessageStatusRepository
}

func NewStatusRecorder(repository MessageStatusRepository) StatusRecorder {
	return StatusRecorder{repository: repository}
}

func (r StatusRecorder) RecordSent(ctx context.Context, input RecordSentInput) (domain.Message, error) {
	message := input.Message
	message.DeliveryState = DeliveryStateSent
	message.ProviderMessageID = input.SendResult.ProviderMessageID
	message.SentAt = input.SendResult.AcceptedAt

	return r.repository.Save(ctx, message)
}

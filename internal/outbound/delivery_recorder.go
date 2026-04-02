package outbound

import (
	"context"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

const (
	DeliveryStateDelivered  = "delivered"
	DeliveryStateHardBounce = "hard_bounce"
	DeliveryStateSoftBounce = "soft_bounce"
	DeliveryStateComplaint  = "complaint"
)

type DeliveryMessageRepository interface {
	Save(ctx context.Context, message domain.Message) (domain.Message, error)
}

type RecordDeliveryStatusInput struct {
	Message    domain.Message
	Status     string
	OccurredAt time.Time
}

type DeliveryRecorder struct {
	repository DeliveryMessageRepository
}

func NewDeliveryRecorder(repository DeliveryMessageRepository) DeliveryRecorder {
	return DeliveryRecorder{repository: repository}
}

func (r DeliveryRecorder) RecordStatus(ctx context.Context, input RecordDeliveryStatusInput) (domain.Message, error) {
	message := input.Message
	message.DeliveryState = input.Status

	switch input.Status {
	case DeliveryStateDelivered:
		message.DeliveredAt = input.OccurredAt
	case DeliveryStateHardBounce, DeliveryStateSoftBounce, DeliveryStateComplaint:
		message.BouncedAt = input.OccurredAt
	}

	return r.repository.Save(ctx, message)
}

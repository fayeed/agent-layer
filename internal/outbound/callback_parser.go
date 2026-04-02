package outbound

import (
	"encoding/json"
	"errors"
	"time"
)

type DeliveryCallbackEvent struct {
	ProviderMessageID string
	Status            string
	OccurredAt        time.Time
}

type callbackPayload struct {
	EventType         string `json:"event_type"`
	ProviderMessageID string `json:"provider_message_id"`
	OccurredAt        string `json:"occurred_at"`
}

type CallbackParser struct{}

func NewCallbackParser() CallbackParser {
	return CallbackParser{}
}

func (CallbackParser) Parse(body []byte) (DeliveryCallbackEvent, error) {
	var payload callbackPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return DeliveryCallbackEvent{}, err
	}

	occurredAt, err := time.Parse(time.RFC3339, payload.OccurredAt)
	if err != nil {
		return DeliveryCallbackEvent{}, err
	}

	status, err := mapCallbackEventType(payload.EventType)
	if err != nil {
		return DeliveryCallbackEvent{}, err
	}

	return DeliveryCallbackEvent{
		ProviderMessageID: payload.ProviderMessageID,
		Status:            status,
		OccurredAt:        occurredAt,
	}, nil
}

func mapCallbackEventType(eventType string) (string, error) {
	switch eventType {
	case DeliveryStateDelivered:
		return DeliveryStateDelivered, nil
	case DeliveryStateHardBounce:
		return DeliveryStateHardBounce, nil
	case DeliveryStateSoftBounce:
		return DeliveryStateSoftBounce, nil
	case DeliveryStateComplaint:
		return DeliveryStateComplaint, nil
	default:
		return "", errors.New("unknown delivery callback event type")
	}
}

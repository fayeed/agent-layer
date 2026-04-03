package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type WebhookReplayService interface {
	ReplayDelivery(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error)
}

type WebhookReplayHandler struct {
	service WebhookReplayService
}

func NewWebhookReplayHandler(service WebhookReplayService) WebhookReplayHandler {
	return WebhookReplayHandler{service: service}
}

func (h WebhookReplayHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	deliveryID := webhookDeliveryIDFromReplayPath(request.URL.Path)
	if deliveryID == "" {
		http.Error(writer, "delivery id is required", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.ReplayDelivery(request.Context(), deliveryID)
	if err != nil {
		writeLookupError(writer, err, "webhook delivery not found", "failed to replay webhook delivery")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(writer).Encode(webhookDeliveryResponse{
		ID:           delivery.ID,
		EventID:      delivery.EventID,
		EventType:    delivery.EventType,
		Status:       delivery.Status,
		AttemptCount: delivery.AttemptCount,
		ResponseCode: delivery.ResponseCode,
	})
}

type webhookDeliveryResponse struct {
	ID           string `json:"id"`
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	Status       string `json:"status"`
	AttemptCount int    `json:"attempt_count"`
	ResponseCode int    `json:"response_code"`
}

func webhookDeliveryIDFromReplayPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 4 || parts[0] != "webhooks" || parts[1] != "deliveries" || parts[3] != "replay" {
		return ""
	}
	return parts[2]
}

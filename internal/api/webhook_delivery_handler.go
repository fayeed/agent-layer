package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type WebhookDeliveryService interface {
	GetWebhookDelivery(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error)
}

type WebhookDeliveryHandler struct {
	service WebhookDeliveryService
}

func NewWebhookDeliveryHandler(service WebhookDeliveryService) WebhookDeliveryHandler {
	return WebhookDeliveryHandler{service: service}
}

func (h WebhookDeliveryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	deliveryID := webhookDeliveryIDFromShowPath(request.URL.Path)
	if deliveryID == "" {
		http.Error(writer, "delivery id is required", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.GetWebhookDelivery(request.Context(), deliveryID)
	if err != nil {
		http.Error(writer, "failed to load webhook delivery", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(webhookDeliveryResponse{
		ID:           delivery.ID,
		EventID:      delivery.EventID,
		EventType:    delivery.EventType,
		Status:       delivery.Status,
		AttemptCount: delivery.AttemptCount,
		ResponseCode: delivery.ResponseCode,
	})
}

func webhookDeliveryIDFromShowPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 3 || parts[0] != "webhooks" || parts[1] != "deliveries" {
		return ""
	}
	return parts[2]
}

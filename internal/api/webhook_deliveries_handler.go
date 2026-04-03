package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type WebhookDeliveriesService interface {
	ListWebhookDeliveries(ctx context.Context) ([]domain.WebhookDelivery, error)
}

type WebhookDeliveriesHandler struct {
	service WebhookDeliveriesService
}

func NewWebhookDeliveriesHandler(service WebhookDeliveriesService) WebhookDeliveriesHandler {
	return WebhookDeliveriesHandler{service: service}
}

func (h WebhookDeliveriesHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	deliveries, err := h.service.ListWebhookDeliveries(request.Context())
	if err != nil {
		http.Error(writer, "failed to load webhook deliveries", http.StatusInternalServerError)
		return
	}

	response := make([]webhookDeliveryResponse, 0, len(deliveries))
	for _, delivery := range deliveries {
		response = append(response, webhookDeliveryResponse{
			ID:           delivery.ID,
			EventID:      delivery.EventID,
			EventType:    delivery.EventType,
			Status:       delivery.Status,
			AttemptCount: delivery.AttemptCount,
			ResponseCode: delivery.ResponseCode,
		})
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(response)
}

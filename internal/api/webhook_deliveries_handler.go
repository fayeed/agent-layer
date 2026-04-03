package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type WebhookDeliveriesService interface {
	ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)
}

type WebhookDeliveriesHandler struct {
	service WebhookDeliveriesService
}

func NewWebhookDeliveriesHandler(service WebhookDeliveriesService) WebhookDeliveriesHandler {
	return WebhookDeliveriesHandler{service: service}
}

func (h WebhookDeliveriesHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	limit, err := webhookDeliveriesLimit(request)
	if err != nil {
		http.Error(writer, "invalid limit parameter", http.StatusBadRequest)
		return
	}

	deliveries, err := h.service.ListWebhookDeliveries(request.Context(), limit)
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

func webhookDeliveriesLimit(request *http.Request) (int, error) {
	value := request.URL.Query().Get("limit")
	if value == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit < 0 {
		return 0, strconv.ErrSyntax
	}

	return limit, nil
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
)

type WebhookRetryService interface {
	RetryDueDeliveries(ctx context.Context, limit int) (WebhookRetryResult, error)
}

type WebhookRetryResult struct {
	Attempted int `json:"attempted"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
}

type WebhookRetryHandler struct {
	service WebhookRetryService
}

func NewWebhookRetryHandler(service WebhookRetryService) WebhookRetryHandler {
	return WebhookRetryHandler{service: service}
}

func (h WebhookRetryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	limit, err := webhookDeliveriesLimit(request)
	if err != nil {
		http.Error(writer, "invalid limit parameter", http.StatusBadRequest)
		return
	}

	result, err := h.service.RetryDueDeliveries(request.Context(), limit)
	if err != nil {
		http.Error(writer, "failed to retry webhook deliveries", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(writer).Encode(result)
}

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type SuppressionsService interface {
	ListSuppressions(ctx context.Context, limit int) ([]domain.SuppressedAddress, error)
}

type SuppressionsHandler struct {
	service SuppressionsService
}

func NewSuppressionsHandler(service SuppressionsService) SuppressionsHandler {
	return SuppressionsHandler{service: service}
}

func (h SuppressionsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	limit, err := webhookDeliveriesLimit(request)
	if err != nil {
		http.Error(writer, "invalid limit parameter", http.StatusBadRequest)
		return
	}

	records, err := h.service.ListSuppressions(request.Context(), limit)
	if err != nil {
		http.Error(writer, "failed to load suppressions", http.StatusInternalServerError)
		return
	}

	response := make([]suppressionResponse, 0, len(records))
	for _, record := range records {
		response = append(response, suppressionResponse{
			ID:             record.ID,
			OrganizationID: record.OrganizationID,
			EmailAddress:   record.EmailAddress,
			Reason:         record.Reason,
			Source:         record.Source,
			CreatedAt:      formatResponseTime(record.CreatedAt),
			UpdatedAt:      formatResponseTime(record.UpdatedAt),
		})
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(response)
}

type suppressionResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	EmailAddress   string `json:"email_address"`
	Reason         string `json:"reason"`
	Source         string `json:"source"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

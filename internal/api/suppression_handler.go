package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type SuppressionService interface {
	GetSuppression(ctx context.Context, suppressionID string) (domain.SuppressedAddress, error)
}

type SuppressionHandler struct {
	service SuppressionService
}

func NewSuppressionHandler(service SuppressionService) SuppressionHandler {
	return SuppressionHandler{service: service}
}

func (h SuppressionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	suppressionID := suppressionIDFromPath(request.URL.Path)
	if suppressionID == "" {
		http.Error(writer, "suppression id is required", http.StatusBadRequest)
		return
	}

	record, err := h.service.GetSuppression(request.Context(), suppressionID)
	if err != nil {
		writeLookupError(writer, err, "suppression not found", "failed to load suppression")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(suppressionResponse{
		ID:             record.ID,
		OrganizationID: record.OrganizationID,
		EmailAddress:   record.EmailAddress,
		Reason:         record.Reason,
		Source:         record.Source,
		CreatedAt:      formatResponseTime(record.CreatedAt),
		UpdatedAt:      formatResponseTime(record.UpdatedAt),
	})
}

func suppressionIDFromPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 2 || parts[0] != "suppressions" {
		return ""
	}
	return parts[1]
}

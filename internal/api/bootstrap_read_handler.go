package api

import (
	"context"
	"encoding/json"
	"net/http"
)

type BootstrapReadService interface {
	GetBootstrap(ctx context.Context) (BootstrapResult, error)
}

type BootstrapReadHandler struct {
	service BootstrapReadService
}

func NewBootstrapReadHandler(service BootstrapReadService) BootstrapReadHandler {
	return BootstrapReadHandler{service: service}
}

func (h BootstrapReadHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	result, err := h.service.GetBootstrap(request.Context())
	if err != nil {
		http.Error(writer, "failed to load bootstrap config", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(bootstrapResponse{
		OrganizationID: result.OrganizationID,
		AgentID:        result.AgentID,
		InboxID:        result.InboxID,
		WebhookURL:     result.WebhookURL,
		InboxAddress:   result.InboxAddress,
	})
}

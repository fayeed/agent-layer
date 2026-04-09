package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type BootstrapService interface {
	BootstrapLocal(ctx context.Context, input BootstrapInput) (BootstrapResult, error)
}

type BootstrapInput struct {
	OrganizationName string `json:"organization_name"`
	AgentName        string `json:"agent_name"`
	AgentStatus      string `json:"agent_status"`
	WebhookURL       string `json:"webhook_url"`
	WebhookSecret    string `json:"webhook_secret"`
	InboxAddress     string `json:"inbox_address"`
	InboxDomain      string `json:"inbox_domain"`
	InboxDisplayName string `json:"inbox_display_name"`
}

type BootstrapResult struct {
	OrganizationID string
	AgentID        string
	InboxID        string
	WebhookURL     string
	InboxAddress   string
}

type bootstrapResponse struct {
	OrganizationID string `json:"organization_id"`
	AgentID        string `json:"agent_id"`
	InboxID        string `json:"inbox_id"`
	WebhookURL     string `json:"webhook_url"`
	InboxAddress   string `json:"inbox_address"`
}

type BootstrapHandler struct {
	service BootstrapService
}

func NewBootstrapHandler(service BootstrapService) BootstrapHandler {
	return BootstrapHandler{service: service}
}

func (h BootstrapHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var payload BootstrapInput
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "invalid json payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.OrganizationName) == "" {
		http.Error(writer, "organization_name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.AgentName) == "" {
		http.Error(writer, "agent_name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.InboxAddress) == "" {
		http.Error(writer, "inbox_address is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.InboxDomain) == "" {
		http.Error(writer, "inbox_domain is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.InboxDisplayName) == "" {
		http.Error(writer, "inbox_display_name is required", http.StatusBadRequest)
		return
	}

	status := domain.AgentStatus(payload.AgentStatus)
	if status != "" && !status.IsValid() {
		http.Error(writer, "invalid agent status", http.StatusBadRequest)
		return
	}

	payload.AgentStatus = string(status)
	result, err := h.service.BootstrapLocal(request.Context(), payload)
	if err != nil {
		http.Error(writer, "failed to bootstrap runtime", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(writer).Encode(bootstrapResponse{
		OrganizationID: result.OrganizationID,
		AgentID:        result.AgentID,
		InboxID:        result.InboxID,
		WebhookURL:     result.WebhookURL,
		InboxAddress:   result.InboxAddress,
	})
}

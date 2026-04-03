package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type ContactService interface {
	GetContact(ctx context.Context, contactID string) (domain.Contact, error)
}

type contactResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	EmailAddress   string `json:"email_address"`
	DisplayName    string `json:"display_name"`
}

type ContactHandler struct {
	service ContactService
}

func NewContactHandler(service ContactService) ContactHandler {
	return ContactHandler{service: service}
}

func (h ContactHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	contactID := contactIDFromPath(request.URL.Path)
	if contactID == "" {
		http.Error(writer, "contact id is required", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContact(request.Context(), contactID)
	if err != nil {
		writeLookupError(writer, err, "contact not found", "failed to load contact")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(contactResponse{
		ID:             contact.ID,
		OrganizationID: contact.OrganizationID,
		EmailAddress:   contact.EmailAddress,
		DisplayName:    contact.DisplayName,
	})
}

func contactIDFromPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 2 || parts[0] != "contacts" {
		return ""
	}
	return parts[1]
}

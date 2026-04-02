package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type CreateContactMemoryInput struct {
	ThreadID string   `json:"thread_id"`
	Note     string   `json:"note"`
	Tags     []string `json:"tags"`
}

type ContactMemoryService interface {
	CreateContactMemory(ctx context.Context, contactID string, input CreateContactMemoryInput) (domain.ContactMemoryEntry, error)
}

type contactMemoryResponse struct {
	ID        string   `json:"id"`
	ContactID string   `json:"contact_id"`
	ThreadID  string   `json:"thread_id"`
	Note      string   `json:"note"`
	Tags      []string `json:"tags"`
}

type ContactMemoryHandler struct {
	service ContactMemoryService
}

func NewContactMemoryHandler(service ContactMemoryService) ContactMemoryHandler {
	return ContactMemoryHandler{service: service}
}

func (h ContactMemoryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	contactID := contactIDFromMemoryPath(request.URL.Path)
	if contactID == "" {
		http.Error(writer, "contact id is required", http.StatusBadRequest)
		return
	}

	var payload CreateContactMemoryInput
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "invalid json payload", http.StatusBadRequest)
		return
	}

	entry, err := h.service.CreateContactMemory(request.Context(), contactID, payload)
	if err != nil {
		http.Error(writer, "failed to create contact memory", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(writer).Encode(contactMemoryResponse{
		ID:        entry.ID,
		ContactID: entry.ContactID,
		ThreadID:  entry.ThreadID,
		Note:      entry.Note,
		Tags:      entry.Tags,
	})
}

func contactIDFromMemoryPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 3 || parts[0] != "contacts" || parts[2] != "memory" {
		return ""
	}
	return parts[1]
}

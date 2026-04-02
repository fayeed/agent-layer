package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type ThreadService interface {
	GetThread(ctx context.Context, threadID string) (domain.Thread, error)
}

type threadResponse struct {
	ID                string `json:"id"`
	OrganizationID    string `json:"organization_id"`
	AgentID           string `json:"agent_id"`
	InboxID           string `json:"inbox_id"`
	ContactID         string `json:"contact_id"`
	SubjectNormalized string `json:"subject_normalized"`
	State             string `json:"state"`
	LastInboundID     string `json:"last_inbound_id"`
	LastOutboundID    string `json:"last_outbound_id"`
}

type ThreadHandler struct {
	service ThreadService
}

func NewThreadHandler(service ThreadService) ThreadHandler {
	return ThreadHandler{service: service}
}

func (h ThreadHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	threadID := threadIDFromThreadPath(request.URL.Path)
	if threadID == "" {
		http.Error(writer, "thread id is required", http.StatusBadRequest)
		return
	}

	thread, err := h.service.GetThread(request.Context(), threadID)
	if err != nil {
		http.Error(writer, "failed to load thread", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(threadResponse{
		ID:                thread.ID,
		OrganizationID:    thread.OrganizationID,
		AgentID:           thread.AgentID,
		InboxID:           thread.InboxID,
		ContactID:         thread.ContactID,
		SubjectNormalized: thread.SubjectNormalized,
		State:             string(thread.State),
		LastInboundID:     thread.LastInboundID,
		LastOutboundID:    thread.LastOutboundID,
	})
}

func threadIDFromThreadPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 2 || parts[0] != "threads" {
		return ""
	}
	return parts[1]
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return filterEmptySegments(path)
}

func filterEmptySegments(path string) []string {
	raw := make([]string, 0)
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				raw = append(raw, path[start:i])
			}
			start = i + 1
		}
	}
	return raw
}

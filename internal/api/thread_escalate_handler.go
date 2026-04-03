package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type ThreadEscalationService interface {
	EscalateThread(ctx context.Context, threadID, reason string) (domain.Thread, error)
}

type threadEscalateRequest struct {
	Reason string `json:"reason"`
}

type ThreadEscalateHandler struct {
	service ThreadEscalationService
}

func NewThreadEscalateHandler(service ThreadEscalationService) ThreadEscalateHandler {
	return ThreadEscalateHandler{service: service}
}

func (h ThreadEscalateHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	threadID := threadIDFromEscalatePath(request.URL.Path)
	if threadID == "" {
		http.Error(writer, "thread id is required", http.StatusBadRequest)
		return
	}

	var payload threadEscalateRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "invalid json payload", http.StatusBadRequest)
		return
	}

	thread, err := h.service.EscalateThread(request.Context(), threadID, payload.Reason)
	if err != nil {
		writeLookupError(writer, err, "thread not found", "failed to escalate thread")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
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

func threadIDFromEscalatePath(path string) string {
	parts := splitPath(path)
	if len(parts) != 3 || parts[0] != "threads" || parts[2] != "escalate" {
		return ""
	}
	return parts[1]
}

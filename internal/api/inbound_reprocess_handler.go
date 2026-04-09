package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type InboundReprocessService interface {
	ReprocessByObjectKey(ctx context.Context, objectKey string) (InboundReprocessResult, error)
}

type InboundReprocessResult struct {
	MessageID string
	ThreadID  string
	Duplicate bool
}

type inboundReprocessRequest struct {
	ObjectKey string `json:"object_key"`
}

type inboundReprocessResponse struct {
	MessageID string `json:"message_id"`
	ThreadID  string `json:"thread_id"`
	Duplicate bool   `json:"duplicate"`
}

type InboundReprocessHandler struct {
	service InboundReprocessService
}

func NewInboundReprocessHandler(service InboundReprocessService) InboundReprocessHandler {
	return InboundReprocessHandler{service: service}
}

func (h InboundReprocessHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var payload inboundReprocessRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "invalid json payload", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.ObjectKey) == "" {
		http.Error(writer, "object_key is required", http.StatusBadRequest)
		return
	}

	result, err := h.service.ReprocessByObjectKey(request.Context(), payload.ObjectKey)
	if err != nil {
		writeLookupError(writer, err, "inbound receipt not found", "failed to reprocess inbound message")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(writer).Encode(inboundReprocessResponse{
		MessageID: result.MessageID,
		ThreadID:  result.ThreadID,
		Duplicate: result.Duplicate,
	})
}

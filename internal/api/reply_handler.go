package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/outbound"
)

type ReplyService interface {
	SendReply(ctx context.Context, input outbound.SendReplyInput) (outbound.SendReplyResult, error)
}

type replyRequest struct {
	OrganizationID   string `json:"organization_id"`
	AgentID          string `json:"agent_id"`
	InboxID          string `json:"inbox_id"`
	ContactID        string `json:"contact_id"`
	IdempotencyKey   string `json:"idempotency_key"`
	ReplyToMessageID string `json:"reply_to_message_id"`
	BodyText         string `json:"body_text"`
	ObjectKey        string `json:"object_key"`
}

type replyResponse struct {
	MessageID         string `json:"message_id"`
	ThreadID          string `json:"thread_id"`
	Subject           string `json:"subject"`
	DeliveryState     string `json:"delivery_state"`
	ProviderMessageID string `json:"provider_message_id"`
}

type ReplyHandler struct {
	service ReplyService
}

func NewReplyHandler(service ReplyService) ReplyHandler {
	return ReplyHandler{service: service}
}

func (h ReplyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	threadID := threadIDFromPath(request.URL.Path)
	if threadID == "" {
		http.Error(writer, "thread id is required", http.StatusBadRequest)
		return
	}

	var payload replyRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "invalid json payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.ReplyToMessageID) == "" {
		http.Error(writer, "reply_to_message_id is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.BodyText) == "" {
		http.Error(writer, "body_text is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.ObjectKey) == "" {
		http.Error(writer, "object_key is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.IdempotencyKey) == "" {
		payload.IdempotencyKey = strings.TrimSpace(request.Header.Get("Idempotency-Key"))
	}

	result, err := h.service.SendReply(request.Context(), outbound.SendReplyInput{
		Organization: domain.Organization{ID: payload.OrganizationID},
		Agent:        domain.Agent{ID: payload.AgentID},
		Inbox:        domain.Inbox{ID: payload.InboxID},
		Thread:       domain.Thread{ID: threadID},
		ReplyToMessage: domain.Message{
			ID: payload.ReplyToMessageID,
		},
		Contact:        domain.Contact{ID: payload.ContactID},
		BodyText:       payload.BodyText,
		ObjectKey:      payload.ObjectKey,
		IdempotencyKey: payload.IdempotencyKey,
	})
	if err != nil {
		writeLookupError(writer, err, "reply context not found", "failed to send reply")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(writer).Encode(replyResponse{
		MessageID:         result.Message.ID,
		ThreadID:          result.Message.ThreadID,
		Subject:           result.Message.Subject,
		DeliveryState:     result.Message.DeliveryState,
		ProviderMessageID: result.Message.ProviderMessageID,
	})
}

func threadIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "threads" || parts[2] != "reply" {
		return ""
	}
	return parts[1]
}

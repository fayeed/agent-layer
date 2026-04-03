package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type ThreadMessagesService interface {
	ListThreadMessages(ctx context.Context, threadID string, limit int) ([]domain.Message, error)
}

type messageResponse struct {
	ID              string `json:"id"`
	ThreadID        string `json:"thread_id"`
	Direction       string `json:"direction"`
	Subject         string `json:"subject"`
	TextBody        string `json:"text_body"`
	MessageIDHeader string `json:"message_id_header"`
}

type ThreadMessagesHandler struct {
	service ThreadMessagesService
}

func NewThreadMessagesHandler(service ThreadMessagesService) ThreadMessagesHandler {
	return ThreadMessagesHandler{service: service}
}

func (h ThreadMessagesHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	threadID := threadIDFromThreadMessagesPath(request.URL.Path)
	if threadID == "" {
		http.Error(writer, "thread id is required", http.StatusBadRequest)
		return
	}

	limit, err := threadMessagesLimit(request)
	if err != nil {
		http.Error(writer, "invalid limit parameter", http.StatusBadRequest)
		return
	}

	messages, err := h.service.ListThreadMessages(request.Context(), threadID, limit)
	if err != nil {
		http.Error(writer, "failed to load thread messages", http.StatusInternalServerError)
		return
	}

	response := make([]messageResponse, 0, len(messages))
	for _, message := range messages {
		response = append(response, messageResponse{
			ID:              message.ID,
			ThreadID:        message.ThreadID,
			Direction:       string(message.Direction),
			Subject:         message.Subject,
			TextBody:        message.TextBody,
			MessageIDHeader: message.MessageIDHeader,
		})
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(response)
}

func threadIDFromThreadMessagesPath(path string) string {
	parts := splitPath(path)
	if len(parts) != 3 || parts[0] != "threads" || parts[2] != "messages" {
		return ""
	}
	return parts[1]
}

func threadMessagesLimit(request *http.Request) (int, error) {
	value := request.URL.Query().Get("limit")
	if value == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit < 0 {
		return 0, strconv.ErrSyntax
	}

	return limit, nil
}

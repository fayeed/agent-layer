package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

type InboundReceiptService interface {
	GetInboundReceipt(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error)
}

type inboundReceiptResponse struct {
	SMTPTransactionID   string   `json:"smtp_transaction_id"`
	OrganizationID      string   `json:"organization_id"`
	AgentID             string   `json:"agent_id"`
	InboxID             string   `json:"inbox_id"`
	EnvelopeSender      string   `json:"envelope_sender"`
	EnvelopeRecipients  []string `json:"envelope_recipients"`
	RawMessageObjectKey string   `json:"raw_message_object_key"`
	ReceivedAt          string   `json:"received_at"`
}

type InboundReceiptHandler struct {
	service InboundReceiptService
}

func NewInboundReceiptHandler(service InboundReceiptService) InboundReceiptHandler {
	return InboundReceiptHandler{service: service}
}

func (h InboundReceiptHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	objectKey := strings.TrimSpace(request.URL.Query().Get("object_key"))
	if objectKey == "" {
		http.Error(writer, "object_key is required", http.StatusBadRequest)
		return
	}

	receipt, err := h.service.GetInboundReceipt(request.Context(), objectKey)
	if err != nil {
		writeLookupError(writer, err, "inbound receipt not found", "failed to load inbound receipt")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(inboundReceiptResponse{
		SMTPTransactionID:   receipt.SMTPTransactionID,
		OrganizationID:      receipt.OrganizationID,
		AgentID:             receipt.AgentID,
		InboxID:             receipt.InboxID,
		EnvelopeSender:      receipt.EnvelopeSender,
		EnvelopeRecipients:  receipt.EnvelopeRecipients,
		RawMessageObjectKey: receipt.RawMessageObjectKey,
		ReceivedAt:          receipt.ReceivedAt.Format(http.TimeFormat),
	})
}

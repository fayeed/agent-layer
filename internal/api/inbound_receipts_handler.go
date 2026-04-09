package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

type InboundReceiptsService interface {
	ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error)
}

type InboundReceiptsHandler struct {
	service InboundReceiptsService
}

func NewInboundReceiptsHandler(service InboundReceiptsService) InboundReceiptsHandler {
	return InboundReceiptsHandler{service: service}
}

func (h InboundReceiptsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	limit, err := inboundReceiptsLimit(request)
	if err != nil {
		http.Error(writer, "invalid limit parameter", http.StatusBadRequest)
		return
	}

	receipts, err := h.service.ListInboundReceipts(request.Context(), limit)
	if err != nil {
		http.Error(writer, "failed to load inbound receipts", http.StatusInternalServerError)
		return
	}

	response := make([]inboundReceiptResponse, 0, len(receipts))
	for _, receipt := range receipts {
		response = append(response, inboundReceiptResponse{
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

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(response)
}

func inboundReceiptsLimit(request *http.Request) (int, error) {
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

package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/outbound"
)

type OutboundCallbackParser interface {
	Parse(body []byte) (outbound.DeliveryCallbackEvent, error)
}

type OutboundCallbackFlow interface {
	Apply(ctx context.Context, input outbound.CallbackFlowInput) (outbound.CallbackFlowResult, error)
}

type outboundCallbackPayload struct {
	ContactEmail string `json:"contact_email"`
}

type outboundCallbackResponse struct {
	MessageID         string `json:"message_id"`
	ProviderMessageID string `json:"provider_message_id"`
	DeliveryState     string `json:"delivery_state"`
	Suppressed        bool   `json:"suppressed"`
}

type OutboundCallbackHandler struct {
	parser OutboundCallbackParser
	flow   OutboundCallbackFlow
}

func NewOutboundCallbackHandler(parser OutboundCallbackParser, flow OutboundCallbackFlow) OutboundCallbackHandler {
	return OutboundCallbackHandler{
		parser: parser,
		flow:   flow,
	}
}

func (h OutboundCallbackHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	var payload outboundCallbackPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(writer, "invalid json payload", http.StatusBadRequest)
		return
	}

	event, err := h.parser.Parse(body)
	if err != nil {
		http.Error(writer, "invalid callback payload", http.StatusBadRequest)
		return
	}

	result, err := h.flow.Apply(request.Context(), outbound.CallbackFlowInput{
		Event: event,
		Contact: domain.Contact{
			EmailAddress: payload.ContactEmail,
		},
	})
	if err != nil {
		http.Error(writer, "failed to apply callback", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(outboundCallbackResponse{
		MessageID:         result.Message.ID,
		ProviderMessageID: result.Message.ProviderMessageID,
		DeliveryState:     result.Message.DeliveryState,
		Suppressed:        result.Suppressed,
	})
}

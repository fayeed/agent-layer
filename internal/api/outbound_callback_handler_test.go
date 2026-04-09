package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/outbound"
)

func TestOutboundCallbackHandlerAppliesCallbackFlow(t *testing.T) {
	parser := &callbackParserStub{
		event: outbound.DeliveryCallbackEvent{
			ProviderMessageID: "ses-123",
			Status:            outbound.DeliveryStateDelivered,
		},
	}
	flow := &callbackFlowStub{
		result: outbound.CallbackFlowResult{
			Message: domain.Message{
				ID:                "message-123",
				ProviderMessageID: "ses-123",
				DeliveryState:     outbound.DeliveryStateDelivered,
			},
		},
	}

	handler := NewOutboundCallbackHandler(parser, flow)
	request := httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{
		"event_type":"delivered",
		"provider_message_id":"ses-123",
		"occurred_at":"2026-04-03T13:00:00Z",
		"contact_email":"sender@example.com"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	if parser.body == "" {
		t.Fatal("expected request body to be parsed")
	}

	if flow.input.Event.ProviderMessageID != "ses-123" {
		t.Fatalf("expected parsed event to reach callback flow, got %#v", flow.input)
	}

	if flow.input.Contact.EmailAddress != "sender@example.com" {
		t.Fatalf("expected contact email from callback payload, got %#v", flow.input.Contact)
	}

	var response outboundCallbackResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.MessageID != "message-123" {
		t.Fatalf("expected updated message id in response, got %#v", response)
	}
}

func TestOutboundCallbackHandlerRejectsInvalidPayload(t *testing.T) {
	handler := NewOutboundCallbackHandler(&callbackParserStub{}, &callbackFlowStub{})
	request := httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestOutboundCallbackHandlerRequiresContactEmail(t *testing.T) {
	handler := NewOutboundCallbackHandler(&callbackParserStub{}, &callbackFlowStub{})
	request := httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{
		"event_type":"delivered",
		"provider_message_id":"ses-123",
		"occurred_at":"2026-04-03T13:00:00Z"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", recorder.Code)
	}
}

func TestOutboundCallbackHandlerReturnsNotFoundForUnknownProviderMessage(t *testing.T) {
	handler := NewOutboundCallbackHandler(&callbackParserStub{
		event: outbound.DeliveryCallbackEvent{
			ProviderMessageID: "ses-unknown",
			Status:            outbound.DeliveryStateDelivered,
		},
	}, &callbackFlowStub{err: outbound.ErrProviderMessageNotFound})

	request := httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{
		"event_type":"delivered",
		"provider_message_id":"ses-unknown",
		"occurred_at":"2026-04-03T13:00:00Z",
		"contact_email":"sender@example.com"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewOutboundCallbackHandler(&callbackParserStub{
		event: outbound.DeliveryCallbackEvent{
			ProviderMessageID: "ses-unknown",
			Status:            outbound.DeliveryStateDelivered,
		},
	}, &callbackFlowStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{
		"event_type":"delivered",
		"provider_message_id":"ses-unknown",
		"occurred_at":"2026-04-03T13:00:00Z",
		"contact_email":"sender@example.com"
	}`))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error status, got %d", recorder.Code)
	}
}

type callbackParserStub struct {
	body  string
	event outbound.DeliveryCallbackEvent
	err   error
}

func (s *callbackParserStub) Parse(body []byte) (outbound.DeliveryCallbackEvent, error) {
	s.body = string(body)
	return s.event, s.err
}

type callbackFlowStub struct {
	input  outbound.CallbackFlowInput
	result outbound.CallbackFlowResult
	err    error
}

func (s *callbackFlowStub) Apply(_ context.Context, input outbound.CallbackFlowInput) (outbound.CallbackFlowResult, error) {
	s.input = input
	return s.result, s.err
}

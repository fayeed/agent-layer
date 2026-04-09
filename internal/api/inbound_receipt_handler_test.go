package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestInboundReceiptHandlerLoadsStoredReceipt(t *testing.T) {
	handler := NewInboundReceiptHandler(inboundReceiptServiceStub{
		receipt: inbound.DurableReceiptRequest{
			SMTPTransactionID:   "smtp-session-123",
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			EnvelopeSender:      "sender@example.com",
			EnvelopeRecipients:  []string{"agent@example.com"},
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 9, 15, 0, 0, 0, time.UTC),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/inbound/receipts?object_key=raw/test-message.eml", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	var response inboundReceiptResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected receipt response json, got error: %v", err)
	}

	if response.SMTPTransactionID != "smtp-session-123" || response.RawMessageObjectKey != "raw/test-message.eml" {
		t.Fatalf("expected inbound receipt response, got %#v", response)
	}
}

func TestInboundReceiptHandlerValidatesAndMapsLookupErrors(t *testing.T) {
	handler := NewInboundReceiptHandler(inboundReceiptServiceStub{})

	request := httptest.NewRequest(http.MethodGet, "/inbound/receipts", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for missing object_key, got %d", recorder.Code)
	}

	handler = NewInboundReceiptHandler(inboundReceiptServiceStub{err: domain.ErrNotFound})
	request = httptest.NewRequest(http.MethodGet, "/inbound/receipts?object_key=raw/missing.eml", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewInboundReceiptHandler(inboundReceiptServiceStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodGet, "/inbound/receipts?object_key=raw/test-message.eml", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error status, got %d", recorder.Code)
	}
}

type inboundReceiptServiceStub struct {
	receipt inbound.DurableReceiptRequest
	err     error
}

func (s inboundReceiptServiceStub) GetInboundReceipt(context.Context, string) (inbound.DurableReceiptRequest, error) {
	return s.receipt, s.err
}

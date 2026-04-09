package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestInboundReceiptsHandlerListsReceipts(t *testing.T) {
	handler := NewInboundReceiptsHandler(&inboundReceiptsServiceStub{
		result: []inbound.DurableReceiptRequest{
			{
				SMTPTransactionID:   "smtp-session-123",
				RawMessageObjectKey: "raw/test-message.eml",
				ReceivedAt:          time.Date(2026, 4, 9, 15, 0, 0, 0, time.UTC),
			},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/inbound/receipts", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	var response []inboundReceiptResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected receipt list response json, got error: %v", err)
	}

	if len(response) != 1 || response[0].RawMessageObjectKey != "raw/test-message.eml" {
		t.Fatalf("expected inbound receipt list response, got %#v", response)
	}
}

func TestInboundReceiptsHandlerSupportsLimitValidation(t *testing.T) {
	service := &inboundReceiptsServiceStub{}
	handler := NewInboundReceiptsHandler(service)

	request := httptest.NewRequest(http.MethodGet, "/inbound/receipts?limit=2", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ok status, got %d", recorder.Code)
	}

	if service.limit != 2 {
		t.Fatalf("expected limit to be forwarded, got %d", service.limit)
	}

	request = httptest.NewRequest(http.MethodGet, "/inbound/receipts?limit=bad", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid limit, got %d", recorder.Code)
	}
}

type inboundReceiptsServiceStub struct {
	result []inbound.DurableReceiptRequest
	limit  int
	err    error
}

func (s *inboundReceiptsServiceStub) ListInboundReceipts(_ context.Context, limit int) ([]inbound.DurableReceiptRequest, error) {
	s.limit = limit
	return s.result, s.err
}

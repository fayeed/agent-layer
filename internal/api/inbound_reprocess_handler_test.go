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
)

func TestInboundReprocessHandlerReprocessesStoredReceipt(t *testing.T) {
	handler := NewInboundReprocessHandler(inboundReprocessServiceStub{
		result: InboundReprocessResult{
			MessageID: "message-123",
			ThreadID:  "thread-123",
			Duplicate: true,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/inbound/reprocess", bytes.NewBufferString(`{
		"object_key":"raw/test-message.eml"
	}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected accepted status, got %d", recorder.Code)
	}

	var response inboundReprocessResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}

	if response.MessageID != "message-123" || response.ThreadID != "thread-123" || !response.Duplicate {
		t.Fatalf("expected reprocess response, got %#v", response)
	}
}

func TestInboundReprocessHandlerValidatesAndMapsLookupErrors(t *testing.T) {
	handler := NewInboundReprocessHandler(inboundReprocessServiceStub{})

	request := httptest.NewRequest(http.MethodPost, "/inbound/reprocess", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for missing object_key, got %d", recorder.Code)
	}

	handler = NewInboundReprocessHandler(inboundReprocessServiceStub{err: domain.ErrNotFound})
	request = httptest.NewRequest(http.MethodPost, "/inbound/reprocess", bytes.NewBufferString(`{
		"object_key":"raw/missing.eml"
	}`))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", recorder.Code)
	}

	handler = NewInboundReprocessHandler(inboundReprocessServiceStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodPost, "/inbound/reprocess", bytes.NewBufferString(`{
		"object_key":"raw/test-message.eml"
	}`))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error status, got %d", recorder.Code)
	}
}

type inboundReprocessServiceStub struct {
	result InboundReprocessResult
	err    error
}

func (s inboundReprocessServiceStub) ReprocessByObjectKey(context.Context, string) (InboundReprocessResult, error) {
	return s.result, s.err
}

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookRetryHandlerReturnsSummary(t *testing.T) {
	handler := NewWebhookRetryHandler(&webhookRetryServiceStub{
		result: WebhookRetryResult{
			Attempted: 2,
			Succeeded: 1,
			Failed:    1,
			Skipped:   3,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/webhooks/deliveries/retry?limit=5", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected accepted response, got %d", recorder.Code)
	}

	var response WebhookRetryResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected json response, got error: %v", err)
	}
	if response.Attempted != 2 || response.Succeeded != 1 || response.Failed != 1 || response.Skipped != 3 {
		t.Fatalf("expected retry response, got %#v", response)
	}
}

func TestWebhookRetryHandlerValidatesAndHandlesErrors(t *testing.T) {
	handler := NewWebhookRetryHandler(&webhookRetryServiceStub{})

	request := httptest.NewRequest(http.MethodPost, "/webhooks/deliveries/retry?limit=bad", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid limit, got %d", recorder.Code)
	}

	handler = NewWebhookRetryHandler(&webhookRetryServiceStub{err: errors.New("boom")})
	request = httptest.NewRequest(http.MethodPost, "/webhooks/deliveries/retry", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error for retry failure, got %d", recorder.Code)
	}
}

type webhookRetryServiceStub struct {
	result WebhookRetryResult
	err    error
}

func (s *webhookRetryServiceStub) RetryDueDeliveries(context.Context, int) (WebhookRetryResult, error) {
	return s.result, s.err
}

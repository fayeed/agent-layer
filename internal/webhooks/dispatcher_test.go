package webhooks

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestDispatcherPostsWebhookAndCapturesResponse(t *testing.T) {
	client := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", request.Method)
		}

		if request.URL.String() != "https://example.com/webhook" {
			t.Fatalf("expected target url to be preserved, got %s", request.URL.String())
		}

		if request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected content type header, got %#v", request.Header)
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("expected request body to be readable, got %v", err)
		}

		if string(body) != `{"event":"message.received"}` {
			t.Fatalf("expected request payload to be sent, got %q", string(body))
		}

		return &http.Response{
			StatusCode: http.StatusAccepted,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, nil
	})

	dispatcher := NewDispatcher(client, func() time.Time {
		return time.Date(2026, 4, 3, 3, 0, 0, 0, time.UTC)
	})

	result, err := dispatcher.Dispatch(context.Background(), DispatchInput{
		URL: "https://example.com/webhook",
		Request: core.WebhookDispatchRequest{
			Delivery: domain.WebhookDelivery{
				ID:      "delivery-123",
				EventID: "event-123",
			},
			Payload: []byte(`{"event":"message.received"}`),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected dispatch to succeed, got error: %v", err)
	}

	if result.StatusCode != http.StatusAccepted {
		t.Fatalf("expected response status code, got %d", result.StatusCode)
	}

	if string(result.Body) != `{"ok":true}` {
		t.Fatalf("expected response body to be captured, got %q", string(result.Body))
	}

	expectedTime := time.Date(2026, 4, 3, 3, 0, 0, 0, time.UTC)
	if !result.DeliveredAt.Equal(expectedTime) {
		t.Fatalf("expected delivered time %v, got %v", expectedTime, result.DeliveredAt)
	}
}

func TestDispatcherReturnsTransportErrors(t *testing.T) {
	dispatcher := NewDispatcher(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	}), func() time.Time {
		return time.Date(2026, 4, 3, 3, 0, 0, 0, time.UTC)
	})

	_, err := dispatcher.Dispatch(context.Background(), DispatchInput{
		URL: "https://example.com/webhook",
		Request: core.WebhookDispatchRequest{
			Delivery: domain.WebhookDelivery{
				ID:      "delivery-123",
				EventID: "event-123",
			},
			Payload: []byte(`{"event":"message.received"}`),
		},
	})
	if err == nil {
		t.Fatal("expected transport error to be returned")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func (f roundTripFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

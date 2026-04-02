package webhooks

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
)

type HTTPDoer interface {
	Do(request *http.Request) (*http.Response, error)
}

type DispatchInput struct {
	URL     string
	Request core.WebhookDispatchRequest
}

type Dispatcher struct {
	client HTTPDoer
	now    Clock
}

func NewDispatcher(client HTTPDoer, now Clock) Dispatcher {
	if client == nil {
		client = &http.Client{}
	}
	if now == nil {
		now = time.Now
	}

	return Dispatcher{
		client: client,
		now:    now,
	}
}

func (d Dispatcher) Dispatch(ctx context.Context, input DispatchInput) (core.WebhookDispatchResult, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, input.URL, bytes.NewReader(input.Request.Payload))
	if err != nil {
		return core.WebhookDispatchResult{}, err
	}

	for key, value := range input.Request.Headers {
		request.Header.Set(key, value)
	}

	response, err := d.client.Do(request)
	if err != nil {
		return core.WebhookDispatchResult{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return core.WebhookDispatchResult{}, err
	}

	return core.WebhookDispatchResult{
		StatusCode:  response.StatusCode,
		Body:        body,
		DeliveredAt: d.now().UTC(),
	}, nil
}

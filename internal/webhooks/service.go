package webhooks

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
)

type MessageReceivedBuilderInterface interface {
	Build(ctx context.Context, input BuildMessageReceivedInput) (core.WebhookDispatchRequest, error)
}

type SignerInterface interface {
	Sign(request core.WebhookDispatchRequest, secret string) (core.WebhookDispatchRequest, error)
}

type DispatcherInterface interface {
	Dispatch(ctx context.Context, input DispatchInput) (core.WebhookDispatchResult, error)
}

type DeliverMessageReceivedInput struct {
	URL           string
	WebhookSecret string
	BuildInput    BuildMessageReceivedInput
}

type DeliverMessageReceivedResult struct {
	Request  core.WebhookDispatchRequest
	Response core.WebhookDispatchResult
}

type Service struct {
	builder    MessageReceivedBuilderInterface
	signer     SignerInterface
	dispatcher DispatcherInterface
}

func NewService(builder MessageReceivedBuilderInterface, signer SignerInterface, dispatcher DispatcherInterface) Service {
	return Service{
		builder:    builder,
		signer:     signer,
		dispatcher: dispatcher,
	}
}

func (s Service) DeliverMessageReceived(ctx context.Context, input DeliverMessageReceivedInput) (DeliverMessageReceivedResult, error) {
	request, err := s.builder.Build(ctx, input.BuildInput)
	if err != nil {
		return DeliverMessageReceivedResult{}, err
	}

	request, err = s.signer.Sign(request, input.WebhookSecret)
	if err != nil {
		return DeliverMessageReceivedResult{}, err
	}

	response, err := s.dispatcher.Dispatch(ctx, DispatchInput{
		URL:     input.URL,
		Request: request,
	})
	if err != nil {
		return DeliverMessageReceivedResult{}, err
	}

	return DeliverMessageReceivedResult{
		Request:  request,
		Response: response,
	}, nil
}

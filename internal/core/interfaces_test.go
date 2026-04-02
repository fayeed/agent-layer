package core

import (
	"context"
	"testing"
)

func TestInterfacesAreImplementedByCompileTimeStubs(t *testing.T) {
	ctx := context.Background()

	var inbound InboundTransport = inboundTransportStub{}
	if err := inbound.Receive(ctx, InboundReceipt{}); err != nil {
		t.Fatalf("unexpected inbound transport error: %v", err)
	}

	var parser MessageParser = messageParserStub{}
	if _, err := parser.Parse(ctx, StoredInboundMessage{}); err != nil {
		t.Fatalf("unexpected message parser error: %v", err)
	}

	var threadResolver ThreadResolver = threadResolverStub{}
	if _, err := threadResolver.Resolve(ctx, ThreadResolutionInput{}); err != nil {
		t.Fatalf("unexpected thread resolver error: %v", err)
	}

	var contactResolver ContactResolver = contactResolverStub{}
	if _, err := contactResolver.Resolve(ctx, ContactResolutionInput{}); err != nil {
		t.Fatalf("unexpected contact resolver error: %v", err)
	}

	var dispatcher WebhookDispatcher = webhookDispatcherStub{}
	if _, err := dispatcher.Dispatch(ctx, WebhookDispatchRequest{}); err != nil {
		t.Fatalf("unexpected webhook dispatcher error: %v", err)
	}

	var provider EmailProvider = emailProviderStub{}
	if _, err := provider.Send(ctx, OutboundSendRequest{}); err != nil {
		t.Fatalf("unexpected email provider send error: %v", err)
	}
	if _, err := provider.GetDeliveryStatus(ctx, "provider-message-id"); err != nil {
		t.Fatalf("unexpected email provider status error: %v", err)
	}
	if _, err := provider.HealthCheck(ctx); err != nil {
		t.Fatalf("unexpected email provider health error: %v", err)
	}
}

type inboundTransportStub struct{}

func (inboundTransportStub) Receive(context.Context, InboundReceipt) error {
	return nil
}

type messageParserStub struct{}

func (messageParserStub) Parse(context.Context, StoredInboundMessage) (ParsedMessage, error) {
	return ParsedMessage{}, nil
}

type threadResolverStub struct{}

func (threadResolverStub) Resolve(context.Context, ThreadResolutionInput) (ThreadResolutionResult, error) {
	return ThreadResolutionResult{}, nil
}

type contactResolverStub struct{}

func (contactResolverStub) Resolve(context.Context, ContactResolutionInput) (ContactResolutionResult, error) {
	return ContactResolutionResult{}, nil
}

type webhookDispatcherStub struct{}

func (webhookDispatcherStub) Dispatch(context.Context, WebhookDispatchRequest) (WebhookDispatchResult, error) {
	return WebhookDispatchResult{}, nil
}

type emailProviderStub struct{}

func (emailProviderStub) Send(context.Context, OutboundSendRequest) (SendResult, error) {
	return SendResult{}, nil
}

func (emailProviderStub) GetDeliveryStatus(context.Context, string) (DeliveryStatus, error) {
	return DeliveryStatus{}, nil
}

func (emailProviderStub) HealthCheck(context.Context) (ProviderHealth, error) {
	return ProviderHealth{}, nil
}

package ses

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestEmailProviderSendUsesRawMIMEAndRecipient(t *testing.T) {
	now := time.Date(2026, 4, 9, 16, 0, 0, 0, time.UTC)
	sender := &sendEmailStub{
		output: &sesv2.SendEmailOutput{MessageId: stringPtr("ses-123")},
	}
	provider := EmailProvider{sender: sender, health: &getAccountStub{}, region: "us-east-1", now: func() time.Time { return now }}

	result, err := provider.Send(context.Background(), core.OutboundSendRequest{
		Inbox:   domain.Inbox{EmailAddress: "agent@example.com"},
		Contact: domain.Contact{EmailAddress: "sender@example.com"},
		RawMIME: []byte("mime-body"),
	})
	if err != nil {
		t.Fatalf("expected send to succeed, got error: %v", err)
	}

	if result.ProviderMessageID != "ses-123" {
		t.Fatalf("expected provider message id, got %#v", result)
	}
	if sender.input == nil || sender.input.Destination == nil || len(sender.input.Destination.ToAddresses) != 1 || sender.input.Destination.ToAddresses[0] != "sender@example.com" {
		t.Fatalf("expected destination address, got %#v", sender.input)
	}
	if sender.input.Content == nil || sender.input.Content.Raw == nil || string(sender.input.Content.Raw.Data) != "mime-body" {
		t.Fatalf("expected raw mime content, got %#v", sender.input)
	}
}

func TestEmailProviderHealthCheckMapsFailure(t *testing.T) {
	provider := EmailProvider{
		sender: &sendEmailStub{},
		health: &getAccountStub{err: errors.New("boom")},
		region: "us-east-1",
		now:    time.Now,
	}

	health, err := provider.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected health check error")
	}
	if health.Healthy {
		t.Fatalf("expected unhealthy result, got %#v", health)
	}
}

type sendEmailStub struct {
	input  *sesv2.SendEmailInput
	output *sesv2.SendEmailOutput
	err    error
}

func (s *sendEmailStub) SendEmail(_ context.Context, input *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	s.input = input
	return s.output, s.err
}

type getAccountStub struct {
	err error
}

func (s *getAccountStub) GetAccount(context.Context, *sesv2.GetAccountInput, ...func(*sesv2.Options)) (*sesv2.GetAccountOutput, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &sesv2.GetAccountOutput{}, nil
}

func stringPtr(value string) *string {
	return &value
}

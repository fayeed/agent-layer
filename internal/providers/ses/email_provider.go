package ses

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	sesv2types "github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/agentlayer/agentlayer/internal/core"
)

type Clock func() time.Time

type sendEmailAPI interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

type getAccountAPI interface {
	GetAccount(ctx context.Context, params *sesv2.GetAccountInput, optFns ...func(*sesv2.Options)) (*sesv2.GetAccountOutput, error)
}

type EmailProvider struct {
	sender  sendEmailAPI
	health  getAccountAPI
	region  string
	now     Clock
}

func NewEmailProvider(ctx context.Context, region string, now Clock) (EmailProvider, error) {
	if now == nil {
		now = time.Now
	}
	loadOptions := []func(*config.LoadOptions) error{}
	if region != "" {
		loadOptions = append(loadOptions, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return EmailProvider{}, err
	}
	client := sesv2.NewFromConfig(cfg)
	return EmailProvider{
		sender: client,
		health: client,
		region: cfg.Region,
		now:    now,
	}, nil
}

func (p EmailProvider) Send(ctx context.Context, request core.OutboundSendRequest) (core.SendResult, error) {
	if len(request.RawMIME) == 0 {
		return core.SendResult{}, errors.New("raw mime is required")
	}
	if request.Inbox.EmailAddress == "" {
		return core.SendResult{}, errors.New("inbox email address is required")
	}
	if request.Contact.EmailAddress == "" {
		return core.SendResult{}, errors.New("contact email address is required")
	}

	output, err := p.sender.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &request.Inbox.EmailAddress,
		Destination: &sesv2types.Destination{
			ToAddresses: []string{request.Contact.EmailAddress},
		},
		Content: &sesv2types.EmailContent{
			Raw: &sesv2types.RawMessage{
				Data: request.RawMIME,
			},
		},
	})
	if err != nil {
		return core.SendResult{}, err
	}

	providerMessageID := ""
	if output != nil && output.MessageId != nil {
		providerMessageID = *output.MessageId
	}

	return core.SendResult{
		ProviderMessageID: providerMessageID,
		AcceptedAt:        p.now().UTC(),
	}, nil
}

func (p EmailProvider) GetDeliveryStatus(_ context.Context, providerMessageID string) (core.DeliveryStatus, error) {
	return core.DeliveryStatus{
		ProviderMessageID: providerMessageID,
		State:             "accepted",
		UpdatedAt:         p.now().UTC(),
	}, nil
}

func (p EmailProvider) HealthCheck(ctx context.Context) (core.ProviderHealth, error) {
	_, err := p.health.GetAccount(ctx, &sesv2.GetAccountInput{})
	if err != nil {
		return core.ProviderHealth{
			ProviderName: "ses",
			Healthy:      false,
			CheckedAt:    p.now().UTC(),
			Details:      err.Error(),
		}, err
	}
	return core.ProviderHealth{
		ProviderName: "ses",
		Healthy:      true,
		CheckedAt:    p.now().UTC(),
		Details:      "region=" + p.region,
	}, nil
}

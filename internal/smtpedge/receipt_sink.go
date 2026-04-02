package smtpedge

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

type StoredMessageHandler interface {
	HandleStoredMessage(ctx context.Context, stored core.StoredInboundMessage) (inbound.HandleResult, error)
}

type ReceiptHandlerSink struct {
	handler StoredMessageHandler
}

func NewReceiptSink(handler StoredMessageHandler) ReceiptHandlerSink {
	return ReceiptHandlerSink{handler: handler}
}

func (s ReceiptHandlerSink) Enqueue(ctx context.Context, receipt inbound.DurableReceiptRequest) error {
	_, err := s.handler.HandleStoredMessage(ctx, core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			SMTPTransactionID:   receipt.SMTPTransactionID,
			OrganizationID:      receipt.OrganizationID,
			AgentID:             receipt.AgentID,
			InboxID:             receipt.InboxID,
			EnvelopeSender:      receipt.EnvelopeSender,
			EnvelopeRecipients:  receipt.EnvelopeRecipients,
			RawMessageObjectKey: receipt.RawMessageObjectKey,
			ReceivedAt:          receipt.ReceivedAt,
		},
	})
	return err
}

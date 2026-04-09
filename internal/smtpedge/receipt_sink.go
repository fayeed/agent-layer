package smtpedge

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

type StoredMessageHandler interface {
	HandleStoredMessage(ctx context.Context, stored core.StoredInboundMessage) (inbound.HandleResult, error)
}

type ReceiptRecorder interface {
	SaveInboundReceipt(ctx context.Context, receipt inbound.DurableReceiptRequest) error
}

type ReceiptHandlerSink struct {
	handler  StoredMessageHandler
	recorder ReceiptRecorder
}

func NewReceiptSink(handler StoredMessageHandler) ReceiptHandlerSink {
	return ReceiptHandlerSink{handler: handler}
}

func NewReceiptSinkWithRecorder(handler StoredMessageHandler, recorder ReceiptRecorder) ReceiptHandlerSink {
	return ReceiptHandlerSink{
		handler:  handler,
		recorder: recorder,
	}
}

func (s ReceiptHandlerSink) Enqueue(ctx context.Context, receipt inbound.DurableReceiptRequest) error {
	if s.recorder != nil {
		if err := s.recorder.SaveInboundReceipt(ctx, receipt); err != nil {
			return err
		}
	}

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

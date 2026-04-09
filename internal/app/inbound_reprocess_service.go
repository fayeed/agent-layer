package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

type InboundReceiptGetter interface {
	GetInboundReceiptByObjectKey(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error)
}

type InboundReprocessService struct {
	receipts InboundReceiptGetter
	inbound  InboundHandler
}

func NewInboundReprocessService(receipts InboundReceiptGetter, inboundHandler InboundHandler) InboundReprocessService {
	return InboundReprocessService{
		receipts: receipts,
		inbound:  inboundHandler,
	}
}

func (s InboundReprocessService) ReprocessByObjectKey(ctx context.Context, objectKey string) (inbound.HandleResult, error) {
	receipt, err := s.receipts.GetInboundReceiptByObjectKey(ctx, objectKey)
	if err != nil {
		return inbound.HandleResult{}, err
	}

	return s.inbound.HandleStoredMessage(ctx, core.StoredInboundMessage{
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
}

package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

type InboundReceiptReadService struct {
	receipts InboundReceiptGetter
}

func NewInboundReceiptReadService(receipts InboundReceiptGetter) InboundReceiptReadService {
	return InboundReceiptReadService{receipts: receipts}
}

func (s InboundReceiptReadService) GetInboundReceipt(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error) {
	return s.receipts.GetInboundReceiptByObjectKey(ctx, objectKey)
}

package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

type InboundReceiptLister interface {
	ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error)
}

type InboundReceiptListService struct {
	receipts InboundReceiptLister
	limit    int
}

func NewInboundReceiptListService(receipts InboundReceiptLister, limit int) InboundReceiptListService {
	if limit <= 0 {
		limit = 20
	}
	return InboundReceiptListService{
		receipts: receipts,
		limit:    limit,
	}
}

func (s InboundReceiptListService) ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error) {
	if limit <= 0 {
		limit = s.limit
	}
	return s.receipts.ListInboundReceipts(ctx, limit)
}

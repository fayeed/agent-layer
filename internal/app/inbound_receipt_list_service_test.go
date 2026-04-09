package app

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestInboundReceiptListServiceUsesDefaultLimit(t *testing.T) {
	lister := &inboundReceiptListerStub{
		result: []inbound.DurableReceiptRequest{
			{RawMessageObjectKey: "raw/test-message.eml", ReceivedAt: time.Date(2026, 4, 9, 15, 0, 0, 0, time.UTC)},
		},
	}

	service := NewInboundReceiptListService(lister, 5)

	receipts, err := service.ListInboundReceipts(context.Background(), 0)
	if err != nil {
		t.Fatalf("expected inbound receipt list to succeed, got error: %v", err)
	}

	if lister.limit != 5 {
		t.Fatalf("expected default limit to be applied, got %d", lister.limit)
	}

	if len(receipts) != 1 || receipts[0].RawMessageObjectKey != "raw/test-message.eml" {
		t.Fatalf("expected listed receipts, got %#v", receipts)
	}
}

func TestInboundReceiptListServiceUsesRequestedLimit(t *testing.T) {
	lister := &inboundReceiptListerStub{}
	service := NewInboundReceiptListService(lister, 20)

	if _, err := service.ListInboundReceipts(context.Background(), 2); err != nil {
		t.Fatalf("expected inbound receipt list to succeed, got error: %v", err)
	}

	if lister.limit != 2 {
		t.Fatalf("expected requested limit to be forwarded, got %d", lister.limit)
	}
}

type inboundReceiptListerStub struct {
	result []inbound.DurableReceiptRequest
	limit  int
	err    error
}

func (s *inboundReceiptListerStub) ListInboundReceipts(_ context.Context, limit int) ([]inbound.DurableReceiptRequest, error) {
	s.limit = limit
	return s.result, s.err
}

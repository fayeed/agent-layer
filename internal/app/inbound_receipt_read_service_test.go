package app

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/inbound"
)

func TestInboundReceiptReadServiceLoadsStoredReceipt(t *testing.T) {
	service := NewInboundReceiptReadService(inboundReceiptGetterStub{
		receipt: inbound.DurableReceiptRequest{
			SMTPTransactionID:   "smtp-session-123",
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 9, 15, 0, 0, 0, time.UTC),
		},
	})

	receipt, err := service.GetInboundReceipt(context.Background(), "raw/test-message.eml")
	if err != nil {
		t.Fatalf("expected inbound receipt read to succeed, got error: %v", err)
	}

	if receipt.SMTPTransactionID != "smtp-session-123" || receipt.RawMessageObjectKey != "raw/test-message.eml" {
		t.Fatalf("expected loaded inbound receipt, got %#v", receipt)
	}
}

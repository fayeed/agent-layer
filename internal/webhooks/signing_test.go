package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestSignerAddsTimestampAndSignatureHeaders(t *testing.T) {
	at := time.Date(2026, 4, 3, 2, 0, 0, 0, time.UTC)
	signer := NewSigner(func() time.Time { return at })

	request, err := signer.Sign(core.WebhookDispatchRequest{
		Delivery: domain.WebhookDelivery{
			EventID: "event-123",
		},
		Payload: []byte(`{"hello":"world"}`),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, "super-secret")
	if err != nil {
		t.Fatalf("expected signing to succeed, got error: %v", err)
	}

	if request.Headers[HeaderSignatureTimestamp] != "2026-04-03T02:00:00Z" {
		t.Fatalf("expected timestamp header to be set, got %#v", request.Headers)
	}

	expectedMAC := hmac.New(sha256.New, []byte("super-secret"))
	expectedMAC.Write([]byte("2026-04-03T02:00:00Z."))
	expectedMAC.Write([]byte(`{"hello":"world"}`))

	expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC.Sum(nil))
	if request.Headers[HeaderSignature] != expectedSignature {
		t.Fatalf("expected signature header %q, got %q", expectedSignature, request.Headers[HeaderSignature])
	}

	if request.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected existing headers to be preserved, got %#v", request.Headers)
	}
}

func TestSignerRejectsEmptySecret(t *testing.T) {
	signer := NewSigner(func() time.Time {
		return time.Date(2026, 4, 3, 2, 0, 0, 0, time.UTC)
	})

	_, err := signer.Sign(core.WebhookDispatchRequest{
		Delivery: domain.WebhookDelivery{
			EventID: "event-123",
		},
		Payload: []byte(`{"hello":"world"}`),
	}, "")
	if err == nil {
		t.Fatal("expected empty secret to fail signing")
	}
}

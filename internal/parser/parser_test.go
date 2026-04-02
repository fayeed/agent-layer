package parser

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
)

func TestParserParsesMultipartMessage(t *testing.T) {
	raw := "From: Sender Example <sender@example.com>\r\n" +
		"To: Agent Example <agent@example.com>\r\n" +
		"Cc: Copy Example <copy@example.com>\r\n" +
		"Reply-To: Reply Example <reply@example.com>\r\n" +
		"Subject: Re: Hello World\r\n" +
		"Message-ID: <message-123@example.com>\r\n" +
		"In-Reply-To: <message-122@example.com>\r\n" +
		"References: <message-100@example.com> <message-122@example.com>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=mixed-boundary\r\n" +
		"\r\n" +
		"--mixed-boundary\r\n" +
		"Content-Type: multipart/alternative; boundary=alt-boundary\r\n" +
		"\r\n" +
		"--alt-boundary\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"Plain body.\r\n" +
		"--alt-boundary\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		"<p>HTML body.</p>\r\n" +
		"--alt-boundary--\r\n" +
		"--mixed-boundary\r\n" +
		"Content-Type: application/pdf\r\n" +
		"Content-Disposition: attachment; filename=\"invoice.pdf\"\r\n" +
		"Content-ID: <attachment-1>\r\n" +
		"\r\n" +
		"%PDF-1.4\r\n" +
		"--mixed-boundary--\r\n"

	reader := rawMessageReaderStub{
		payloads: map[string][]byte{
			"raw/test-message.eml": []byte(raw),
		},
	}

	parser := New(reader)

	parsed, err := parser.Parse(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("expected parse to succeed, got error: %v", err)
	}

	if parsed.MessageIDHeader != "<message-123@example.com>" {
		t.Fatalf("expected message id header to be parsed, got %q", parsed.MessageIDHeader)
	}

	if parsed.InReplyTo != "<message-122@example.com>" {
		t.Fatalf("expected in-reply-to to be parsed, got %q", parsed.InReplyTo)
	}

	if len(parsed.References) != 2 {
		t.Fatalf("expected 2 references, got %d", len(parsed.References))
	}

	if parsed.Subject != "Re: Hello World" {
		t.Fatalf("expected subject to be preserved, got %q", parsed.Subject)
	}

	if parsed.SubjectNormalized != "hello world" {
		t.Fatalf("expected normalized subject, got %q", parsed.SubjectNormalized)
	}

	if parsed.TextBody != "Plain body." {
		t.Fatalf("expected text body to be parsed, got %q", parsed.TextBody)
	}

	if parsed.HTMLBody != "<p>HTML body.</p>" {
		t.Fatalf("expected html body to be parsed, got %q", parsed.HTMLBody)
	}

	if parsed.From.Email != "sender@example.com" || parsed.From.DisplayName != "Sender Example" {
		t.Fatalf("expected from address to be parsed, got %#v", parsed.From)
	}

	if len(parsed.To) != 1 || parsed.To[0].Email != "agent@example.com" {
		t.Fatalf("expected to address to be parsed, got %#v", parsed.To)
	}

	if len(parsed.CC) != 1 || parsed.CC[0].Email != "copy@example.com" {
		t.Fatalf("expected cc address to be parsed, got %#v", parsed.CC)
	}

	if len(parsed.ReplyTo) != 1 || parsed.ReplyTo[0].Email != "reply@example.com" {
		t.Fatalf("expected reply-to address to be parsed, got %#v", parsed.ReplyTo)
	}

	if len(parsed.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(parsed.Attachments))
	}

	attachment := parsed.Attachments[0]
	if attachment.FileName != "invoice.pdf" {
		t.Fatalf("expected attachment filename, got %q", attachment.FileName)
	}

	if attachment.ContentType != "application/pdf" {
		t.Fatalf("expected attachment content type, got %q", attachment.ContentType)
	}

	if attachment.ContentID != "<attachment-1>" {
		t.Fatalf("expected attachment content id, got %q", attachment.ContentID)
	}

	if attachment.ContentDisposition != "attachment; filename=\"invoice.pdf\"" {
		t.Fatalf("expected attachment disposition, got %q", attachment.ContentDisposition)
	}

	if attachment.SizeBytes == 0 {
		t.Fatal("expected attachment size to be captured")
	}
}

type rawMessageReaderStub struct {
	payloads map[string][]byte
}

func (r rawMessageReaderStub) Get(_ context.Context, objectKey string) ([]byte, error) {
	return r.payloads[objectKey], nil
}

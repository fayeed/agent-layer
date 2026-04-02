package smtpedge

import (
	"testing"
	"time"
)

func TestNewServerAppliesDefaultConfig(t *testing.T) {
	server := NewServer(NewBackend(func() CoreSession {
		return &coreSessionStub{}
	}), Config{})

	if server.Domain != "localhost" {
		t.Fatalf("expected default domain, got %q", server.Domain)
	}

	if server.MaxRecipients != 1 {
		t.Fatalf("expected default max recipients, got %d", server.MaxRecipients)
	}

	if server.MaxMessageBytes != 25*1024*1024 {
		t.Fatalf("expected default max message bytes, got %d", server.MaxMessageBytes)
	}

	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("expected default read timeout, got %v", server.ReadTimeout)
	}

	if server.WriteTimeout != 10*time.Second {
		t.Fatalf("expected default write timeout, got %v", server.WriteTimeout)
	}
}

func TestNewServerAppliesExplicitConfig(t *testing.T) {
	server := NewServer(NewBackend(func() CoreSession {
		return &coreSessionStub{}
	}), Config{
		Addr:            "127.0.0.1:2525",
		Domain:          "mail.agentlayer.dev",
		MaxRecipients:   3,
		MaxMessageBytes: 4 * 1024 * 1024,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    7 * time.Second,
	})

	if server.Addr != "127.0.0.1:2525" {
		t.Fatalf("expected configured address, got %q", server.Addr)
	}

	if server.Domain != "mail.agentlayer.dev" {
		t.Fatalf("expected configured domain, got %q", server.Domain)
	}

	if server.MaxRecipients != 3 {
		t.Fatalf("expected configured max recipients, got %d", server.MaxRecipients)
	}

	if server.MaxMessageBytes != 4*1024*1024 {
		t.Fatalf("expected configured max message bytes, got %d", server.MaxMessageBytes)
	}

	if server.ReadTimeout != 5*time.Second {
		t.Fatalf("expected configured read timeout, got %v", server.ReadTimeout)
	}

	if server.WriteTimeout != 7*time.Second {
		t.Fatalf("expected configured write timeout, got %v", server.WriteTimeout)
	}
}

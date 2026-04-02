package smtpedge

import (
	"bytes"
	"context"
	"io"
	"testing"

	smtp "github.com/emersion/go-smtp"
)

func TestBackendCreatesLibraryCompatibleSession(t *testing.T) {
	coreSession := &coreSessionStub{}
	backend := NewBackend(func() CoreSession {
		return coreSession
	})

	session, err := backend.NewSession(nil)
	if err != nil {
		t.Fatalf("expected backend to create session, got error: %v", err)
	}

	if _, ok := session.(*AdapterSession); !ok {
		t.Fatalf("expected smtp session adapter, got %T", session)
	}
}

func TestAdapterSessionDelegatesMailRcptAndData(t *testing.T) {
	coreSession := &coreSessionStub{}
	session := NewAdapterSession(coreSession)

	if err := session.Mail("sender@example.com", &smtp.MailOptions{}); err != nil {
		t.Fatalf("expected mail delegation to succeed, got error: %v", err)
	}

	if err := session.Rcpt("agent@example.com", &smtp.RcptOptions{}); err != nil {
		t.Fatalf("expected rcpt delegation to succeed, got error: %v", err)
	}

	if err := session.Data(bytes.NewBufferString("raw mime body")); err != nil {
		t.Fatalf("expected data delegation to succeed, got error: %v", err)
	}

	if coreSession.from != "sender@example.com" {
		t.Fatalf("expected MAIL FROM to reach core session, got %q", coreSession.from)
	}

	if coreSession.to != "agent@example.com" {
		t.Fatalf("expected RCPT TO to reach core session, got %q", coreSession.to)
	}

	if string(coreSession.data) != "raw mime body" {
		t.Fatalf("expected DATA to reach core session, got %q", string(coreSession.data))
	}
}

type coreSessionStub struct {
	from string
	to   string
	data []byte
}

func (s *coreSessionStub) Mail(_ context.Context, from string) error {
	s.from = from
	return nil
}

func (s *coreSessionStub) Rcpt(_ context.Context, to string) error {
	s.to = to
	return nil
}

func (s *coreSessionStub) Data(_ context.Context, reader io.Reader) error {
	payload, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.data = payload
	return nil
}

package smtpedge

import (
	"context"
	"io"

	smtp "github.com/emersion/go-smtp"
)

type CoreSession interface {
	Mail(ctx context.Context, from string) error
	Rcpt(ctx context.Context, to string) error
	Data(ctx context.Context, reader io.Reader) error
}

type Backend struct {
	newSession func() CoreSession
}

func NewBackend(newSession func() CoreSession) Backend {
	return Backend{newSession: newSession}
}

func (b Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return NewAdapterSession(b.newSession()), nil
}

type AdapterSession struct {
	core CoreSession
}

func NewAdapterSession(core CoreSession) *AdapterSession {
	return &AdapterSession{core: core}
}

func (s *AdapterSession) Mail(from string, _ *smtp.MailOptions) error {
	return s.core.Mail(context.Background(), from)
}

func (s *AdapterSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	return s.core.Rcpt(context.Background(), to)
}

func (s *AdapterSession) Data(reader io.Reader) error {
	return s.core.Data(context.Background(), reader)
}

func (s *AdapterSession) Reset() {}

func (s *AdapterSession) Logout() error {
	return nil
}

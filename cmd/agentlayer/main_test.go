package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/outbound"
)

func TestNewServerExposesHealthEndpoint(t *testing.T) {
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected health endpoint to return 200, got %d", recorder.Code)
	}

	if recorder.Body.String() != "ok\n" {
		t.Fatalf("expected health response body, got %q", recorder.Body.String())
	}
}

func TestNewServerRegistersV0RouteShapes(t *testing.T) {
	server := newServer()

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/threads/thread-123/reply"},
		{method: http.MethodPost, path: "/threads/thread-123/escalate"},
		{method: http.MethodGet, path: "/threads/thread-123"},
		{method: http.MethodGet, path: "/threads/thread-123/messages"},
		{method: http.MethodGet, path: "/contacts/contact-123"},
		{method: http.MethodPost, path: "/contacts/contact-123/memory"},
		{method: http.MethodPost, path: "/provider/callbacks/outbound"},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		if recorder.Code == http.StatusNotFound {
			t.Fatalf("expected route %s %s to be registered", tt.method, tt.path)
		}
	}
}

func TestNewServerWiresThreadHandler(t *testing.T) {
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected real thread handler path to return 500 from placeholder service, got %d", recorder.Code)
	}
}

func TestNewServerWiresContactHandler(t *testing.T) {
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/contacts/contact-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected real contact handler path to return 500 from placeholder service, got %d", recorder.Code)
	}
}

func TestNewServerWiresRemainingHandlers(t *testing.T) {
	server := newServer()

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodPost, path: "/threads/thread-123/reply", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/threads/thread-123/escalate", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodGet, path: "/threads/thread-123/messages", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/contacts/contact-123/memory", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/provider/callbacks/outbound", body: "{}", want: http.StatusBadRequest},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		if recorder.Code != tt.want {
			t.Fatalf("expected route %s %s to return %d from wired handler, got %d", tt.method, tt.path, tt.want, recorder.Code)
		}
	}
}

func TestNewSMTPServerUsesDefaults(t *testing.T) {
	t.Setenv("AGENTLAYER_SMTP_ADDR", "")
	t.Setenv("AGENTLAYER_SMTP_DOMAIN", "")

	server := newSMTPServer()

	if server.Addr != "localhost:2525" {
		t.Fatalf("expected default smtp addr, got %q", server.Addr)
	}

	if server.Domain != "localhost" {
		t.Fatalf("expected default smtp domain, got %q", server.Domain)
	}
}

func TestNewSMTPServerUsesEnvOverrides(t *testing.T) {
	t.Setenv("AGENTLAYER_SMTP_ADDR", "127.0.0.1:2626")
	t.Setenv("AGENTLAYER_SMTP_DOMAIN", "mail.agentlayer.dev")

	server := newSMTPServer()

	if server.Addr != "127.0.0.1:2626" {
		t.Fatalf("expected configured smtp addr, got %q", server.Addr)
	}

	if server.Domain != "mail.agentlayer.dev" {
		t.Fatalf("expected configured smtp domain, got %q", server.Domain)
	}
}

func TestSMTPAddressHelpers(t *testing.T) {
	t.Setenv("AGENTLAYER_SMTP_ADDR", "0.0.0.0:2526")
	t.Setenv("AGENTLAYER_SMTP_DOMAIN", "smtp.example.com")

	if got := smtpAddress(); got != "0.0.0.0:2526" {
		t.Fatalf("expected smtp addr helper to read env, got %q", got)
	}

	if got := smtpDomain(); got != "smtp.example.com" {
		t.Fatalf("expected smtp domain helper to read env, got %q", got)
	}
}

func TestNewInboundServiceUsesComposedPlaceholderDependencies(t *testing.T) {
	service := newInboundService()

	_, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{})
	if err == nil {
		t.Fatal("expected placeholder inbound service to return an error")
	}
}

func TestNewInboundProcessorUsesRealProcessorChain(t *testing.T) {
	processor := newInboundProcessor()

	_, err := processor.Process(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID:      "org-123",
			AgentID:             "agent-123",
			InboxID:             "inbox-123",
			RawMessageObjectKey: "raw/test-message.eml",
			ReceivedAt:          time.Date(2026, 4, 3, 17, 0, 0, 0, time.UTC),
		},
	})
	if err == nil {
		t.Fatal("expected placeholder raw message reader to fail")
	}
}

func TestNewInboundRecorderUsesRealRecorderChain(t *testing.T) {
	recorder := newInboundRecorder()

	_, err := recorder.Record(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID: "org-123",
			InboxID:        "inbox-123",
			ReceivedAt:     time.Date(2026, 4, 3, 17, 5, 0, 0, time.UTC),
		},
	}, inbound.ProcessResult{
		Contact: domain.Contact{ID: "contact-123"},
		Thread:  domain.Thread{ID: "thread-123"},
	})
	if err == nil {
		t.Fatal("expected placeholder inbound repositories to fail")
	}
}

func TestNewThreadReadServiceUsesApplicationService(t *testing.T) {
	service := newThreadReadService()

	_, err := service.GetThread(context.Background(), "thread-123")
	if err == nil {
		t.Fatal("expected placeholder thread repository to fail")
	}
}

func TestNewContactReadServiceUsesApplicationService(t *testing.T) {
	service := newContactReadService()

	_, err := service.GetContact(context.Background(), "contact-123")
	if err == nil {
		t.Fatal("expected placeholder contact repository to fail")
	}
}

func TestNewReplyServiceUsesRealOutboundComposition(t *testing.T) {
	service := newReplyService()

	_, err := service.SendReply(context.Background(), outbound.SendReplyInput{
		Organization: domain.Organization{ID: "org-123"},
		Agent:        domain.Agent{ID: "agent-123"},
		Inbox: domain.Inbox{
			ID:           "inbox-123",
			EmailAddress: "agent@example.com",
		},
		Thread: domain.Thread{
			ID: "thread-123",
		},
		ReplyToMessage: domain.Message{
			ID:              "message-100",
			ThreadID:        "thread-123",
			MessageIDHeader: "<message-100@example.com>",
			Subject:         "Hello World",
		},
		Contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
		BodyText:  "Thanks for reaching out.",
		ObjectKey: "outbound/reply-123.eml",
	})
	if err == nil {
		t.Fatal("expected placeholder outbound repositories/provider to fail")
	}
}

func TestNewOutboundCallbackFlowUsesRealCallbackComposition(t *testing.T) {
	flow := newOutboundCallbackFlow()

	_, err := flow.Apply(context.Background(), outbound.CallbackFlowInput{
		Event: outbound.DeliveryCallbackEvent{
			ProviderMessageID: "ses-123",
			Status:            outbound.DeliveryStateDelivered,
			OccurredAt:        time.Date(2026, 4, 3, 17, 10, 0, 0, time.UTC),
		},
		Contact: domain.Contact{
			EmailAddress: "sender@example.com",
		},
	})
	if err == nil {
		t.Fatal("expected placeholder callback dependencies to fail")
	}
}

func TestRunServersStartsHTTPAndSMTP(t *testing.T) {
	httpServer := &serveStub{err: errors.New("http stopped"), started: make(chan struct{})}
	smtpServer := &serveStub{err: errors.New("smtp stopped"), started: make(chan struct{})}

	done := make(chan error, 1)
	go func() {
		done <- runServers(httpServer, smtpServer)
	}()

	<-httpServer.started
	<-smtpServer.started
	err := <-done
	if err == nil {
		t.Fatal("expected runServers to return an error")
	}

	if httpServer.calls != 1 {
		t.Fatalf("expected http server to start once, got %d", httpServer.calls)
	}

	if smtpServer.calls != 1 {
		t.Fatalf("expected smtp server to start once, got %d", smtpServer.calls)
	}
}

func TestRunServersReturnsFirstServerError(t *testing.T) {
	want := errors.New("smtp failed first")
	httpServer := &serveStub{err: errors.New("http failed second"), delay: 20 * time.Millisecond}
	smtpServer := &serveStub{err: want, delay: 1 * time.Millisecond}

	err := runServers(httpServer, smtpServer)
	if !errors.Is(err, want) {
		t.Fatalf("expected first error %v, got %v", want, err)
	}
}

type serveStub struct {
	calls   int
	err     error
	delay   time.Duration
	started chan struct{}
}

func (s *serveStub) ListenAndServe() error {
	s.calls++
	if s.started != nil {
		close(s.started)
	}
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.err
}

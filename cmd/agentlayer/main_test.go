package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/app"
	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/outbound"
)

func TestNewServerExposesHealthEndpoint(t *testing.T) {
	runtimeStore = newRuntimeStore()
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
	runtimeStore = newRuntimeStore()
	server := newServer()

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/bootstrap"},
		{method: http.MethodPost, path: "/bootstrap"},
		{method: http.MethodGet, path: "/webhooks/deliveries"},
		{method: http.MethodGet, path: "/webhooks/deliveries/delivery-123"},
		{method: http.MethodPost, path: "/webhooks/deliveries/delivery-123/replay"},
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
	runtimeStore = newRuntimeStore()
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected real thread handler path to return 500 from placeholder service, got %d", recorder.Code)
	}
}

func TestNewServerWiresContactHandler(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/contacts/contact-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected real contact handler path to return 500 from placeholder service, got %d", recorder.Code)
	}
}

func TestNewServerWiresRemainingHandlers(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodGet, path: "/bootstrap", want: http.StatusOK},
		{method: http.MethodPost, path: "/bootstrap", body: "{}", want: http.StatusCreated},
		{method: http.MethodGet, path: "/webhooks/deliveries", want: http.StatusOK},
		{method: http.MethodGet, path: "/webhooks/deliveries/delivery-123", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/webhooks/deliveries/delivery-123/replay", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/threads/thread-123/reply", body: "{}", want: http.StatusInternalServerError},
		{method: http.MethodPost, path: "/threads/thread-123/escalate", body: "{}", want: http.StatusAccepted},
		{method: http.MethodGet, path: "/threads/thread-123/messages", want: http.StatusOK},
		{method: http.MethodPost, path: "/contacts/contact-123/memory", body: "{}", want: http.StatusCreated},
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
	runtimeStore = newRuntimeStore()
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
	runtimeStore = newRuntimeStore()
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
	runtimeStore = newRuntimeStore()
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
	runtimeStore = newRuntimeStore()
	service := newInboundService()

	_, err := service.HandleStoredMessage(context.Background(), core.StoredInboundMessage{})
	if err == nil {
		t.Fatal("expected inbound service to fail until raw mime is present in the runtime store")
	}
}

func TestNewInboundProcessorUsesRealProcessorChain(t *testing.T) {
	runtimeStore = newRuntimeStore()
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
	runtimeStore = newRuntimeStore()
	recorder := newInboundRecorder()

	result, err := recorder.Record(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			OrganizationID: "org-local",
			InboxID:        "inbox-local",
			ReceivedAt:     time.Date(2026, 4, 3, 17, 5, 0, 0, time.UTC),
		},
	}, inbound.ProcessResult{
		Contact: domain.Contact{ID: "contact-123", EmailAddress: "sender@example.com"},
		Thread: domain.Thread{
			ID:    "thread-123",
			State: domain.ThreadStateActive,
		},
	})
	if err != nil {
		t.Fatalf("expected runtime store-backed inbound recorder to succeed, got error: %v", err)
	}

	if result.Message.ThreadID != "thread-123" {
		t.Fatalf("expected recorded inbound message, got %#v", result.Message)
	}
}

func TestNewThreadReadServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	_, err := runtimeStore.Save(context.Background(), domain.Thread{
		ID:             "thread-read-123",
		OrganizationID: "org-local",
		State:          domain.ThreadStateActive,
	})
	if err != nil {
		t.Fatalf("expected thread seed to succeed, got error: %v", err)
	}

	service := newThreadReadService()

	thread, err := service.GetThread(context.Background(), "thread-read-123")
	if err != nil {
		t.Fatalf("expected runtime store-backed thread read to succeed, got error: %v", err)
	}

	if thread.ID != "thread-read-123" {
		t.Fatalf("expected returned thread, got %#v", thread)
	}
}

func TestNewContactReadServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	_, err := runtimeStore.UpsertByEmail(context.Background(), domain.Contact{
		ID:           "contact-read-123",
		EmailAddress: "reader@example.com",
	})
	if err != nil {
		t.Fatalf("expected contact seed to succeed, got error: %v", err)
	}

	service := newContactReadService()

	contact, err := service.GetContact(context.Background(), "contact-read-123")
	if err != nil {
		t.Fatalf("expected runtime store-backed contact read to succeed, got error: %v", err)
	}

	if contact.ID != "contact-read-123" {
		t.Fatalf("expected returned contact, got %#v", contact)
	}
}

func TestNewThreadEscalationServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	service := newThreadEscalationService()

	thread, err := service.EscalateThread(context.Background(), "thread-escalate-123", "needs human review")
	if err != nil {
		t.Fatalf("expected runtime store-backed escalation to succeed, got error: %v", err)
	}

	if thread.State != domain.ThreadStateEscalated {
		t.Fatalf("expected escalated thread, got %#v", thread)
	}
}

func TestNewContactMemoryServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	service := newContactMemoryService()

	entry, err := service.CreateContactMemory(context.Background(), "contact-123", api.CreateContactMemoryInput{
		ThreadID: "thread-123",
		Note:     "Prefers email follow-up.",
		Tags:     []string{"preference"},
	})
	if err != nil {
		t.Fatalf("expected runtime store-backed contact memory write to succeed, got error: %v", err)
	}

	if entry.ID == "" {
		t.Fatalf("expected created memory entry, got %#v", entry)
	}
}

func TestNewBootstrapServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	service := newBootstrapService()

	result, err := service.BootstrapLocal(context.Background(), app.BootstrapInput{
		OrganizationName: "Acme Support",
		AgentName:        "Acme Agent",
		AgentStatus:      domain.AgentStatusPaused,
		WebhookURL:       "https://example.com/webhook",
		WebhookSecret:    "super-secret",
		InboxAddress:     "agent@example.com",
		InboxDomain:      "example.com",
		InboxDisplayName: "Acme Inbox",
	})
	if err != nil {
		t.Fatalf("expected runtime store-backed bootstrap to succeed, got error: %v", err)
	}

	if result.Agent.WebhookURL != "https://example.com/webhook" {
		t.Fatalf("expected webhook url to be persisted, got %#v", result.Agent)
	}

	inbox, err := runtimeStore.GetInboxByID(context.Background(), "inbox-local")
	if err != nil {
		t.Fatalf("expected persisted inbox lookup to succeed, got error: %v", err)
	}

	if inbox.EmailAddress != "agent@example.com" {
		t.Fatalf("expected bootstrapped inbox address, got %#v", inbox)
	}
}

func TestNewBootstrapReadServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()

	_, err := newBootstrapService().BootstrapLocal(context.Background(), app.BootstrapInput{
		OrganizationName: "Acme Support",
		AgentName:        "Acme Agent",
		WebhookURL:       "https://example.com/webhook",
		InboxAddress:     "agent@example.com",
		InboxDomain:      "example.com",
		InboxDisplayName: "Acme Inbox",
	})
	if err != nil {
		t.Fatalf("expected bootstrap seed to succeed, got error: %v", err)
	}

	result, err := newBootstrapReadService().GetBootstrap(context.Background())
	if err != nil {
		t.Fatalf("expected bootstrap read to succeed, got error: %v", err)
	}

	if result.Organization.Name != "Acme Support" {
		t.Fatalf("expected organization name from store, got %#v", result.Organization)
	}

	if result.Agent.WebhookURL != "https://example.com/webhook" {
		t.Fatalf("expected webhook url from store, got %#v", result.Agent)
	}
}

func TestNewWebhookDeliveryReadServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	_, err := runtimeStore.SaveWebhookDelivery(context.Background(), domain.WebhookDelivery{
		ID:           "delivery-123",
		EventID:      "event-123",
		EventType:    "message.received",
		Status:       "failed",
		AttemptCount: 2,
		ResponseCode: 500,
	})
	if err != nil {
		t.Fatalf("expected webhook delivery seed to succeed, got error: %v", err)
	}

	delivery, err := newWebhookDeliveryReadService().GetWebhookDelivery(context.Background(), "delivery-123")
	if err != nil {
		t.Fatalf("expected webhook delivery read to succeed, got error: %v", err)
	}

	if delivery.ID != "delivery-123" || delivery.ResponseCode != 500 {
		t.Fatalf("expected loaded webhook delivery, got %#v", delivery)
	}
}

func TestNewWebhookDeliveryListServiceUsesApplicationService(t *testing.T) {
	runtimeStore = newRuntimeStore()
	_, _ = runtimeStore.SaveWebhookDelivery(context.Background(), domain.WebhookDelivery{
		ID:        "delivery-older",
		EventID:   "event-older",
		UpdatedAt: time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC),
	})
	_, _ = runtimeStore.SaveWebhookDelivery(context.Background(), domain.WebhookDelivery{
		ID:        "delivery-newer",
		EventID:   "event-newer",
		UpdatedAt: time.Date(2026, 4, 3, 11, 0, 0, 0, time.UTC),
	})

	deliveries, err := newWebhookDeliveryListService().ListWebhookDeliveries(context.Background())
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 2 || deliveries[0].ID != "delivery-newer" {
		t.Fatalf("expected recency-ordered webhook deliveries, got %#v", deliveries)
	}
}

func TestNewReplyServiceUsesRealOutboundComposition(t *testing.T) {
	runtimeStore = newRuntimeStore()
	service := newReplyService()

	result, err := service.SendReply(context.Background(), outbound.SendReplyInput{
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
	if err != nil {
		t.Fatalf("expected runtime store-backed outbound reply to succeed, got error: %v", err)
	}

	if result.Message.DeliveryState != "sent" {
		t.Fatalf("expected sent outbound message, got %#v", result.Message)
	}

	if result.SendResult.ProviderMessageID == "" {
		t.Fatalf("expected provider result to include a provider message id, got %#v", result.SendResult)
	}
}

func TestNewOutboundCallbackFlowUsesRealCallbackComposition(t *testing.T) {
	runtimeStore = newRuntimeStore()
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

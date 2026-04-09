package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	"github.com/agentlayer/agentlayer/internal/smtpedge"
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
		{method: http.MethodGet, path: "/inbound/receipts/list"},
		{method: http.MethodGet, path: "/inbound/receipts"},
		{method: http.MethodPost, path: "/inbound/reprocess"},
		{method: http.MethodGet, path: "/webhooks/deliveries"},
		{method: http.MethodPost, path: "/threads/thread-123/reply"},
		{method: http.MethodPost, path: "/threads/thread-123/escalate"},
		{method: http.MethodGet, path: "/threads/thread-123/messages"},
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

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected real thread handler path to return 404 for missing thread, got %d", recorder.Code)
	}
}

func TestNewServerWiresContactHandler(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()
	request := httptest.NewRequest(http.MethodGet, "/contacts/contact-123", nil)
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected real contact handler path to return 404 for missing contact, got %d", recorder.Code)
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
		{method: http.MethodPost, path: "/bootstrap", body: "{}", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/inbound/receipts/list", want: http.StatusOK},
		{method: http.MethodGet, path: "/inbound/receipts", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/inbound/reprocess", body: "{}", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/webhooks/deliveries", want: http.StatusOK},
		{method: http.MethodGet, path: "/webhooks/deliveries/delivery-123", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/webhooks/deliveries/delivery-123/replay", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/threads/thread-123/reply", body: "{}", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/threads/thread-123/escalate", body: "{}", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/threads/thread-123/messages", want: http.StatusOK},
		{method: http.MethodPost, path: "/contacts/contact-123/memory", body: "{}", want: http.StatusNotFound},
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

func TestRuntimeEnvHelpers(t *testing.T) {
	t.Setenv("AGENTLAYER_DATABASE_URL", "postgres://agentlayer:agentlayer@localhost:5432/agentlayer?sslmode=disable")
	t.Setenv("AGENTLAYER_RAW_DATA_DIR", "/tmp/agentlayer-raw")
	t.Setenv("AGENTLAYER_AUTO_MIGRATE", "true")

	if got := databaseURL(); got == "" {
		t.Fatal("expected database url helper to read env")
	}
	if got := rawDataDir(); got != "/tmp/agentlayer-raw" {
		t.Fatalf("expected raw data dir helper to read env, got %q", got)
	}
	if !autoMigrateEnabled() {
		t.Fatal("expected auto migrate helper to be enabled")
	}
}

func TestSMTPReceiptIdentifierHelpers(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 34, 56, 123456789, time.UTC)

	sessionID := smtpedge.NewSessionID(now)
	if !strings.HasPrefix(sessionID, "smtp-20260409T123456.123456789Z-") {
		t.Fatalf("expected generated smtp session id, got %q", sessionID)
	}

	objectKey := smtpedge.NewRawMessageObjectKey(now, domain.Inbox{ID: "inbox-local"})
	if !strings.HasPrefix(objectKey, "raw/2026/04/09/inbox-local/") {
		t.Fatalf("expected generated raw message object key, got %q", objectKey)
	}
	if !strings.HasSuffix(objectKey, ".eml") {
		t.Fatalf("expected eml suffix for raw object key, got %q", objectKey)
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

	deliveries, err := newWebhookDeliveryListService().ListWebhookDeliveries(context.Background(), 0)
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 2 || deliveries[0].ID != "delivery-newer" {
		t.Fatalf("expected recency-ordered webhook deliveries, got %#v", deliveries)
	}
}

func TestBootstrapAndInboundFlowPersistWebhookDelivery(t *testing.T) {
	runtimeStore = newRuntimeStore()
	receivedCh, webhookServer := newWebhookCaptureServer()
	defer webhookServer.Close()

	server := newServer()
	bootstrapLocalRuntime(t, server, webhookServer.URL, "active")

	result, err := handleTestInboundMessage(t, "raw/integration-message.eml")
	if err != nil {
		t.Fatalf("expected inbound runtime flow to succeed, got error: %v", err)
	}

	if result.Message.ID == "" || result.Thread.ID == "" {
		t.Fatalf("expected handled inbound result to persist message and thread, got %#v", result)
	}

	received := waitForWebhook(t, receivedCh)
	if received.Headers.Get("X-AgentLayer-Signature") == "" {
		t.Fatalf("expected signed webhook headers, got %#v", received.Headers)
	}

	var payload struct {
		EventType string `json:"event_type"`
		Message   struct {
			ID string `json:"id"`
		} `json:"message"`
	}
	if err := json.Unmarshal(received.Body, &payload); err != nil {
		t.Fatalf("expected webhook payload json, got error: %v", err)
	}

	if payload.EventType != "message.received" || payload.Message.ID == "" {
		t.Fatalf("expected message.received webhook payload, got %#v", payload)
	}

	deliveries, err := runtimeStore.ListWebhookDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected webhook deliveries to be persisted, got error: %v", err)
	}

	if len(deliveries) != 1 || deliveries[0].Status != "succeeded" {
		t.Fatalf("expected succeeded webhook delivery record, got %#v", deliveries)
	}

	if deliveries[0].RequestURL != webhookServer.URL {
		t.Fatalf("expected persisted request url, got %#v", deliveries[0])
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries", nil)
	readRecorder := httptest.NewRecorder()
	server.ServeHTTP(readRecorder, readRequest)

	if readRecorder.Code != http.StatusOK {
		t.Fatalf("expected webhook delivery list endpoint to succeed, got %d", readRecorder.Code)
	}

	var response []map[string]any
	if err := json.Unmarshal(readRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected webhook delivery list json, got error: %v", err)
	}

	if len(response) != 1 {
		t.Fatalf("expected one webhook delivery from admin endpoint, got %#v", response)
	}
}

func TestBootstrapAndInboundFlowSkipsWebhookWhenAgentPaused(t *testing.T) {
	runtimeStore = newRuntimeStore()
	receivedCh, webhookServer := newWebhookCaptureServer()
	defer webhookServer.Close()

	server := newServer()
	bootstrapLocalRuntime(t, server, webhookServer.URL, "paused")

	if _, err := handleTestInboundMessage(t, "raw/paused-message.eml"); err != nil {
		t.Fatalf("expected paused-agent inbound flow to succeed, got error: %v", err)
	}

	select {
	case received := <-receivedCh:
		t.Fatalf("expected paused agent to skip webhook delivery, got %#v", received)
	case <-time.After(100 * time.Millisecond):
	}

	deliveries, err := runtimeStore.ListWebhookDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 0 {
		t.Fatalf("expected paused agent to persist no webhook deliveries, got %#v", deliveries)
	}
}

func TestBootstrapAndInboundFlowDeduplicatesRepeatedInboundMessage(t *testing.T) {
	runtimeStore = newRuntimeStore()
	receivedCh, webhookServer := newWebhookCaptureServer()
	defer webhookServer.Close()

	server := newServer()
	bootstrapLocalRuntime(t, server, webhookServer.URL, "active")

	first, err := handleTestInboundMessage(t, "raw/duplicate-message-1.eml")
	if err != nil {
		t.Fatalf("expected first inbound runtime flow to succeed, got error: %v", err)
	}

	if first.Duplicate {
		t.Fatalf("expected first inbound result not to be duplicate, got %#v", first)
	}

	received := waitForWebhook(t, receivedCh)
	if received.Headers.Get("X-AgentLayer-Signature") == "" {
		t.Fatalf("expected first webhook delivery to be signed, got %#v", received.Headers)
	}

	second, err := handleTestInboundMessage(t, "raw/duplicate-message-2.eml")
	if err != nil {
		t.Fatalf("expected duplicate inbound runtime flow to succeed, got error: %v", err)
	}

	if !second.Duplicate {
		t.Fatalf("expected duplicate inbound result to be marked duplicate, got %#v", second)
	}

	if second.Message.ID != first.Message.ID {
		t.Fatalf("expected duplicate inbound to reuse stored message, got first=%#v second=%#v", first.Message, second.Message)
	}

	select {
	case duplicateWebhook := <-receivedCh:
		t.Fatalf("expected duplicate inbound message to skip webhook redelivery, got %#v", duplicateWebhook)
	case <-time.After(100 * time.Millisecond):
	}

	messages, err := runtimeStore.ListByThreadID(context.Background(), first.Thread.ID, 10)
	if err != nil {
		t.Fatalf("expected thread messages lookup to succeed, got error: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected one persisted inbound message after duplicate processing, got %#v", messages)
	}

	deliveries, err := runtimeStore.ListWebhookDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 1 {
		t.Fatalf("expected one webhook delivery after duplicate processing, got %#v", deliveries)
	}
}

func TestInboundReprocessEndpointReusesStoredReceiptWithoutRedeliveringWebhook(t *testing.T) {
	runtimeStore = newRuntimeStore()
	receivedCh, webhookServer := newWebhookCaptureServer()
	defer webhookServer.Close()

	server := newServer()
	bootstrapLocalRuntime(t, server, webhookServer.URL, "active")

	first, err := handleTestInboundMessage(t, "raw/reprocess-message.eml")
	if err != nil {
		t.Fatalf("expected initial inbound handling to succeed, got error: %v", err)
	}

	_ = waitForWebhook(t, receivedCh)

	request := httptest.NewRequest(http.MethodPost, "/inbound/reprocess", bytes.NewBufferString(`{
		"object_key":"raw/reprocess-message.eml"
	}`))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected reprocess endpoint to succeed, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected reprocess response json, got error: %v", err)
	}

	if response["message_id"] != first.Message.ID || response["duplicate"] != true {
		t.Fatalf("expected duplicate reprocess response, got %#v", response)
	}

	select {
	case duplicateWebhook := <-receivedCh:
		t.Fatalf("expected reprocess to skip webhook redelivery, got %#v", duplicateWebhook)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestInboundReceiptEndpointLoadsStoredReceipt(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

	if _, err := handleTestInboundMessage(t, "raw/receipt-message.eml"); err != nil {
		t.Fatalf("expected inbound handling to succeed, got error: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/inbound/receipts?object_key=raw/receipt-message.eml", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected inbound receipt endpoint to succeed, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected receipt response json, got error: %v", err)
	}

	if response["raw_message_object_key"] != "raw/receipt-message.eml" || response["smtp_transaction_id"] != "smtp-test-session" {
		t.Fatalf("expected stored inbound receipt response, got %#v", response)
	}
}

func TestInboundReceiptsEndpointListsStoredReceiptsWithLimit(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

	if _, err := handleTestInboundMessage(t, "raw/receipt-list-1.eml"); err != nil {
		t.Fatalf("expected first inbound handling to succeed, got error: %v", err)
	}

	if err := runtimeStore.SaveInboundReceipt(context.Background(), inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-test-session-2",
		OrganizationID:      "org-local",
		AgentID:             "agent-local",
		InboxID:             "inbox-local",
		EnvelopeSender:      "sender2@example.com",
		EnvelopeRecipients:  []string{"agent@localhost"},
		RawMessageObjectKey: "raw/receipt-list-2.eml",
		ReceivedAt:          time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("expected second inbound receipt seed to succeed, got error: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/inbound/receipts/list?limit=1", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected inbound receipts endpoint to succeed, got %d", recorder.Code)
	}

	var response []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected receipt list response json, got error: %v", err)
	}

	if len(response) != 1 || response[0]["raw_message_object_key"] != "raw/receipt-list-2.eml" {
		t.Fatalf("expected limited most-recent inbound receipt list, got %#v", response)
	}
}

func TestWebhookReplayFlowReplaysStoredDelivery(t *testing.T) {
	runtimeStore = newRuntimeStore()
	receivedCh, webhookServer := newWebhookCaptureServer()
	defer webhookServer.Close()

	server := newServer()
	bootstrapLocalRuntime(t, server, webhookServer.URL, "active")

	if _, err := handleTestInboundMessage(t, "raw/replay-message.eml"); err != nil {
		t.Fatalf("expected inbound runtime flow to succeed, got error: %v", err)
	}

	first := waitForWebhook(t, receivedCh)
	if first.Headers.Get("X-AgentLayer-Signature") == "" {
		t.Fatalf("expected initial webhook delivery to be signed, got %#v", first.Headers)
	}

	deliveries, err := runtimeStore.ListWebhookDeliveries(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected webhook delivery list to succeed, got error: %v", err)
	}

	if len(deliveries) != 1 {
		t.Fatalf("expected one stored webhook delivery, got %#v", deliveries)
	}

	replayRequest := httptest.NewRequest(http.MethodPost, "/webhooks/deliveries/"+deliveries[0].ID+"/replay", nil)
	replayRecorder := httptest.NewRecorder()
	server.ServeHTTP(replayRecorder, replayRequest)

	if replayRecorder.Code != http.StatusAccepted {
		t.Fatalf("expected replay endpoint to succeed, got %d", replayRecorder.Code)
	}

	second := waitForWebhook(t, receivedCh)
	if second.Headers.Get("X-AgentLayer-Signature") == "" {
		t.Fatalf("expected replayed webhook delivery to be signed, got %#v", second.Headers)
	}

	updated, err := runtimeStore.GetWebhookDeliveryByID(context.Background(), deliveries[0].ID)
	if err != nil {
		t.Fatalf("expected stored webhook delivery lookup to succeed, got error: %v", err)
	}

	if updated.AttemptCount != 2 || updated.Status != "succeeded" {
		t.Fatalf("expected replay to update attempt count, got %#v", updated)
	}

	showRequest := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries/"+updated.ID, nil)
	showRecorder := httptest.NewRecorder()
	server.ServeHTTP(showRecorder, showRequest)

	if showRecorder.Code != http.StatusOK {
		t.Fatalf("expected webhook delivery read endpoint to succeed, got %d", showRecorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(showRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected webhook delivery read json, got error: %v", err)
	}

	if got, ok := response["attempt_count"].(float64); !ok || got != 2 {
		t.Fatalf("expected replayed attempt count in response, got %#v", response)
	}
}

func TestWebhookDeliveryListEndpointHonorsLimitIntegration(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

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

	request := httptest.NewRequest(http.MethodGet, "/webhooks/deliveries?limit=1", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected webhook delivery list endpoint to succeed, got %d", recorder.Code)
	}

	var response []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected webhook delivery list json, got error: %v", err)
	}

	if len(response) != 1 || response[0]["id"] != "delivery-newer" {
		t.Fatalf("expected limited recency-ordered response, got %#v", response)
	}
}

func TestThreadMessagesEndpointHonorsLimitIntegration(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

	_, _ = runtimeStore.Create(context.Background(), domain.Message{
		ID:        "message-older",
		ThreadID:  "thread-123",
		Direction: domain.MessageDirectionInbound,
		Subject:   "Older",
	})
	_, _ = runtimeStore.Create(context.Background(), domain.Message{
		ID:        "message-newer",
		ThreadID:  "thread-123",
		Direction: domain.MessageDirectionOutbound,
		Subject:   "Newer",
	})

	request := httptest.NewRequest(http.MethodGet, "/threads/thread-123/messages?limit=1", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected thread messages endpoint to succeed, got %d", recorder.Code)
	}

	var response []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected thread messages json, got error: %v", err)
	}

	if len(response) != 1 || response[0]["id"] != "message-older" {
		t.Fatalf("expected limited thread message response, got %#v", response)
	}
}

func TestOutboundCallbackEndpointAppliesDeliveredStateIntegration(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

	_, err := runtimeStore.SaveMessage(context.Background(), domain.Message{
		ID:                "message-123",
		ThreadID:          "thread-123",
		Direction:         domain.MessageDirectionOutbound,
		ProviderMessageID: "ses-123",
		DeliveryState:     "sent",
	})
	if err != nil {
		t.Fatalf("expected outbound message seed to succeed, got error: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{
		"event_type":"delivered",
		"provider_message_id":"ses-123",
		"occurred_at":"2026-04-03T23:05:00Z",
		"contact_email":"sender@example.com"
	}`))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected outbound callback endpoint to succeed, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected outbound callback response json, got error: %v", err)
	}

	if response["delivery_state"] != "delivered" {
		t.Fatalf("expected delivered state in callback response, got %#v", response)
	}

	message, found, err := runtimeStore.FindByProviderMessageID(context.Background(), "ses-123")
	if err != nil || !found {
		t.Fatalf("expected updated outbound message lookup, got found=%v err=%v", found, err)
	}

	if message.DeliveryState != "delivered" {
		t.Fatalf("expected outbound message delivery state to be updated, got %#v", message)
	}
}

func TestOutboundCallbackEndpointAppliesSuppressionIntegration(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()

	_, err := runtimeStore.SaveMessage(context.Background(), domain.Message{
		ID:                "message-123",
		ThreadID:          "thread-123",
		Direction:         domain.MessageDirectionOutbound,
		ProviderMessageID: "ses-bounce-123",
		DeliveryState:     "sent",
	})
	if err != nil {
		t.Fatalf("expected outbound message seed to succeed, got error: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/provider/callbacks/outbound", bytes.NewBufferString(`{
		"event_type":"hard_bounce",
		"provider_message_id":"ses-bounce-123",
		"occurred_at":"2026-04-03T23:06:00Z",
		"contact_email":"sender@example.com"
	}`))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected hard bounce callback endpoint to succeed, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected callback response json, got error: %v", err)
	}

	if response["suppressed"] != true {
		t.Fatalf("expected suppression to be applied, got %#v", response)
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

func TestNewReplyHandlerServiceLoadsRuntimeState(t *testing.T) {
	runtimeStore = newRuntimeStore()
	seedReplyRuntimeState(t)

	service := newReplyHandlerService()
	result, err := service.SendReply(context.Background(), outbound.SendReplyInput{
		Thread:         domain.Thread{ID: "thread-123"},
		ReplyToMessage: domain.Message{ID: "message-inbound-123"},
		BodyText:       "Thanks for reaching out.",
		ObjectKey:      "outbound/reply-123.eml",
	})
	if err != nil {
		t.Fatalf("expected reply handler service to succeed, got error: %v", err)
	}

	if result.Message.ThreadID != "thread-123" || result.Message.DeliveryState != "sent" {
		t.Fatalf("expected sent reply message, got %#v", result.Message)
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

func TestReplyEndpointSendsReplyIntegration(t *testing.T) {
	runtimeStore = newRuntimeStore()
	server := newServer()
	seedReplyRuntimeState(t)

	request := httptest.NewRequest(http.MethodPost, "/threads/thread-123/reply", bytes.NewBufferString(`{
		"reply_to_message_id":"message-inbound-123",
		"body_text":"Thanks for reaching out.",
		"object_key":"outbound/reply-http-123.eml"
	}`))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected reply endpoint to succeed, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected reply response json, got error: %v", err)
	}

	if response["delivery_state"] != "sent" {
		t.Fatalf("expected sent reply response, got %#v", response)
	}

	messages, err := runtimeStore.ListByThreadID(context.Background(), "thread-123", 10)
	if err != nil {
		t.Fatalf("expected thread message list to succeed, got error: %v", err)
	}

	if len(messages) < 2 {
		t.Fatalf("expected reply message to be persisted, got %#v", messages)
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

type receivedWebhook struct {
	Headers http.Header
	Body    []byte
}

func newWebhookCaptureServer() (chan receivedWebhook, *httptest.Server) {
	receivedCh := make(chan receivedWebhook, 4)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		receivedCh <- receivedWebhook{
			Headers: request.Header.Clone(),
			Body:    body,
		}
		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	return receivedCh, server
}

func bootstrapLocalRuntime(t *testing.T, server http.Handler, webhookURL, status string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/bootstrap", bytes.NewBufferString(`{
		"organization_name":"Acme Support",
		"agent_name":"Acme Agent",
		"agent_status":"`+status+`",
		"webhook_url":"`+webhookURL+`",
		"webhook_secret":"dev-secret",
		"inbox_address":"agent@localhost",
		"inbox_domain":"localhost",
		"inbox_display_name":"Acme Inbox"
	}`))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected bootstrap to succeed, got %d", recorder.Code)
	}
}

func handleTestInboundMessage(t *testing.T, objectKey string) (inbound.HandleResult, error) {
	t.Helper()

	raw := "From: Sender Example <sender@example.com>\r\n" +
		"To: Acme Inbox <agent@localhost>\r\n" +
		"Subject: Hello World\r\n" +
		"Message-ID: <message-123@example.com>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"Plain body.\r\n"

	if err := runtimeStore.Put(context.Background(), objectKey, []byte(raw)); err != nil {
		t.Fatalf("expected raw message seed to succeed, got error: %v", err)
	}

	receipt := inbound.DurableReceiptRequest{
		SMTPTransactionID:   "smtp-test-session",
		OrganizationID:      "org-local",
		AgentID:             "agent-local",
		InboxID:             "inbox-local",
		EnvelopeSender:      "sender@example.com",
		EnvelopeRecipients:  []string{"agent@localhost"},
		RawMessageObjectKey: objectKey,
		ReceivedAt:          time.Date(2026, 4, 3, 23, 0, 0, 0, time.UTC),
	}
	if err := runtimeStore.SaveInboundReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("expected inbound receipt seed to succeed, got error: %v", err)
	}

	return newInboundService().HandleStoredMessage(context.Background(), core.StoredInboundMessage{
		Receipt: core.InboundReceipt{
			SMTPTransactionID:   receipt.SMTPTransactionID,
			OrganizationID:      receipt.OrganizationID,
			AgentID:             receipt.AgentID,
			InboxID:             receipt.InboxID,
			EnvelopeSender:      receipt.EnvelopeSender,
			EnvelopeRecipients:  receipt.EnvelopeRecipients,
			RawMessageObjectKey: receipt.RawMessageObjectKey,
			ReceivedAt:          receipt.ReceivedAt,
		},
	})
}

func seedReplyRuntimeState(t *testing.T) {
	t.Helper()

	_, err := runtimeStore.UpsertByEmail(context.Background(), domain.Contact{
		ID:           "contact-123",
		EmailAddress: "sender@example.com",
		DisplayName:  "Sender Example",
	})
	if err != nil {
		t.Fatalf("expected contact seed to succeed, got error: %v", err)
	}

	_, err = runtimeStore.Save(context.Background(), domain.Thread{
		ID:             "thread-123",
		OrganizationID: "org-local",
		AgentID:        "agent-local",
		InboxID:        "inbox-local",
		ContactID:      "contact-123",
		State:          domain.ThreadStateActive,
	})
	if err != nil {
		t.Fatalf("expected thread seed to succeed, got error: %v", err)
	}

	_, err = runtimeStore.Create(context.Background(), domain.Message{
		ID:              "message-inbound-123",
		OrganizationID:  "org-local",
		ThreadID:        "thread-123",
		InboxID:         "inbox-local",
		ContactID:       "contact-123",
		Direction:       domain.MessageDirectionInbound,
		Subject:         "Hello World",
		MessageIDHeader: "<message-inbound-123@example.com>",
		TextBody:        "Inbound message",
		CreatedAt:       time.Date(2026, 4, 3, 22, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected inbound message seed to succeed, got error: %v", err)
	}
}

func waitForWebhook(t *testing.T, receivedCh <-chan receivedWebhook) receivedWebhook {
	t.Helper()

	select {
	case received := <-receivedCh:
		return received
	case <-time.After(1 * time.Second):
		t.Fatal("expected webhook delivery to be received")
		return receivedWebhook{}
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

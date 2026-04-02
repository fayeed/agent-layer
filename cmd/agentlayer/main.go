package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/outbound"
	"github.com/agentlayer/agentlayer/internal/smtpedge"
	smtp "github.com/emersion/go-smtp"
)

func main() {
	httpAddr := serverAddress()
	smtpServer := newSMTPServer()
	httpServer := newHTTPServer(httpAddr, newServer())

	log.Printf("agentlayer http listening on %s", httpAddr)
	log.Printf("agentlayer smtp configured on %s", smtpServer.Addr)

	if err := runServers(httpServer, smtpServer); err != nil {
		log.Fatal(err)
	}
}

type serveServer interface {
	ListenAndServe() error
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.Handle("POST /threads/{threadID}/reply", api.NewReplyHandler(notImplementedReplyService{}))
	mux.Handle("POST /threads/{threadID}/escalate", api.NewThreadEscalateHandler(notImplementedThreadEscalationService{}))
	mux.Handle("GET /threads/{threadID}", api.NewThreadHandler(notImplementedThreadService{}))
	mux.Handle("GET /threads/{threadID}/messages", api.NewThreadMessagesHandler(notImplementedThreadMessagesService{}))
	mux.Handle("GET /contacts/{contactID}", api.NewContactHandler(notImplementedContactService{}))
	mux.Handle("POST /contacts/{contactID}/memory", api.NewContactMemoryHandler(notImplementedContactMemoryService{}))
	mux.Handle("POST /provider/callbacks/outbound", api.NewOutboundCallbackHandler(notImplementedOutboundCallbackParser{}, notImplementedOutboundCallbackFlow{}))
	return mux
}

func handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok\n"))
}

func handleNotImplemented(writer http.ResponseWriter, _ *http.Request) {
	http.Error(writer, "service wiring not implemented", http.StatusNotImplemented)
}

func serverAddress() string {
	if value := os.Getenv("AGENTLAYER_ADDR"); value != "" {
		return value
	}
	return ":8080"
}

func smtpAddress() string {
	if value := os.Getenv("AGENTLAYER_SMTP_ADDR"); value != "" {
		return value
	}
	return "localhost:2525"
}

func smtpDomain() string {
	if value := os.Getenv("AGENTLAYER_SMTP_DOMAIN"); value != "" {
		return value
	}
	return "localhost"
}

func newSMTPServer() *smtp.Server {
	return smtpedge.NewServer(
		smtpedge.NewBackend(func() smtpedge.CoreSession {
			session := smtpedge.NewSession(
				notImplementedInboxLookup{},
				notImplementedRawMessageStore{},
				notImplementedReceiptSink{},
				time.Now,
				func() string { return "raw/generated.eml" },
				"smtp-session-placeholder",
			)
			return &session
		}),
		smtpedge.Config{
			Addr:   smtpAddress(),
			Domain: smtpDomain(),
		},
	)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

func runServers(httpServer serveServer, smtpServer serveServer) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	go func() {
		errCh <- smtpServer.ListenAndServe()
	}()

	return <-errCh
}

type notImplementedThreadService struct{}

func (notImplementedThreadService) GetThread(context.Context, string) (domain.Thread, error) {
	return domain.Thread{}, errors.New("thread service not implemented")
}

type notImplementedContactService struct{}

func (notImplementedContactService) GetContact(context.Context, string) (domain.Contact, error) {
	return domain.Contact{}, errors.New("contact service not implemented")
}

type notImplementedReplyService struct{}

func (notImplementedReplyService) SendReply(context.Context, outbound.SendReplyInput) (outbound.SendReplyResult, error) {
	return outbound.SendReplyResult{}, errors.New("reply service not implemented")
}

type notImplementedThreadEscalationService struct{}

func (notImplementedThreadEscalationService) EscalateThread(context.Context, string, string) (domain.Thread, error) {
	return domain.Thread{}, errors.New("thread escalation service not implemented")
}

type notImplementedThreadMessagesService struct{}

func (notImplementedThreadMessagesService) ListThreadMessages(context.Context, string) ([]domain.Message, error) {
	return nil, errors.New("thread messages service not implemented")
}

type notImplementedContactMemoryService struct{}

func (notImplementedContactMemoryService) CreateContactMemory(context.Context, string, api.CreateContactMemoryInput) (domain.ContactMemoryEntry, error) {
	return domain.ContactMemoryEntry{}, errors.New("contact memory service not implemented")
}

type notImplementedOutboundCallbackParser struct{}

func (notImplementedOutboundCallbackParser) Parse([]byte) (outbound.DeliveryCallbackEvent, error) {
	return outbound.DeliveryCallbackEvent{}, errors.New("outbound callback parser not implemented")
}

type notImplementedOutboundCallbackFlow struct{}

func (notImplementedOutboundCallbackFlow) Apply(context.Context, outbound.CallbackFlowInput) (outbound.CallbackFlowResult, error) {
	return outbound.CallbackFlowResult{}, errors.New("outbound callback flow not implemented")
}

type notImplementedInboxLookup struct{}

func (notImplementedInboxLookup) FindByEmailAddress(context.Context, string) (domain.Inbox, bool, error) {
	return domain.Inbox{}, false, errors.New("smtp inbox lookup not implemented")
}

type notImplementedRawMessageStore struct{}

func (notImplementedRawMessageStore) Put(context.Context, string, []byte) error {
	return errors.New("smtp raw message store not implemented")
}

type notImplementedReceiptSink struct{}

func (notImplementedReceiptSink) Enqueue(context.Context, inbound.DurableReceiptRequest) error {
	return errors.New("smtp receipt sink not implemented")
}

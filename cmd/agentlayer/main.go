package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/app"
	"github.com/agentlayer/agentlayer/internal/contacts"
	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/outbound"
	"github.com/agentlayer/agentlayer/internal/parser"
	"github.com/agentlayer/agentlayer/internal/smtpedge"
	"github.com/agentlayer/agentlayer/internal/threading"
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
	mux.Handle("POST /threads/{threadID}/reply", api.NewReplyHandler(newReplyService()))
	mux.Handle("POST /threads/{threadID}/escalate", api.NewThreadEscalateHandler(newThreadEscalationService()))
	mux.Handle("GET /threads/{threadID}", api.NewThreadHandler(newThreadReadService()))
	mux.Handle("GET /threads/{threadID}/messages", api.NewThreadMessagesHandler(newThreadMessagesReadService()))
	mux.Handle("GET /contacts/{contactID}", api.NewContactHandler(newContactReadService()))
	mux.Handle("POST /contacts/{contactID}/memory", api.NewContactMemoryHandler(newContactMemoryService()))
	mux.Handle("POST /provider/callbacks/outbound", api.NewOutboundCallbackHandler(outbound.NewCallbackParser(), newOutboundCallbackFlow()))
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
				smtpedge.NewReceiptSink(newInboundService()),
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

func newInboundService() inbound.Service {
	return inbound.NewService(
		newInboundProcessor(),
		newInboundRecorder(),
	)
}

func newInboundProcessor() inbound.Processor {
	return inbound.NewProcessor(
		parser.New(notImplementedRawMessageReader{}),
		contacts.NewResolver(notImplementedContactLookup{}),
		threading.NewResolver(notImplementedThreadLookup{}),
	)
}

func newThreadReadService() app.ThreadReadService {
	return app.NewThreadReadService(notImplementedThreadRepository{})
}

func newThreadMessagesReadService() app.ThreadMessagesReadService {
	return app.NewThreadMessagesReadService(notImplementedThreadMessagesRepository{}, 20)
}

func newContactReadService() app.ContactReadService {
	return app.NewContactReadService(notImplementedContactRepository{})
}

func newThreadEscalationService() app.ThreadEscalationService {
	return app.NewThreadEscalationService(notImplementedThreadSaveRepository{}, time.Now)
}

func newContactMemoryService() app.ContactMemoryService {
	return app.NewContactMemoryService(notImplementedContactMemoryRepository{}, time.Now)
}

func newInboundRecorder() inbound.Recorder {
	return inbound.NewRecorder(
		notImplementedInboundContactRepository{},
		notImplementedInboundThreadRepository{},
		notImplementedInboundMessageRepository{},
	)
}

func newReplyService() outbound.Service {
	return outbound.NewService(
		outbound.NewAssembler(func() string { return "<generated@agentlayer.local>" }),
		outbound.NewRecorderWithThreads(notImplementedOutboundCreateMessageRepository{}, notImplementedOutboundThreadRepository{}),
		outbound.NewSender(notImplementedEmailProvider{}),
		outbound.NewStatusRecorder(notImplementedOutboundSaveMessageRepository{}),
		time.Now,
	)
}

func newOutboundCallbackFlow() outbound.CallbackFlow {
	return outbound.NewCallbackFlow(
		outbound.NewCallbackService(
			notImplementedProviderMessageLookup{},
			outbound.NewDeliveryRecorder(notImplementedOutboundSaveMessageRepository{}),
		),
		outbound.NewSuppressionService(notImplementedSuppressionRepository{}),
	)
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

type notImplementedThreadMessagesService struct{}

func (notImplementedThreadMessagesService) ListThreadMessages(context.Context, string) ([]domain.Message, error) {
	return nil, errors.New("thread messages service not implemented")
}

type notImplementedInboxLookup struct{}

func (notImplementedInboxLookup) FindByEmailAddress(context.Context, string) (domain.Inbox, bool, error) {
	return domain.Inbox{}, false, errors.New("smtp inbox lookup not implemented")
}

type notImplementedRawMessageStore struct{}

func (notImplementedRawMessageStore) Put(context.Context, string, []byte) error {
	return errors.New("smtp raw message store not implemented")
}

type notImplementedRawMessageReader struct{}

func (notImplementedRawMessageReader) Get(context.Context, string) ([]byte, error) {
	return nil, errors.New("raw message reader not implemented")
}

type notImplementedContactLookup struct{}

func (notImplementedContactLookup) FindByEmail(context.Context, string, string) (domain.Contact, bool, error) {
	return domain.Contact{}, false, errors.New("contact lookup not implemented")
}

type notImplementedThreadLookup struct{}

func (notImplementedThreadLookup) FindByMessageID(context.Context, string) (domain.Thread, bool, error) {
	return domain.Thread{}, false, errors.New("thread lookup not implemented")
}

type notImplementedThreadRepository struct{}

func (notImplementedThreadRepository) GetByID(context.Context, string) (domain.Thread, error) {
	return domain.Thread{}, errors.New("thread repository not implemented")
}

type notImplementedThreadMessagesRepository struct{}

func (notImplementedThreadMessagesRepository) ListByThreadID(context.Context, string, int) ([]domain.Message, error) {
	return nil, errors.New("thread messages repository not implemented")
}

type notImplementedContactRepository struct{}

func (notImplementedContactRepository) GetByID(context.Context, string) (domain.Contact, error) {
	return domain.Contact{}, errors.New("contact repository not implemented")
}

type notImplementedThreadSaveRepository struct{}

func (notImplementedThreadSaveRepository) Save(context.Context, domain.Thread) (domain.Thread, error) {
	return domain.Thread{}, errors.New("thread save repository not implemented")
}

type notImplementedContactMemoryRepository struct{}

func (notImplementedContactMemoryRepository) Create(context.Context, domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error) {
	return domain.ContactMemoryEntry{}, errors.New("contact memory repository not implemented")
}

type notImplementedInboundContactRepository struct{}

func (notImplementedInboundContactRepository) UpsertByEmail(context.Context, domain.Contact) (domain.Contact, error) {
	return domain.Contact{}, errors.New("inbound contact repository not implemented")
}

type notImplementedInboundThreadRepository struct{}

func (notImplementedInboundThreadRepository) Save(context.Context, domain.Thread) (domain.Thread, error) {
	return domain.Thread{}, errors.New("inbound thread repository not implemented")
}

type notImplementedInboundMessageRepository struct{}

func (notImplementedInboundMessageRepository) Create(context.Context, domain.Message) (domain.Message, error) {
	return domain.Message{}, errors.New("inbound message repository not implemented")
}

type notImplementedOutboundThreadRepository struct{}

func (notImplementedOutboundThreadRepository) Save(context.Context, domain.Thread) (domain.Thread, error) {
	return domain.Thread{}, errors.New("outbound thread repository not implemented")
}

type notImplementedOutboundCreateMessageRepository struct{}

func (notImplementedOutboundCreateMessageRepository) Create(context.Context, domain.Message) (domain.Message, error) {
	return domain.Message{}, errors.New("outbound create message repository not implemented")
}

type notImplementedOutboundSaveMessageRepository struct{}

func (notImplementedOutboundSaveMessageRepository) Save(context.Context, domain.Message) (domain.Message, error) {
	return domain.Message{}, errors.New("outbound save message repository not implemented")
}

type notImplementedEmailProvider struct{}

func (notImplementedEmailProvider) Send(context.Context, core.OutboundSendRequest) (core.SendResult, error) {
	return core.SendResult{}, errors.New("email provider not implemented")
}

func (notImplementedEmailProvider) GetDeliveryStatus(context.Context, string) (core.DeliveryStatus, error) {
	return core.DeliveryStatus{}, errors.New("email provider not implemented")
}

func (notImplementedEmailProvider) HealthCheck(context.Context) (core.ProviderHealth, error) {
	return core.ProviderHealth{}, errors.New("email provider not implemented")
}

type notImplementedProviderMessageLookup struct{}

func (notImplementedProviderMessageLookup) FindByProviderMessageID(context.Context, string) (domain.Message, bool, error) {
	return domain.Message{}, false, errors.New("provider message lookup not implemented")
}

type notImplementedSuppressionRepository struct{}

func (notImplementedSuppressionRepository) Save(context.Context, domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	return domain.SuppressedAddress{}, errors.New("suppression repository not implemented")
}

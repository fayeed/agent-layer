package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/app"
	"github.com/agentlayer/agentlayer/internal/contacts"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/outbound"
	"github.com/agentlayer/agentlayer/internal/parser"
	devprovider "github.com/agentlayer/agentlayer/internal/providers/dev"
	"github.com/agentlayer/agentlayer/internal/smtpedge"
	"github.com/agentlayer/agentlayer/internal/store/blobfs"
	memorystore "github.com/agentlayer/agentlayer/internal/store/memory"
	"github.com/agentlayer/agentlayer/internal/threading"
	"github.com/agentlayer/agentlayer/internal/webhooks"
	smtp "github.com/emersion/go-smtp"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var runtimeStore = newRuntimeStore()

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
	mux.Handle("GET /bootstrap", api.NewBootstrapReadHandler(newBootstrapReadHandlerService()))
	mux.Handle("POST /bootstrap", api.NewBootstrapHandler(newBootstrapHandlerService()))
	mux.Handle("GET /inbound/receipts/list", api.NewInboundReceiptsHandler(newInboundReceiptsHandlerService()))
	mux.Handle("GET /inbound/receipts", api.NewInboundReceiptHandler(newInboundReceiptHandlerService()))
	mux.Handle("POST /inbound/reprocess", api.NewInboundReprocessHandler(newInboundReprocessHandlerService()))
	mux.Handle("GET /webhooks/deliveries", api.NewWebhookDeliveriesHandler(newWebhookDeliveriesHandlerService()))
	mux.Handle("GET /webhooks/deliveries/{deliveryID}", api.NewWebhookDeliveryHandler(newWebhookDeliveryHandlerService()))
	mux.Handle("POST /webhooks/deliveries/{deliveryID}/replay", api.NewWebhookReplayHandler(newWebhookReplayHandlerService()))
	mux.Handle("POST /threads/{threadID}/reply", api.NewReplyHandler(newReplyHandlerService()))
	mux.Handle("POST /threads/{threadID}/escalate", api.NewThreadEscalateHandler(newThreadEscalationHandlerService()))
	mux.Handle("GET /threads/{threadID}", api.NewThreadHandler(newThreadReadService()))
	mux.Handle("GET /threads/{threadID}/messages", api.NewThreadMessagesHandler(newThreadMessagesReadService()))
	mux.Handle("GET /contacts/{contactID}", api.NewContactHandler(newContactReadService()))
	mux.Handle("POST /contacts/{contactID}/memory", api.NewContactMemoryHandler(newContactMemoryHandlerService()))
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

func webhookURL() string {
	return os.Getenv("AGENTLAYER_WEBHOOK_URL")
}

func webhookSecret() string {
	return os.Getenv("AGENTLAYER_WEBHOOK_SECRET")
}

func databaseURL() string {
	return os.Getenv("AGENTLAYER_DATABASE_URL")
}

func rawDataDir() string {
	if value := os.Getenv("AGENTLAYER_RAW_DATA_DIR"); value != "" {
		return value
	}
	return ".agentlayer-data/raw"
}

func newRuntimeStore() appStore {
	if databaseURL() != "" {
		db, err := sql.Open("pgx", databaseURL())
		if err != nil {
			panic(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			panic(err)
		}
		store := newPostgresRuntimeStore(db, blobfs.NewStore(rawDataDir()))
		seedLocalRuntime(store)
		return store
	}

	store := memorystore.NewStore()
	seedLocalRuntime(store)
	return store
}

func seedLocalRuntime(store interface {
	SaveOrganization(ctx context.Context, organization domain.Organization) (domain.Organization, error)
	SaveAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error)
	SaveInbox(ctx context.Context, inbox domain.Inbox) (domain.Inbox, error)
}) {
	_, _ = store.SaveOrganization(context.Background(), domain.Organization{
		ID:   "org-local",
		Name: "AgentLayer Local",
	})
	_, _ = store.SaveAgent(context.Background(), domain.Agent{
		ID:             "agent-local",
		OrganizationID: "org-local",
		Name:           "Local Agent",
		Status:         domain.AgentStatusActive,
	})
	_, _ = store.SaveInbox(context.Background(), domain.Inbox{
		ID:             "inbox-local",
		OrganizationID: "org-local",
		AgentID:        "agent-local",
		EmailAddress:   "agent@localhost",
		Domain:         "localhost",
		DisplayName:    "AgentLayer Local",
	})
}

func newSMTPServer() *smtp.Server {
	return smtpedge.NewServer(
		smtpedge.NewBackend(func() smtpedge.CoreSession {
			now := time.Now().UTC()
			sessionID := smtpedge.NewSessionID(now)
			session := smtpedge.NewSession(
				runtimeStore,
				runtimeStore,
				smtpedge.NewReceiptSinkWithRecorder(newInboundService(), runtimeStore),
				func() time.Time { return now },
				smtpedge.NewRawMessageObjectKey,
				sessionID,
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

func newInboundService() smtpedge.StoredMessageHandler {
	base := inbound.NewServiceWithDuplicateLookup(
		newInboundProcessor(),
		newInboundRecorder(),
		runtimeStore,
	)
	organization := currentOrganization()
	agent := currentAgent()
	inbox := currentInbox()

	return app.NewInboundRuntimeService(
		base,
		runtimeStore,
		contactMemoryListerAdapter{store: runtimeStore},
		newMessageReceivedDeliveryService(),
		time.Now,
		app.InboundRuntimeConfig{
			Organization:  organization,
			Agent:         agent,
			Inbox:         inbox,
			WebhookURL:    agent.WebhookURL,
			WebhookSecret: agent.WebhookSecret,
			HistoryLimit:  20,
			MemoryLimit:   10,
		},
	)
}

func currentOrganization() domain.Organization {
	organization, err := runtimeStore.GetOrganizationByID(context.Background(), "org-local")
	if err != nil {
		return domain.Organization{ID: "org-local", Name: "AgentLayer Local"}
	}
	return organization
}

func currentAgent() domain.Agent {
	agent, err := runtimeStore.GetAgentByID(context.Background(), "agent-local")
	if err != nil {
		return domain.Agent{ID: "agent-local", OrganizationID: "org-local", Name: "Local Agent", Status: domain.AgentStatusActive, WebhookURL: webhookURL(), WebhookSecret: webhookSecret()}
	}
	if agent.WebhookURL == "" {
		agent.WebhookURL = webhookURL()
	}
	if agent.WebhookSecret == "" {
		agent.WebhookSecret = webhookSecret()
	}
	return agent
}

func currentInbox() domain.Inbox {
	inbox, err := runtimeStore.GetInboxByID(context.Background(), "inbox-local")
	if err != nil {
		return domain.Inbox{ID: "inbox-local", OrganizationID: "org-local", AgentID: "agent-local", EmailAddress: "agent@localhost", Domain: "localhost", DisplayName: "AgentLayer Local"}
	}
	return inbox
}

func newInboundProcessor() inbound.Processor {
	return inbound.NewProcessor(
		parser.New(runtimeStore),
		contacts.NewResolver(runtimeStore),
		threading.NewResolver(runtimeStore),
	)
}

func newThreadReadService() app.ThreadReadService {
	return app.NewThreadReadService(runtimeStore)
}

func newInboundReprocessService() app.InboundReprocessService {
	return app.NewInboundReprocessService(runtimeStore, newInboundService())
}

func newInboundReceiptReadService() app.InboundReceiptReadService {
	return app.NewInboundReceiptReadService(runtimeStore)
}

func newInboundReceiptListService() app.InboundReceiptListService {
	return app.NewInboundReceiptListService(runtimeStore, 20)
}

func newInboundReceiptsHandlerService() api.InboundReceiptsService {
	return inboundReceiptListServiceAdapter{service: newInboundReceiptListService()}
}

func newInboundReceiptHandlerService() api.InboundReceiptService {
	return inboundReceiptReadServiceAdapter{service: newInboundReceiptReadService()}
}

func newInboundReprocessHandlerService() api.InboundReprocessService {
	return inboundReprocessServiceAdapter{service: newInboundReprocessService()}
}

func newThreadMessagesReadService() app.ThreadMessagesReadService {
	return app.NewThreadMessagesReadService(runtimeStore, 20)
}

func newContactReadService() app.ContactReadService {
	return app.NewContactReadService(contactGetterAdapter{store: runtimeStore})
}

func newThreadEscalationService() app.ThreadEscalationService {
	return app.NewThreadEscalationService(runtimeStore, time.Now)
}

func newThreadEscalationHandlerService() api.ThreadEscalationService {
	return threadEscalationServiceAdapter{
		service: newThreadEscalationService(),
		threads: runtimeStore,
	}
}

func newContactMemoryService() app.ContactMemoryService {
	return app.NewContactMemoryService(contactMemoryWriterAdapter{store: runtimeStore}, time.Now)
}

func newContactMemoryHandlerService() api.ContactMemoryService {
	return contactMemoryServiceAdapter{
		service:  newContactMemoryService(),
		contacts: contactGetterAdapter{store: runtimeStore},
		threads:  runtimeStore,
	}
}

func newBootstrapService() app.BootstrapService {
	return app.NewBootstrapService(runtimeStore, runtimeStore, runtimeStore, time.Now)
}

func newBootstrapHandlerService() api.BootstrapService {
	return bootstrapServiceAdapter{service: newBootstrapService()}
}

func newBootstrapReadService() app.BootstrapReadService {
	return app.NewBootstrapReadService(runtimeStore, runtimeStore, runtimeStore)
}

func newBootstrapReadHandlerService() api.BootstrapReadService {
	return bootstrapReadServiceAdapter{service: newBootstrapReadService()}
}

func newWebhookDeliveryReadService() app.WebhookDeliveryReadService {
	return app.NewWebhookDeliveryReadService(webhookDeliveryGetterAdapter{store: runtimeStore})
}

func newWebhookDeliveryListService() app.WebhookDeliveryListService {
	return app.NewWebhookDeliveryListService(webhookDeliveryGetterAdapter{store: runtimeStore}, 20)
}

func newWebhookDeliveriesHandlerService() api.WebhookDeliveriesService {
	return webhookDeliveryListServiceAdapter{service: newWebhookDeliveryListService()}
}

func newWebhookDeliveryHandlerService() api.WebhookDeliveryService {
	return webhookDeliveryReadServiceAdapter{service: newWebhookDeliveryReadService()}
}

func newWebhookReplayHandlerService() api.WebhookReplayService {
	return webhookReplayServiceAdapter{service: newWebhookReplayService()}
}

func newInboundRecorder() inbound.Recorder {
	return inbound.NewRecorder(
		runtimeStore,
		runtimeStore,
		runtimeStore,
	)
}

func newReplyService() outbound.Service {
	return outbound.NewService(
		outbound.NewAssembler(func() string { return "<generated@agentlayer.local>" }),
		outbound.NewRecorderWithThreads(runtimeStore, runtimeStore),
		outbound.NewSender(devprovider.NewEmailProvider(time.Now)),
		outbound.NewStatusRecorder(messageStatusRepositoryAdapter{store: runtimeStore}),
		time.Now,
	)
}

func newReplyHandlerService() api.ReplyService {
	return replyServiceAdapter{
		service:  newReplyService(),
		orgs:     runtimeStore,
		agents:   runtimeStore,
		inboxes:  runtimeStore,
		threads:  runtimeStore,
		contacts: contactGetterAdapter{store: runtimeStore},
		messages: messageGetterAdapter{store: runtimeStore},
	}
}

func newOutboundCallbackFlow() outbound.CallbackFlow {
	return outbound.NewCallbackFlow(
		outbound.NewCallbackService(
			runtimeStore,
			outbound.NewDeliveryRecorder(messageStatusRepositoryAdapter{store: runtimeStore}),
		),
		outbound.NewSuppressionService(suppressionRepositoryAdapter{store: runtimeStore}),
	)
}

func newMessageReceivedDeliveryService() webhooks.DeliveryService {
	dispatcher := webhooks.NewDispatcher(&http.Client{Timeout: 5 * time.Second}, time.Now)
	base := webhooks.NewService(
		webhooks.NewMessageReceivedBuilder(),
		webhooks.NewSigner(time.Now),
		dispatcher,
	)
	return webhooks.NewDeliveryService(
		base,
		webhooks.NewRecorder(webhookDeliveryRepositoryAdapter{store: runtimeStore}),
	)
}

func newWebhookReplayService() webhooks.ReplayService {
	dispatcher := webhooks.NewDispatcher(&http.Client{Timeout: 5 * time.Second}, time.Now)
	return webhooks.NewReplayService(
		webhookDeliveryGetterAdapter{store: runtimeStore},
		dispatcher,
		webhooks.NewRecorder(webhookDeliveryRepositoryAdapter{store: runtimeStore}),
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

type notImplementedProviderMessageLookup struct{}

func (notImplementedProviderMessageLookup) FindByProviderMessageID(context.Context, string) (domain.Message, bool, error) {
	return domain.Message{}, false, errors.New("provider message lookup not implemented")
}

type notImplementedSuppressionRepository struct{}

func (notImplementedSuppressionRepository) Save(context.Context, domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	return domain.SuppressedAddress{}, errors.New("suppression repository not implemented")
}

type contactGetterAdapter struct{ store appStore }

type bootstrapServiceAdapter struct{ service app.BootstrapService }

type bootstrapReadServiceAdapter struct{ service app.BootstrapReadService }

type replyServiceAdapter struct {
	service  outbound.Service
	orgs     organizationGetter
	agents   agentGetter
	inboxes  inboxGetter
	threads  threadGetter
	contacts contactGetter
	messages messageGetter
}

type threadEscalationServiceAdapter struct {
	service app.ThreadEscalationService
	threads threadGetter
}

type contactMemoryServiceAdapter struct {
	service  app.ContactMemoryService
	contacts contactGetter
	threads  threadGetter
}

type webhookDeliveryListServiceAdapter struct {
	service app.WebhookDeliveryListService
}

type webhookDeliveryReadServiceAdapter struct {
	service app.WebhookDeliveryReadService
}

type webhookReplayServiceAdapter struct{ service webhooks.ReplayService }

type inboundReprocessServiceAdapter struct {
	service app.InboundReprocessService
}

type inboundReceiptReadServiceAdapter struct {
	service app.InboundReceiptReadService
}

type inboundReceiptListServiceAdapter struct {
	service app.InboundReceiptListService
}

func (a bootstrapServiceAdapter) BootstrapLocal(ctx context.Context, input api.BootstrapInput) (api.BootstrapResult, error) {
	result, err := a.service.BootstrapLocal(ctx, app.BootstrapInput{
		OrganizationName: input.OrganizationName,
		AgentName:        input.AgentName,
		AgentStatus:      domain.AgentStatus(input.AgentStatus),
		WebhookURL:       input.WebhookURL,
		WebhookSecret:    input.WebhookSecret,
		InboxAddress:     input.InboxAddress,
		InboxDomain:      input.InboxDomain,
		InboxDisplayName: input.InboxDisplayName,
	})
	if err != nil {
		return api.BootstrapResult{}, err
	}

	return api.BootstrapResult{
		OrganizationID: result.Organization.ID,
		AgentID:        result.Agent.ID,
		InboxID:        result.Inbox.ID,
		WebhookURL:     result.Agent.WebhookURL,
		InboxAddress:   result.Inbox.EmailAddress,
	}, nil
}

func (a inboundReprocessServiceAdapter) ReprocessByObjectKey(ctx context.Context, objectKey string) (api.InboundReprocessResult, error) {
	result, err := a.service.ReprocessByObjectKey(ctx, objectKey)
	if err != nil {
		return api.InboundReprocessResult{}, err
	}

	return api.InboundReprocessResult{
		MessageID: result.Message.ID,
		ThreadID:  result.Thread.ID,
		Duplicate: result.Duplicate,
	}, nil
}

func (a inboundReceiptReadServiceAdapter) GetInboundReceipt(ctx context.Context, objectKey string) (inbound.DurableReceiptRequest, error) {
	return a.service.GetInboundReceipt(ctx, objectKey)
}

func (a inboundReceiptListServiceAdapter) ListInboundReceipts(ctx context.Context, limit int) ([]inbound.DurableReceiptRequest, error) {
	return a.service.ListInboundReceipts(ctx, limit)
}

func (a bootstrapReadServiceAdapter) GetBootstrap(ctx context.Context) (api.BootstrapResult, error) {
	result, err := a.service.GetBootstrap(ctx)
	if err != nil {
		return api.BootstrapResult{}, err
	}

	return api.BootstrapResult{
		OrganizationID: result.Organization.ID,
		AgentID:        result.Agent.ID,
		InboxID:        result.Inbox.ID,
		WebhookURL:     result.Agent.WebhookURL,
		InboxAddress:   result.Inbox.EmailAddress,
	}, nil
}

func (a replyServiceAdapter) SendReply(ctx context.Context, input outbound.SendReplyInput) (outbound.SendReplyResult, error) {
	organizationID := input.Organization.ID
	if organizationID == "" {
		organizationID = "org-local"
	}
	agentID := input.Agent.ID
	if agentID == "" {
		agentID = "agent-local"
	}
	inboxID := input.Inbox.ID
	if inboxID == "" {
		inboxID = "inbox-local"
	}

	organization, err := a.orgs.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	agent, err := a.agents.GetAgentByID(ctx, agentID)
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	inbox, err := a.inboxes.GetInboxByID(ctx, inboxID)
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	thread, err := a.threads.GetByID(ctx, input.Thread.ID)
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	contactID := input.Contact.ID
	if contactID == "" {
		contactID = thread.ContactID
	}
	contact, err := a.contacts.GetByID(ctx, contactID)
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	replyToMessage, err := a.messages.GetMessageByID(ctx, input.ReplyToMessage.ID)
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	return a.service.SendReply(ctx, outbound.SendReplyInput{
		Organization:   organization,
		Agent:          agent,
		Inbox:          inbox,
		Thread:         thread,
		ReplyToMessage: replyToMessage,
		Contact:        contact,
		BodyText:       input.BodyText,
		ObjectKey:      input.ObjectKey,
	})
}

func (a threadEscalationServiceAdapter) EscalateThread(ctx context.Context, threadID, reason string) (domain.Thread, error) {
	if _, err := a.threads.GetByID(ctx, threadID); err != nil {
		return domain.Thread{}, err
	}
	return a.service.EscalateThread(ctx, threadID, reason)
}

func (a contactMemoryServiceAdapter) CreateContactMemory(ctx context.Context, contactID string, input api.CreateContactMemoryInput) (domain.ContactMemoryEntry, error) {
	if _, err := a.contacts.GetByID(ctx, contactID); err != nil {
		return domain.ContactMemoryEntry{}, err
	}
	if input.ThreadID != "" {
		if _, err := a.threads.GetByID(ctx, input.ThreadID); err != nil {
			return domain.ContactMemoryEntry{}, err
		}
	}
	return a.service.CreateContactMemory(ctx, contactID, input)
}

func (a webhookDeliveryListServiceAdapter) ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	return a.service.ListWebhookDeliveries(ctx, limit)
}

func (a webhookDeliveryReadServiceAdapter) GetWebhookDelivery(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	return a.service.GetWebhookDelivery(ctx, deliveryID)
}

func (a webhookReplayServiceAdapter) ReplayDelivery(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	result, err := a.service.ReplayDelivery(ctx, deliveryID)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}
	return result.Delivery, nil
}

func (a contactGetterAdapter) GetByID(ctx context.Context, contactID string) (domain.Contact, error) {
	return a.store.GetContactByID(ctx, contactID)
}

type contactMemoryWriterAdapter struct{ store appStore }

func (a contactMemoryWriterAdapter) Create(ctx context.Context, entry domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error) {
	return a.store.CreateMemory(ctx, entry)
}

type contactMemoryListerAdapter struct{ store appStore }

func (a contactMemoryListerAdapter) ListMemoryByContactID(ctx context.Context, contactID string, limit int) ([]domain.ContactMemoryEntry, error) {
	return a.store.ListMemoryByContactID(ctx, contactID, limit)
}

type messageStatusRepositoryAdapter struct{ store appStore }

func (a messageStatusRepositoryAdapter) Save(ctx context.Context, message domain.Message) (domain.Message, error) {
	return a.store.SaveMessage(ctx, message)
}

type suppressionRepositoryAdapter struct{ store appStore }

func (a suppressionRepositoryAdapter) Save(ctx context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error) {
	return a.store.SaveSuppression(ctx, record)
}

type webhookDeliveryRepositoryAdapter struct{ store appStore }

func (a webhookDeliveryRepositoryAdapter) Save(ctx context.Context, delivery domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	return a.store.SaveWebhookDelivery(ctx, delivery)
}

type webhookDeliveryGetterAdapter struct{ store appStore }

func (a webhookDeliveryGetterAdapter) GetWebhookDeliveryByID(ctx context.Context, deliveryID string) (domain.WebhookDelivery, error) {
	return a.store.GetWebhookDeliveryByID(ctx, deliveryID)
}

func (a webhookDeliveryGetterAdapter) ListWebhookDeliveries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	return a.store.ListWebhookDeliveries(ctx, limit)
}

type organizationGetter interface {
	GetOrganizationByID(ctx context.Context, organizationID string) (domain.Organization, error)
}

type agentGetter interface {
	GetAgentByID(ctx context.Context, agentID string) (domain.Agent, error)
}

type inboxGetter interface {
	GetInboxByID(ctx context.Context, inboxID string) (domain.Inbox, error)
}

type threadGetter interface {
	GetByID(ctx context.Context, threadID string) (domain.Thread, error)
}

type contactGetter interface {
	GetByID(ctx context.Context, contactID string) (domain.Contact, error)
}

type messageGetter interface {
	GetMessageByID(ctx context.Context, messageID string) (domain.Message, error)
}

type messageGetterAdapter struct{ store appStore }

func (a messageGetterAdapter) GetMessageByID(ctx context.Context, messageID string) (domain.Message, error) {
	return a.store.GetMessageByID(ctx, messageID)
}

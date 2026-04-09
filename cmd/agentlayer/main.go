package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/agentlayer/agentlayer/db/migrations"
	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/app"
	"github.com/agentlayer/agentlayer/internal/contacts"
	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
	"github.com/agentlayer/agentlayer/internal/outbound"
	"github.com/agentlayer/agentlayer/internal/parser"
	"github.com/agentlayer/agentlayer/internal/platform/idempotency"
	devprovider "github.com/agentlayer/agentlayer/internal/providers/dev"
	sesprovider "github.com/agentlayer/agentlayer/internal/providers/ses"
	"github.com/agentlayer/agentlayer/internal/smtpedge"
	"github.com/agentlayer/agentlayer/internal/store/blobfs"
	"github.com/agentlayer/agentlayer/internal/store/blobs3"
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

	if err := runServers(httpServer, smtpServer, newWebhookRetryWorker()); err != nil {
		log.Fatal(err)
	}
}

type serveServer interface {
	ListenAndServe() error
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.Handle("GET /readyz", newReadinessHandler())
	mux.Handle("GET /bootstrap", api.NewBootstrapReadHandler(newBootstrapReadHandlerService()))
	mux.Handle("POST /bootstrap", api.NewBootstrapHandler(newBootstrapHandlerService()))
	mux.Handle("GET /inbound/receipts/list", api.NewInboundReceiptsHandler(newInboundReceiptsHandlerService()))
	mux.Handle("GET /inbound/receipts", api.NewInboundReceiptHandler(newInboundReceiptHandlerService()))
	mux.Handle("POST /inbound/reprocess", api.NewInboundReprocessHandler(newInboundReprocessHandlerService()))
	mux.Handle("GET /webhooks/deliveries", api.NewWebhookDeliveriesHandler(newWebhookDeliveriesHandlerService()))
	mux.Handle("POST /webhooks/deliveries/retry", api.NewWebhookRetryHandler(newWebhookRetryHandlerService()))
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

func rawStoreType() string {
	if value := os.Getenv("AGENTLAYER_RAW_STORE"); value != "" {
		return value
	}
	return "fs"
}

func rawS3Bucket() string {
	return os.Getenv("AGENTLAYER_S3_BUCKET")
}

func rawS3Endpoint() string {
	return os.Getenv("AGENTLAYER_S3_ENDPOINT")
}

func rawS3PathStyle() bool {
	value := os.Getenv("AGENTLAYER_S3_PATH_STYLE")
	if value == "" {
		return true
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return true
	}
	return enabled
}

func rawS3AccessKeyID() string {
	return os.Getenv("AGENTLAYER_S3_ACCESS_KEY_ID")
}

func rawS3SecretAccessKey() string {
	return os.Getenv("AGENTLAYER_S3_SECRET_ACCESS_KEY")
}

func emailProviderType() string {
	if value := os.Getenv("AGENTLAYER_EMAIL_PROVIDER"); value != "" {
		return value
	}
	return "dev"
}

func awsRegion() string {
	return os.Getenv("AWS_REGION")
}

func validateEmailProviderConfig() error {
	switch emailProviderType() {
	case "dev":
		return nil
	case "ses":
		if strings.TrimSpace(awsRegion()) == "" {
			return errors.New("AWS_REGION is required when AGENTLAYER_EMAIL_PROVIDER=ses")
		}
		return nil
	default:
		return fmt.Errorf("unsupported email provider %q", emailProviderType())
	}
}

func validateRawStoreConfig() error {
	switch rawStoreType() {
	case "fs":
		return nil
	case "s3":
		if strings.TrimSpace(rawS3Bucket()) == "" {
			return errors.New("AGENTLAYER_S3_BUCKET is required when AGENTLAYER_RAW_STORE=s3")
		}
		if strings.TrimSpace(awsRegion()) == "" {
			return errors.New("AWS_REGION is required when AGENTLAYER_RAW_STORE=s3")
		}
		return nil
	default:
		return fmt.Errorf("unsupported raw store %q", rawStoreType())
	}
}

func autoMigrateEnabled() bool {
	value := os.Getenv("AGENTLAYER_AUTO_MIGRATE")
	return value == "1" || value == "true" || value == "TRUE"
}

func newRuntimeStore() appStore {
	if databaseURL() != "" {
		if err := validateRawStoreConfig(); err != nil {
			panic(err)
		}
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
		if autoMigrateEnabled() {
			if err := applyV0Schema(ctx, db); err != nil {
				_ = db.Close()
				panic(err)
			}
		}
		rawStore, rawStoreDescription, err := newRawStore(context.Background())
		if err != nil {
			_ = db.Close()
			panic(err)
		}
		store := newPostgresRuntimeStore(db, rawStore)
		log.Printf("agentlayer runtime store: postgres raw=%s auto_migrate=%t", rawStoreDescription, autoMigrateEnabled())
		seedLocalRuntime(store)
		return store
	}

	store := memorystore.NewStore()
	log.Printf("agentlayer runtime store: memory")
	seedLocalRuntime(store)
	return store
}

func newRawStore(ctx context.Context) (interface {
	Put(ctx context.Context, objectKey string, data []byte) error
	Get(ctx context.Context, objectKey string) ([]byte, error)
}, string, error) {
	switch rawStoreType() {
	case "fs":
		return blobfs.NewStore(rawDataDir()), "fs:" + rawDataDir(), nil
	case "s3":
		store, err := blobs3.NewStore(ctx, blobs3.Config{
			Region:          awsRegion(),
			Bucket:          rawS3Bucket(),
			Endpoint:        rawS3Endpoint(),
			PathStyle:       rawS3PathStyle(),
			AccessKeyID:     rawS3AccessKeyID(),
			SecretAccessKey: rawS3SecretAccessKey(),
		})
		if err != nil {
			return nil, "", err
		}
		description := "s3:" + rawS3Bucket()
		if endpoint := rawS3Endpoint(); endpoint != "" {
			description += "@" + endpoint
		}
		return store, description, nil
	default:
		return nil, "", fmt.Errorf("unsupported raw store %q", rawStoreType())
	}
}

func applyV0Schema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, migrations.V0CoreSQL())
	return err
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

func newWebhookRetrySweepService() webhooks.RetrySweepService {
	return webhooks.NewRetrySweepService(
		webhookDeliveryGetterAdapter{store: runtimeStore},
		newWebhookReplayService(),
		time.Now,
	)
}

func newWebhookRetryHandlerService() api.WebhookRetryService {
	return webhookRetryServiceAdapter{service: newWebhookRetrySweepService()}
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
		outbound.NewRecorderWithStore(runtimeStore, runtimeStore, runtimeStore),
		outbound.NewSender(newEmailProvider()),
		outbound.NewStatusRecorder(messageStatusRepositoryAdapter{store: runtimeStore}),
		suppressionCheckerAdapter{store: runtimeStore},
		time.Now,
	)
}

func newEmailProvider() outbound.EmailProvider {
	if err := validateEmailProviderConfig(); err != nil {
		panic(err)
	}

	switch emailProviderType() {
	case "ses":
		provider, err := sesprovider.NewEmailProvider(context.Background(), awsRegion(), time.Now)
		if err != nil {
			panic(err)
		}
		return provider
	default:
		return devprovider.NewEmailProvider(time.Now)
	}
}

func newReplyHandlerService() api.ReplyService {
	return replyServiceAdapter{
		service:  newReplyService(),
		store:    runtimeStore,
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

func runServers(httpServer serveServer, smtpServer serveServer, workers ...backgroundWorker) error {
	errCh := make(chan error, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, worker := range workers {
		if worker == nil {
			continue
		}
		go worker.Run(ctx)
	}

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
	store    appStore
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

type webhookRetryServiceAdapter struct {
	service webhooks.RetrySweepService
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
	if input.IdempotencyKey != "" {
		replyKey := idempotency.ReplySubmissionKey(input.Thread.ID, input.IdempotencyKey)
		existing, ok, err := a.store.FindReplyBySubmissionKey(ctx, replyKey)
		if err != nil {
			return outbound.SendReplyResult{}, err
		}
		if ok {
			thread, err := a.threads.GetByID(ctx, input.Thread.ID)
			if err != nil {
				return outbound.SendReplyResult{}, err
			}
			return outbound.SendReplyResult{
				Thread:  thread,
				Message: existing,
				SendResult: core.SendResult{
					ProviderMessageID: existing.ProviderMessageID,
					AcceptedAt:        existing.SentAt,
				},
			}, nil
		}
	}

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

	result, err := a.service.SendReply(ctx, outbound.SendReplyInput{
		Organization:   organization,
		Agent:          agent,
		Inbox:          inbox,
		Thread:         thread,
		ReplyToMessage: replyToMessage,
		Contact:        contact,
		BodyText:       input.BodyText,
		ObjectKey:      input.ObjectKey,
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return outbound.SendReplyResult{}, err
	}

	if input.IdempotencyKey != "" {
		replyKey := idempotency.ReplySubmissionKey(input.Thread.ID, input.IdempotencyKey)
		if err := a.store.SaveReplySubmission(ctx, replyKey, result.Message.ID); err != nil {
			return outbound.SendReplyResult{}, err
		}
	}

	return result, nil
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

func (a webhookRetryServiceAdapter) RetryDueDeliveries(ctx context.Context, limit int) (api.WebhookRetryResult, error) {
	result, err := a.service.RetryDueDeliveries(ctx, limit)
	if err != nil {
		return api.WebhookRetryResult{}, err
	}
	return api.WebhookRetryResult{
		Attempted: result.Attempted,
		Succeeded: result.Succeeded,
		Failed:    result.Failed,
		Skipped:   result.Skipped,
	}, nil
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

type suppressionCheckerAdapter struct{ store appStore }

func (a suppressionCheckerAdapter) IsSuppressed(ctx context.Context, organizationID, emailAddress string) (bool, error) {
	return a.store.IsSuppressed(ctx, organizationID, emailAddress)
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

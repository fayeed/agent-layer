package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/outbound"
)

func main() {
	addr := serverAddress()
	log.Printf("agentlayer listening on %s", addr)

	if err := http.ListenAndServe(addr, newServer()); err != nil {
		log.Fatal(err)
	}
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

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/domain"
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
	mux.HandleFunc("POST /threads/{threadID}/reply", handleNotImplemented)
	mux.HandleFunc("POST /threads/{threadID}/escalate", handleNotImplemented)
	mux.Handle("GET /threads/{threadID}", api.NewThreadHandler(notImplementedThreadService{}))
	mux.HandleFunc("GET /threads/{threadID}/messages", handleNotImplemented)
	mux.HandleFunc("GET /contacts/{contactID}", handleNotImplemented)
	mux.HandleFunc("POST /contacts/{contactID}/memory", handleNotImplemented)
	mux.HandleFunc("POST /provider/callbacks/outbound", handleNotImplemented)
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

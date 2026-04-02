package main

import (
	"log"
	"net/http"
	"os"
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
	mux.HandleFunc("GET /threads/{threadID}", handleNotImplemented)
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

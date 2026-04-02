package smtpedge

import (
	"time"

	smtp "github.com/emersion/go-smtp"
)

type Config struct {
	Addr            string
	Domain          string
	MaxRecipients   int
	MaxMessageBytes int64
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func NewServer(backend Backend, cfg Config) *smtp.Server {
	server := smtp.NewServer(backend)
	server.Addr = defaultString(cfg.Addr, "localhost:2525")
	server.Domain = defaultString(cfg.Domain, "localhost")
	server.MaxRecipients = defaultInt(cfg.MaxRecipients, 1)
	server.MaxMessageBytes = defaultInt64(cfg.MaxMessageBytes, 25*1024*1024)
	server.ReadTimeout = defaultDuration(cfg.ReadTimeout, 10*time.Second)
	server.WriteTimeout = defaultDuration(cfg.WriteTimeout, 10*time.Second)
	return server
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func defaultInt(value, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}

func defaultInt64(value, fallback int64) int64 {
	if value != 0 {
		return value
	}
	return fallback
}

func defaultDuration(value, fallback time.Duration) time.Duration {
	if value != 0 {
		return value
	}
	return fallback
}

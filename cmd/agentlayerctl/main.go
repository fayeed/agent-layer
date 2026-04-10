package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "bootstrap":
		if err := runBootstrap(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "ready":
		if err := runReady(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "show":
		if err := runShow(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "retry-webhooks":
		if err := runRetryWebhooks(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "send-sample":
		if err := runSendSample(os.Args[2:]); err != nil {
			fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: agentlayerctl <bootstrap|ready|show|retry-webhooks|send-sample> [flags]\n")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func runBootstrap(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	baseURL := fs.String("base-url", envOrDefault("AGENTLAYER_BASE_URL", "http://localhost:8080"), "")
	webhookURL := fs.String("webhook-url", envOrDefault("AGENTLAYER_WEBHOOK_URL", "http://localhost:3000/webhook"), "")
	webhookSecret := fs.String("webhook-secret", envOrDefault("AGENTLAYER_WEBHOOK_SECRET", "dev-secret"), "")
	orgName := fs.String("org-name", envOrDefault("AGENTLAYER_ORG_NAME", "Acme Support"), "")
	agentName := fs.String("agent-name", envOrDefault("AGENTLAYER_AGENT_NAME", "Acme Agent"), "")
	agentStatus := fs.String("agent-status", envOrDefault("AGENTLAYER_AGENT_STATUS", "active"), "")
	inboxAddress := fs.String("inbox-address", envOrDefault("AGENTLAYER_INBOX_ADDRESS", "agent@localhost"), "")
	inboxDomain := fs.String("inbox-domain", envOrDefault("AGENTLAYER_INBOX_DOMAIN", "localhost"), "")
	inboxDisplayName := fs.String("inbox-display-name", envOrDefault("AGENTLAYER_INBOX_DISPLAY_NAME", "Acme Inbox"), "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	payload := map[string]string{
		"organization_name":  *orgName,
		"agent_name":         *agentName,
		"agent_status":       *agentStatus,
		"webhook_url":        *webhookURL,
		"webhook_secret":     *webhookSecret,
		"inbox_address":      *inboxAddress,
		"inbox_domain":       *inboxDomain,
		"inbox_display_name": *inboxDisplayName,
	}
	return printJSONRequest(http.MethodPost, strings.TrimRight(*baseURL, "/")+"/bootstrap", payload)
}

func runShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	baseURL := fs.String("base-url", envOrDefault("AGENTLAYER_BASE_URL", "http://localhost:8080"), "")
	webhookLimit := fs.Int("webhook-limit", envIntOrDefault("AGENTLAYER_WEBHOOK_LIMIT", 5), "")
	receiptLimit := fs.Int("receipt-limit", envIntOrDefault("AGENTLAYER_RECEIPT_LIMIT", 5), "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	for _, endpoint := range []string{
		"/readyz",
		"/bootstrap",
		fmt.Sprintf("/webhooks/deliveries?limit=%d", *webhookLimit),
		fmt.Sprintf("/inbound/receipts/list?limit=%d", *receiptLimit),
	} {
		resp, err := http.Get(strings.TrimRight(*baseURL, "/") + endpoint)
		if err != nil {
			return err
		}
		if err := printResponse(endpoint, resp); err != nil {
			return err
		}
	}
	return nil
}

func runReady(args []string) error {
	fs := flag.NewFlagSet("ready", flag.ContinueOnError)
	baseURL := fs.String("base-url", envOrDefault("AGENTLAYER_BASE_URL", "http://localhost:8080"), "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	resp, err := http.Get(strings.TrimRight(*baseURL, "/") + "/readyz")
	if err != nil {
		return err
	}
	return printResponse("/readyz", resp)
}

func runRetryWebhooks(args []string) error {
	fs := flag.NewFlagSet("retry-webhooks", flag.ContinueOnError)
	baseURL := fs.String("base-url", envOrDefault("AGENTLAYER_BASE_URL", "http://localhost:8080"), "")
	limit := fs.Int("limit", envIntOrDefault("AGENTLAYER_WEBHOOK_RETRY_LIMIT", 20), "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/webhooks/deliveries/retry?limit=%d", strings.TrimRight(*baseURL, "/"), *limit), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return printResponse("/webhooks/deliveries/retry", resp)
}

func runSendSample(args []string) error {
	fs := flag.NewFlagSet("send-sample", flag.ContinueOnError)
	smtpAddr := fs.String("smtp-addr", envOrDefault("AGENTLAYER_SMTP_ADDR", "localhost:2525"), "")
	from := fs.String("from", envOrDefault("AGENTLAYER_SAMPLE_FROM", "sender@example.com"), "")
	to := fs.String("to", envOrDefault("AGENTLAYER_SAMPLE_TO", "agent@localhost"), "")
	subject := fs.String("subject", envOrDefault("AGENTLAYER_SAMPLE_SUBJECT", "Hello from agentlayerctl"), "")
	body := fs.String("body", envOrDefault("AGENTLAYER_SAMPLE_BODY", "This is a sample inbound message for AgentLayer."), "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	messageID := fmt.Sprintf("<sample-%d@agentlayer.local>", time.Now().UTC().UnixNano())
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", *from),
		fmt.Sprintf("To: %s", *to),
		fmt.Sprintf("Subject: %s", *subject),
		fmt.Sprintf("Message-ID: %s", messageID),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		*body,
		"",
	}, "\r\n")

	if err := smtp.SendMail(*smtpAddr, nil, *from, []string{*to}, []byte(msg)); err != nil {
		return err
	}

	fmt.Printf("sent sample email to %s via %s\n", *to, *smtpAddr)
	return nil
}

func printJSONRequest(method, url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return printResponse(url, resp)
}

func printResponse(label string, resp *http.Response) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("\n== %s (%d) ==\n%s\n", label, resp.StatusCode, strings.TrimSpace(string(data)))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

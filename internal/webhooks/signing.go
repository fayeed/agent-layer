package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
)

const (
	HeaderSignature          = "X-AgentLayer-Signature"
	HeaderSignatureTimestamp = "X-AgentLayer-Signature-Timestamp"
)

type Clock func() time.Time

type Signer struct {
	now Clock
}

func NewSigner(now Clock) Signer {
	if now == nil {
		now = time.Now
	}

	return Signer{now: now}
}

func (s Signer) Sign(request core.WebhookDispatchRequest, secret string) (core.WebhookDispatchRequest, error) {
	if secret == "" {
		return core.WebhookDispatchRequest{}, errors.New("webhook secret is required")
	}

	headers := make(map[string]string, len(request.Headers)+2)
	for key, value := range request.Headers {
		headers[key] = value
	}

	timestamp := s.now().UTC().Format("2006-01-02T15:04:05Z")
	headers[HeaderSignatureTimestamp] = timestamp
	headers[HeaderSignature] = "sha256=" + signPayload(secret, timestamp, request.Payload)

	request.Headers = headers
	return request, nil
}

func signPayload(secret, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

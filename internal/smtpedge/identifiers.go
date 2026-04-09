package smtpedge

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func NewSessionID(now time.Time) string {
	return "smtp-" + now.UTC().Format("20060102T150405.000000000Z") + "-" + randomHex(6)
}

func NewRawMessageObjectKey(now time.Time, inbox domain.Inbox) string {
	date := now.UTC().Format("2006/01/02")
	inboxID := sanitizePathSegment(inbox.ID)
	if inboxID == "" {
		inboxID = "unknown"
	}
	return "raw/" + date + "/" + inboxID + "/" + randomHex(12) + ".eml"
}

func sanitizePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteRune('-')
		}
	}

	return strings.Trim(builder.String(), "-")
}

func randomHex(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", byteLen*2)
	}
	return hex.EncodeToString(buf)
}

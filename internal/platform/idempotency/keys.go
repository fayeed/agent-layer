package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
)

func InboundMessageKey(inboxID, messageIDHeader string) string {
	return scopedKey("inbound", inboxID, messageIDHeader)
}

func ReplySubmissionKey(threadID, requestID string) string {
	return scopedKey("reply", threadID, requestID)
}

func scopedKey(scope, primary, secondary string) string {
	sum := sha256.Sum256([]byte(scope + ":" + primary + ":" + secondary))
	return scope + ":" + hex.EncodeToString(sum[:])
}

package inbound

import (
	"errors"
	"time"
)

type DurableReceiptRequest struct {
	SMTPTransactionID   string
	OrganizationID      string
	AgentID             string
	InboxID             string
	EnvelopeSender      string
	EnvelopeRecipients  []string
	RawMessageObjectKey string
	ReceivedAt          time.Time
}

func (r DurableReceiptRequest) Validate() error {
	switch {
	case r.SMTPTransactionID == "":
		return errors.New("smtp transaction id is required")
	case r.OrganizationID == "":
		return errors.New("organization id is required")
	case r.AgentID == "":
		return errors.New("agent id is required")
	case r.InboxID == "":
		return errors.New("inbox id is required")
	case r.EnvelopeSender == "":
		return errors.New("envelope sender is required")
	case len(r.EnvelopeRecipients) == 0:
		return errors.New("at least one envelope recipient is required")
	case r.RawMessageObjectKey == "":
		return errors.New("raw message object key is required")
	case r.ReceivedAt.IsZero():
		return errors.New("received at is required")
	default:
		return nil
	}
}

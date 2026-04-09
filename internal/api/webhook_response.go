package api

import "time"

type webhookDeliveryResponse struct {
	ID            string `json:"id"`
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	Status        string `json:"status"`
	AttemptCount  int    `json:"attempt_count"`
	ResponseCode  int    `json:"response_code"`
	NextAttemptAt string `json:"next_attempt_at,omitempty"`
}

func formatResponseTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

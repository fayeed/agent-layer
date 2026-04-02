package outbound

import (
	"testing"
	"time"
)

func TestCallbackParserParsesDeliveryEvent(t *testing.T) {
	parser := NewCallbackParser()

	event, err := parser.Parse([]byte(`{
		"event_type":"delivered",
		"provider_message_id":"ses-123",
		"occurred_at":"2026-04-03T13:00:00Z"
	}`))
	if err != nil {
		t.Fatalf("expected parse to succeed, got error: %v", err)
	}

	if event.ProviderMessageID != "ses-123" {
		t.Fatalf("expected provider message id, got %#v", event)
	}

	if event.Status != DeliveryStateDelivered {
		t.Fatalf("expected delivered status, got %#v", event)
	}

	expected := time.Date(2026, 4, 3, 13, 0, 0, 0, time.UTC)
	if !event.OccurredAt.Equal(expected) {
		t.Fatalf("expected occurred time %v, got %#v", expected, event)
	}
}

func TestCallbackParserParsesBounceAndComplaintEvents(t *testing.T) {
	parser := NewCallbackParser()

	bounce, err := parser.Parse([]byte(`{
		"event_type":"hard_bounce",
		"provider_message_id":"ses-456",
		"occurred_at":"2026-04-03T13:05:00Z"
	}`))
	if err != nil {
		t.Fatalf("expected bounce parse to succeed, got error: %v", err)
	}

	if bounce.Status != DeliveryStateHardBounce {
		t.Fatalf("expected hard bounce status, got %#v", bounce)
	}

	complaint, err := parser.Parse([]byte(`{
		"event_type":"complaint",
		"provider_message_id":"ses-789",
		"occurred_at":"2026-04-03T13:10:00Z"
	}`))
	if err != nil {
		t.Fatalf("expected complaint parse to succeed, got error: %v", err)
	}

	if complaint.Status != DeliveryStateComplaint {
		t.Fatalf("expected complaint status, got %#v", complaint)
	}
}

func TestCallbackParserRejectsUnknownEventType(t *testing.T) {
	parser := NewCallbackParser()

	_, err := parser.Parse([]byte(`{
		"event_type":"mystery",
		"provider_message_id":"ses-000",
		"occurred_at":"2026-04-03T13:15:00Z"
	}`))
	if err == nil {
		t.Fatal("expected unknown event type to fail")
	}
}

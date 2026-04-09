package store

import "testing"

func TestPersistenceModelsExist(t *testing.T) {
	models := []any{
		OrganizationModel{},
		AgentModel{},
		InboxModel{},
		ContactModel{},
		ThreadModel{},
		MessageModel{},
		MessageAttachmentModel{},
		ContactMemoryModel{},
		InboundReceiptModel{},
		WebhookDeliveryModel{},
		SuppressedAddressModel{},
		ProviderConfigModel{},
		AuditLogModel{},
	}

	if len(models) != 13 {
		t.Fatalf("expected 13 persistence models, got %d", len(models))
	}
}

func TestMutableLifecycleFieldsRemainAddressable(t *testing.T) {
	message := MessageModel{}
	message.ProviderMessageID = "provider-123"
	message.DeliveryState = "delivered"

	if message.ProviderMessageID != "provider-123" {
		t.Fatal("expected provider message ID to be assignable")
	}

	if message.DeliveryState != "delivered" {
		t.Fatal("expected delivery state to be assignable")
	}

	webhook := WebhookDeliveryModel{}
	webhook.RequestURL = "https://example.com/webhook"
	webhook.RequestPayload = []byte(`{"ok":true}`)
	webhook.RequestHeaders = []byte(`{"X-Test":"1"}`)

	if webhook.RequestURL != "https://example.com/webhook" {
		t.Fatal("expected webhook request url to be assignable")
	}

	if string(webhook.RequestPayload) != `{"ok":true}` {
		t.Fatal("expected webhook request payload to be assignable")
	}
}

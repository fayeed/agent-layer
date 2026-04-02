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
		WebhookDeliveryModel{},
		SuppressedAddressModel{},
		ProviderConfigModel{},
		AuditLogModel{},
	}

	if len(models) != 12 {
		t.Fatalf("expected 12 persistence models, got %d", len(models))
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
}

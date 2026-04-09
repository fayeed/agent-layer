package store

import (
	"encoding/json"

	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

func OrganizationToModel(organization domain.Organization) OrganizationModel {
	return OrganizationModel{
		ID:        organization.ID,
		Name:      organization.Name,
		CreatedAt: organization.CreatedAt,
		UpdatedAt: organization.UpdatedAt,
	}
}

func OrganizationFromModel(model OrganizationModel) domain.Organization {
	return domain.Organization{
		ID:        model.ID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func AgentToModel(agent domain.Agent) AgentModel {
	return AgentModel{
		ID:             agent.ID,
		OrganizationID: agent.OrganizationID,
		Name:           agent.Name,
		Status:         string(agent.Status),
		WebhookURL:     agent.WebhookURL,
		WebhookSecret:  agent.WebhookSecret,
		CreatedAt:      agent.CreatedAt,
		UpdatedAt:      agent.UpdatedAt,
	}
}

func AgentFromModel(model AgentModel) domain.Agent {
	return domain.Agent{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		Name:           model.Name,
		Status:         domain.AgentStatus(model.Status),
		WebhookURL:     model.WebhookURL,
		WebhookSecret:  model.WebhookSecret,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

func InboxToModel(inbox domain.Inbox) InboxModel {
	return InboxModel{
		ID:               inbox.ID,
		OrganizationID:   inbox.OrganizationID,
		AgentID:          inbox.AgentID,
		EmailAddress:     inbox.EmailAddress,
		Domain:           inbox.Domain,
		DisplayName:      inbox.DisplayName,
		OutboundIdentity: inbox.OutboundIdentity,
		CreatedAt:        inbox.CreatedAt,
		UpdatedAt:        inbox.UpdatedAt,
	}
}

func InboxFromModel(model InboxModel) domain.Inbox {
	return domain.Inbox{
		ID:               model.ID,
		OrganizationID:   model.OrganizationID,
		AgentID:          model.AgentID,
		EmailAddress:     model.EmailAddress,
		Domain:           model.Domain,
		DisplayName:      model.DisplayName,
		OutboundIdentity: model.OutboundIdentity,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}

func ContactToModel(contact domain.Contact) ContactModel {
	return ContactModel{
		ID:             contact.ID,
		OrganizationID: contact.OrganizationID,
		EmailAddress:   contact.EmailAddress,
		DisplayName:    contact.DisplayName,
		LastSeenAt:     contact.LastSeenAt,
		CreatedAt:      contact.CreatedAt,
		UpdatedAt:      contact.UpdatedAt,
	}
}

func ContactFromModel(model ContactModel) domain.Contact {
	return domain.Contact{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		EmailAddress:   model.EmailAddress,
		DisplayName:    model.DisplayName,
		LastSeenAt:     model.LastSeenAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

func ThreadToModel(thread domain.Thread) ThreadModel {
	return ThreadModel{
		ID:                thread.ID,
		OrganizationID:    thread.OrganizationID,
		AgentID:           thread.AgentID,
		InboxID:           thread.InboxID,
		ContactID:         thread.ContactID,
		SubjectNormalized: thread.SubjectNormalized,
		State:             string(thread.State),
		LastInboundID:     thread.LastInboundID,
		LastOutboundID:    thread.LastOutboundID,
		LastActivityAt:    thread.LastActivityAt,
		CreatedAt:         thread.CreatedAt,
		UpdatedAt:         thread.UpdatedAt,
	}
}

func ThreadFromModel(model ThreadModel) domain.Thread {
	return domain.Thread{
		ID:                model.ID,
		OrganizationID:    model.OrganizationID,
		AgentID:           model.AgentID,
		InboxID:           model.InboxID,
		ContactID:         model.ContactID,
		SubjectNormalized: model.SubjectNormalized,
		State:             domain.ThreadState(model.State),
		LastInboundID:     model.LastInboundID,
		LastOutboundID:    model.LastOutboundID,
		LastActivityAt:    model.LastActivityAt,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

func MessageToModel(message domain.Message) MessageModel {
	return MessageModel{
		ID:                message.ID,
		OrganizationID:    message.OrganizationID,
		ThreadID:          message.ThreadID,
		InboxID:           message.InboxID,
		ContactID:         message.ContactID,
		Direction:         string(message.Direction),
		Subject:           message.Subject,
		SubjectNormalized: message.SubjectNormalized,
		MessageIDHeader:   message.MessageIDHeader,
		InReplyTo:         message.InReplyTo,
		References:        append([]string(nil), message.References...),
		TextBody:          message.TextBody,
		HTMLBody:          message.HTMLBody,
		RawMIMEObjectKey:  message.RawMIMEObjectKey,
		ProviderMessageID: message.ProviderMessageID,
		DeliveryState:     message.DeliveryState,
		SentAt:            message.SentAt,
		DeliveredAt:       message.DeliveredAt,
		BouncedAt:         message.BouncedAt,
		CreatedAt:         message.CreatedAt,
	}
}

func MessageFromModel(model MessageModel) domain.Message {
	return domain.Message{
		ID:                model.ID,
		OrganizationID:    model.OrganizationID,
		ThreadID:          model.ThreadID,
		InboxID:           model.InboxID,
		ContactID:         model.ContactID,
		Direction:         domain.MessageDirection(model.Direction),
		Subject:           model.Subject,
		SubjectNormalized: model.SubjectNormalized,
		MessageIDHeader:   model.MessageIDHeader,
		InReplyTo:         model.InReplyTo,
		References:        append([]string(nil), model.References...),
		TextBody:          model.TextBody,
		HTMLBody:          model.HTMLBody,
		RawMIMEObjectKey:  model.RawMIMEObjectKey,
		ProviderMessageID: model.ProviderMessageID,
		DeliveryState:     model.DeliveryState,
		SentAt:            model.SentAt,
		DeliveredAt:       model.DeliveredAt,
		BouncedAt:         model.BouncedAt,
		CreatedAt:         model.CreatedAt,
	}
}

func ContactMemoryToModel(entry domain.ContactMemoryEntry, organizationID string) ContactMemoryModel {
	return ContactMemoryModel{
		ID:             entry.ID,
		OrganizationID: organizationID,
		ContactID:      entry.ContactID,
		ThreadID:       entry.ThreadID,
		Note:           entry.Note,
		Tags:           append([]string(nil), entry.Tags...),
		CreatedAt:      entry.CreatedAt,
	}
}

func ContactMemoryFromModel(model ContactMemoryModel) domain.ContactMemoryEntry {
	return domain.ContactMemoryEntry{
		ID:        model.ID,
		ContactID: model.ContactID,
		ThreadID:  model.ThreadID,
		Note:      model.Note,
		Tags:      append([]string(nil), model.Tags...),
		CreatedAt: model.CreatedAt,
	}
}

func InboundReceiptToModel(receipt inbound.DurableReceiptRequest) InboundReceiptModel {
	return InboundReceiptModel{
		RawMessageObjectKey: receipt.RawMessageObjectKey,
		SMTPTransactionID:   receipt.SMTPTransactionID,
		OrganizationID:      receipt.OrganizationID,
		AgentID:             receipt.AgentID,
		InboxID:             receipt.InboxID,
		EnvelopeSender:      receipt.EnvelopeSender,
		EnvelopeRecipients:  append([]string(nil), receipt.EnvelopeRecipients...),
		ReceivedAt:          receipt.ReceivedAt,
		CreatedAt:           receipt.ReceivedAt,
	}
}

func InboundReceiptFromModel(model InboundReceiptModel) inbound.DurableReceiptRequest {
	return inbound.DurableReceiptRequest{
		SMTPTransactionID:   model.SMTPTransactionID,
		OrganizationID:      model.OrganizationID,
		AgentID:             model.AgentID,
		InboxID:             model.InboxID,
		EnvelopeSender:      model.EnvelopeSender,
		EnvelopeRecipients:  append([]string(nil), model.EnvelopeRecipients...),
		RawMessageObjectKey: model.RawMessageObjectKey,
		ReceivedAt:          model.ReceivedAt,
	}
}

func WebhookDeliveryToModel(delivery domain.WebhookDelivery) (WebhookDeliveryModel, error) {
	headers, err := json.Marshal(delivery.RequestHeaders)
	if err != nil {
		return WebhookDeliveryModel{}, err
	}

	return WebhookDeliveryModel{
		ID:             delivery.ID,
		OrganizationID: delivery.OrganizationID,
		AgentID:        delivery.AgentID,
		EventType:      delivery.EventType,
		EventID:        delivery.EventID,
		RequestURL:     delivery.RequestURL,
		RequestPayload: append([]byte(nil), delivery.RequestPayload...),
		RequestHeaders: headers,
		Status:         delivery.Status,
		AttemptCount:   delivery.AttemptCount,
		ResponseCode:   delivery.ResponseCode,
		ResponseBody:   append([]byte(nil), delivery.ResponseBody...),
		LastAttemptAt:  delivery.LastAttemptAt,
		NextAttemptAt:  delivery.NextAttemptAt,
		CreatedAt:      delivery.CreatedAt,
		UpdatedAt:      delivery.UpdatedAt,
	}, nil
}

func WebhookDeliveryFromModel(model WebhookDeliveryModel) (domain.WebhookDelivery, error) {
	var headers map[string]string
	if len(model.RequestHeaders) > 0 {
		if err := json.Unmarshal(model.RequestHeaders, &headers); err != nil {
			return domain.WebhookDelivery{}, err
		}
	}

	return domain.WebhookDelivery{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		AgentID:        model.AgentID,
		EventType:      model.EventType,
		EventID:        model.EventID,
		RequestURL:     model.RequestURL,
		RequestPayload: append([]byte(nil), model.RequestPayload...),
		RequestHeaders: headers,
		Status:         model.Status,
		AttemptCount:   model.AttemptCount,
		ResponseCode:   model.ResponseCode,
		ResponseBody:   append([]byte(nil), model.ResponseBody...),
		LastAttemptAt:  model.LastAttemptAt,
		NextAttemptAt:  model.NextAttemptAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

func SuppressedAddressToModel(record domain.SuppressedAddress) SuppressedAddressModel {
	return SuppressedAddressModel{
		ID:             record.ID,
		OrganizationID: record.OrganizationID,
		EmailAddress:   record.EmailAddress,
		Reason:         record.Reason,
		Source:         record.Source,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}

func SuppressedAddressFromModel(model SuppressedAddressModel) domain.SuppressedAddress {
	return domain.SuppressedAddress{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		EmailAddress:   model.EmailAddress,
		Reason:         model.Reason,
		Source:         model.Source,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

func ProviderConfigToModel(config domain.ProviderConfig, configJSON []byte) ProviderConfigModel {
	return ProviderConfigModel{
		ID:             config.ID,
		OrganizationID: config.OrganizationID,
		ProviderType:   config.ProviderType,
		IsDefault:      config.IsDefault,
		ConfigJSON:     append([]byte(nil), configJSON...),
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
	}
}

func ProviderConfigFromModel(model ProviderConfigModel) domain.ProviderConfig {
	return domain.ProviderConfig{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		ProviderType:   model.ProviderType,
		IsDefault:      model.IsDefault,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}
